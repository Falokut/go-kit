package tg_botx

import (
	"context"
	"net/http"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/tg_bot"
	"github.com/pkg/errors"
)

const initializationSpinlockTime = 500 * time.Millisecond

type Muxer interface {
	Handle(ctx context.Context, msg tg_bot.Update) (tg_bot.Chattable, error)
}

const (
	updatesRetryDelay = time.Second * 3
)

type Bot struct {
	logger     log.Logger
	mux        *atomic.Value
	cli        *atomic.Pointer[tg_bot.BotApi]
	prevCfg    *atomic.Value
	logCtx     context.Context // nolint:containedctx
	shutdownCh chan any
}

func New(logger log.Logger) *Bot {
	prevCfg := &atomic.Value{}
	prevCfg.Store(Config{})
	return &Bot{
		cli:        &atomic.Pointer[tg_bot.BotApi]{},
		mux:        &atomic.Value{},
		prevCfg:    prevCfg,
		shutdownCh: make(chan any, 1),
		logCtx:     context.Background(),
		logger:     logger,
	}
}

func (bot *Bot) UpgradeMux(logCtx context.Context, mux Muxer) {
	bot.mux.Store(mux)
	bot.logger.Debug(logCtx, "bot client: mux initialization done")
}

func (bot *Bot) UpgradeConfig(logCtx context.Context, cfg Config) error {
	if reflect.DeepEqual(bot.prevCfg.Load(), cfg) {
		bot.logger.Debug(logCtx, "bot client: configs are equal. skipping config initialization")
		return nil
	}

	bot.logCtx = logCtx
	endpoint := tg_bot.ApiEndpoint
	if cfg.ApiEndpoint != "" {
		endpoint = cfg.ApiEndpoint
	}
	botApi, err := tg_bot.NewBotApiWithApiEndpoint(logCtx, cfg.Token, endpoint, bot.logger)
	if err != nil {
		return errors.WithMessage(err, "new bot api")
	}
	bot.cli.Store(botApi)
	bot.prevCfg.Store(cfg)
	bot.logger.Debug(logCtx, "bot client: config initialization done")
	return nil
}

func (bot *Bot) RegisterCommands(commands ...tg_bot.BotCommand) error {
	err := bot.Api().Send(tg_bot.NewSetMyCommands(commands...))
	if err != nil {
		return errors.WithMessage(err, "send bot commands")
	}
	return nil
}

func (bot *Bot) Api() *tg_bot.BotApi {
	return bot.cli.Load()
}

func (bot *Bot) Send(msg tg_bot.Chattable) error {
	err := bot.Api().Send(msg)
	if err != nil {
		return errors.WithMessage(err, "send bot commands")
	}
	return nil
}

func (bot *Bot) ClearAllCommands() error {
	err := bot.Api().Send(tg_bot.NewDeleteMyCommands())
	if err != nil {
		return errors.WithMessage(err, "send delete default scope commands")
	}
	return nil
}

// Serve fetches updates and sends them to muxer.
// nolint:gocognit,cyclop,funlen
func (bot *Bot) Serve(ctx context.Context) error {
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-bot.shutdownCh:
			return nil
		default:
			mux, cfg, cli := bot.waitForInitialization()

			retryDelay := updatesRetryDelay
			if cfg.RetryDelaySec > 0 {
				retryDelay = time.Second * time.Duration(cfg.RetryDelaySec)
			}

			currentOffset, shouldStopServing, err := bot.processUpdates(ctx, offset, mux, cfg, cli)
			switch {
			case err != nil:
				bot.logger.Error(ctx, "tg bot serve", log.Error(err))
				continue
			case errors.As(err, &RetryWithDelayError{}):
				bot.logger.Error(ctx, "tg bot serve",
					log.Error(err),
					log.Time("retryAfter", time.Now().UTC().Add(retryDelay)),
				)
				time.Sleep(retryDelay)
				continue
			case shouldStopServing:
				bot.logger.Warn(ctx, "stop serving", log.String("reason", "another bot instance is running"))
				return nil
			}
			offset = currentOffset
		}
	}
}

func (bot *Bot) Shutdown() {
	bot.logger.Debug(bot.logCtx, "bot stopping the update receiver routine")
	close(bot.shutdownCh)
}

// processUpdates fetches and processes updates, retrying if needed.
func (bot *Bot) processUpdates(
	ctx context.Context,
	offset int,
	mux Muxer,
	cfg Config,
	cli *tg_bot.BotApi,
) (int, bool, error) {
	updatesCfg := tg_bot.UpdatesConfig{
		Offset:         offset,
		Limit:          cfg.Limit,
		Timeout:        cfg.Timeout,
		AllowedUpdates: cfg.AllowedUpdates,
	}

	apiResp, err := cli.Request(updatesCfg)
	if err != nil {
		shouldStopServing, err := bot.handleUpdateRequestApiError(err)
		return 0, shouldStopServing, err
	}

	var updates []tg_bot.Update
	err = json.Unmarshal(apiResp.Result, &updates)
	if err != nil {
		return 0, false, errors.WithMessage(err, "unmarshal updates")
	}

	for _, update := range updates {
		if update.UpdateId < updatesCfg.Offset {
			continue
		}

		resp, err := bot.handleUpdateWithRetry(ctx, mux, update, cfg)
		if err != nil {
			return 0, false, errors.WithMessage(err, "handle update with retry")
		}

		if resp == nil {
			updatesCfg.Offset = update.UpdateId + 1
			continue
		}

		err = cli.Send(resp)
		if err != nil {
			return 0, false, errors.WithMessage(err, "bot send response")
		}

		updatesCfg.Offset = update.UpdateId + 1
	}

	return updatesCfg.Offset, false, nil
}

func (bot *Bot) handleUpdateRequestApiError(err error) (bool, error) {
	var apiErr *tg_bot.Error
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusConflict {
		return true, nil
	}
	return false, NewRetryWithDelayError(err.Error())
}

func (bot *Bot) waitForInitialization() (Muxer, Config, *tg_bot.BotApi) {
	for {
		mux, muxOk := bot.mux.Load().(Muxer)
		cfg, _ := bot.prevCfg.Load().(Config)
		cli := bot.cli.Load()
		if muxOk && cfg.Token != "" && cli != nil {
			return mux, cfg, cli
		}

		time.Sleep(initializationSpinlockTime)
	}
}

// handleUpdateWithRetry processes the update and retries if there's an error.
func (bot *Bot) handleUpdateWithRetry(
	ctx context.Context,
	mux Muxer,
	update tg_bot.Update,
	cfg Config,
) (tg_bot.Chattable, error) {
	var resp tg_bot.Chattable
	var err error
	retry := 0

	// Retry logic for processing the update
	for {
		resp, err = mux.Handle(ctx, update)
		if err == nil {
			return resp, nil
		}

		// If max retry count is reached, stop retrying
		if cfg.MaxRetryCount != -1 && retry >= cfg.MaxRetryCount {
			bot.logger.Warn(ctx,
				"mux handle error reached max retry count",
				log.Error(err),
				log.Time("retryAfter", time.Now().UTC().Add(updatesRetryDelay)),
				log.Int("retry", retry),
				log.Int("maxRetryCount", cfg.MaxRetryCount),
			)
			break
		}

		// Retry and sleep before trying again
		retry++
		time.Sleep(updatesRetryDelay)
	}

	return resp, NewRetryWithDelayError(err.Error())
}
