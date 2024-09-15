// Package telegram_bot has functions and types used for interacting with
// the Telegram Bot API.
package telegram_bot

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/log"

	"github.com/Falokut/go-kit/json"
)

const (
	updatesRetryTimeout = time.Second * 3
)

type Muxer interface {
	Handle(ctx context.Context, msg Update) (Chattable, error)
}

// HTTPClient is the type needed for the bot to perform HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// BotAPI allows you to interact with the Telegram Bot API.
type BotAPI struct {
	Token       string `json:"token"`
	Buffer      int    `json:"buffer"`
	logger      log.Logger
	Self        User       `json:"-"`
	Client      HTTPClient `json:"-"`
	ctx         context.Context
	shutdownCh  chan any
	mux         atomic.Value
	apiEndpoint string
}

// NewBotAPI creates a new BotAPI instance.
//
// It requires a token, provIded by @BotFather on Telegram.
func NewBotAPI(ctx context.Context, token string, logger log.Logger) (*BotAPI, error) {
	return NewBotAPIWithClient(ctx, token, APIEndpoint, &http.Client{}, logger)
}

// NewBotAPIWithAPIEndpoint creates a new BotAPI instance
// and allows you to pass API endpoint.
//
// It requires a token, provIded by @BotFather on Telegram and API endpoint.
func NewBotAPIWithAPIEndpoint(ctx context.Context, token, apiEndpoint string, logger log.Logger) (*BotAPI, error) {
	return NewBotAPIWithClient(ctx, token, apiEndpoint, &http.Client{}, logger)
}

// NewBotAPIWithClient creates a new BotAPI instance
// and allows you to pass a http.Client.
//
// It requires a token, provIded by @BotFather on Telegram and API endpoint.
func NewBotAPIWithClient(ctx context.Context,
	token, apiEndpoint string, client HTTPClient, logger log.Logger) (*BotAPI, error) {
	bot := &BotAPI{
		Token:       token,
		Client:      client,
		logger:      logger,
		Buffer:      100,
		ctx:         ctx,
		shutdownCh:  make(chan any),
		apiEndpoint: apiEndpoint,
		mux:         atomic.Value{},
	}

	self, err := bot.GetMe()
	if err != nil {
		return nil, err
	}

	bot.Self = self
	bot.logger.Debug(ctx, "bot authorized on account", log.Any("accountName", bot.Self.UserName))
	return bot, nil
}

// SetAPIEndpoint changes the Telegram Bot API endpoint used by the instance.
func (bot *BotAPI) SetAPIEndpoint(apiEndpoint string) {
	bot.apiEndpoint = apiEndpoint
}
func (bot *BotAPI) UpsertCommands(commands []SetMyCommandsConfig) error {
	for i, commandsConfig := range commands {
		err := bot.Send(commandsConfig)
		if err != nil {
			return errors.WithMessagef(err, "send [%d] commands config", i)
		}
	}
	return nil
}

func (bot *BotAPI) Upgrade(mux Muxer) {
	bot.mux.Store(mux)
}

func buildParams(in Params) url.Values {
	if in == nil {
		return url.Values{}
	}

	out := url.Values{}
	for key, value := range in {
		out.Set(key, value)
	}

	return out
}

// MakeRequest makes a request to a specific endpoint with our token.
func (bot *BotAPI) MakeRequest(endpoint string, params Params) (*APIResponse, error) {
	bot.logger.Debug(bot.ctx, "bot request",
		log.Any("endpoint", endpoint),
		log.Any("params", params),
	)

	method := fmt.Sprintf(bot.apiEndpoint, bot.Token, endpoint)
	values := buildParams(params)
	req, err := http.NewRequest("POST", method, strings.NewReader(values.Encode()))
	if err != nil {
		return &APIResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := bot.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	apiResp, err := bot.decodeAPIResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	bot.logger.Debug(bot.ctx, "bot response",
		log.Any("endpoint", endpoint),
		log.Any("response", apiResp),
	)

	if apiResp.Ok {
		return apiResp, nil
	}

	var parameters ResponseParameters
	if apiResp.Parameters != nil {
		parameters = *apiResp.Parameters
	}

	return nil, &Error{
		Code:               apiResp.ErrorCode,
		Message:            apiResp.Description,
		ResponseParameters: parameters,
	}
}

// decodeAPIResponse decode http.Response.Body stream to APIResponse struct
// for efficient memory usage
func (bot *BotAPI) decodeAPIResponse(body io.Reader) (*APIResponse, error) {
	var resp APIResponse
	err := json.NewDecoder(body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UploadFiles makes a request to the API with files.
func (bot *BotAPI) UploadFiles(endpoint string, params Params, files []RequestFile) (*APIResponse, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer m.Close()

		for field, value := range params {
			if err := m.WriteField(field, value); err != nil {
				w.CloseWithError(err)
				return
			}
		}

		for _, file := range files {
			if file.Data.NeedsUpload() {
				name, reader, err := file.Data.UploadData()
				if err != nil {
					w.CloseWithError(err)
					return
				}

				part, err := m.CreateFormFile(file.Name, name)
				if err != nil {
					w.CloseWithError(err)
					return
				}

				if _, err := io.Copy(part, reader); err != nil {
					w.CloseWithError(err)
					return
				}

				if closer, ok := reader.(io.ReadCloser); ok {
					if err = closer.Close(); err != nil {
						w.CloseWithError(err)
						return
					}
				}
			} else {
				value := file.Data.SendData()

				if err := m.WriteField(file.Name, value); err != nil {
					w.CloseWithError(err)
					return
				}
			}
		}
	}()

	bot.logger.Debug(bot.ctx, "bot upload files",
		log.Any("endpoint", endpoint),
		log.Any("params", params),
		log.Any("numFiles", len(files)),
	)

	method := fmt.Sprintf(bot.apiEndpoint, bot.Token, endpoint)
	req, err := http.NewRequest("POST", method, r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", m.FormDataContentType())

	resp, err := bot.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	apiResp, err := bot.decodeAPIResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	bot.logger.Debug(bot.ctx, "bot upload files response",
		log.Any("endpoint", endpoint),
		log.Any("response", apiResp),
	)

	if apiResp.Ok {
		return apiResp, nil
	}

	var parameters ResponseParameters
	if apiResp.Parameters != nil {
		parameters = *apiResp.Parameters
	}

	return nil, &Error{
		Message:            apiResp.Description,
		ResponseParameters: parameters,
	}
}

// GetFileDirectURL returns direct URL to file
//
// It requires the FileId.
func (bot *BotAPI) GetFileDirectURL(fileId string) (string, error) {
	file, err := bot.GetFile(FileConfig{fileId})

	if err != nil {
		return "", err
	}

	return file.Link(bot.Token), nil
}

// GetMe fetches the currently authenticated bot.
//
// This method is called upon creation to valIdate the token,
// and so you may get this data from BotAPI.Self without the need for
// another request.
func (bot *BotAPI) GetMe() (User, error) {
	resp, err := bot.MakeRequest("getMe", nil)
	if err != nil {
		return User{}, err
	}

	var user User
	err = json.Unmarshal(resp.Result, &user)
	return user, err
}

// IsMessageToMe returns true if message directed to this bot.
//
// It requires the Message.
func (bot *BotAPI) IsMessageToMe(message Message) bool {
	return strings.Contains(message.Text, "@"+bot.Self.UserName)
}

func hasFilesNeedingUpload(files []RequestFile) bool {
	for _, file := range files {
		if file.Data.NeedsUpload() {
			return true
		}
	}

	return false
}

// Request sends a Chattable to Telegram, and returns the APIResponse.
func (bot *BotAPI) Request(c Chattable) (*APIResponse, error) {
	params, err := c.params()
	if err != nil {
		return nil, err
	}

	if t, ok := c.(Fileable); ok {
		files := t.files()

		// If we have files that need to be uploaded, we should delegate the
		// request to UploadFile.
		if hasFilesNeedingUpload(files) {
			return bot.UploadFiles(t.method(), params, files)
		}

		// However, if there are no files to be uploaded, there's likely things
		// that need to be turned into params instead.
		for _, file := range files {
			params[file.Name] = file.Data.SendData()
		}
	}

	return bot.MakeRequest(c.method(), params)
}

// Send will send a Chattable item to Telegram
func (bot *BotAPI) Send(c Chattable) error {
	_, err := bot.Request(c)
	if err != nil {
		return err
	}

	return nil
}

// SendMediaGroup sends a media group and returns the resulting messages.
func (bot *BotAPI) SendMediaGroup(config MediaGroupConfig) ([]Message, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var messages []Message
	err = json.Unmarshal(resp.Result, &messages)
	return messages, err
}

// GetUserProfilePhotos gets a user's profile photos.
//
// It requires UserId.
// Offset and Limit are optional.
func (bot *BotAPI) GetUserProfilePhotos(config UserProfilePhotosConfig) (UserProfilePhotos, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return UserProfilePhotos{}, err
	}

	var profilePhotos UserProfilePhotos
	err = json.Unmarshal(resp.Result, &profilePhotos)
	return profilePhotos, err
}

// GetFile returns a File which can download a file from Telegram.
//
// Requires FileId.
func (bot *BotAPI) GetFile(config FileConfig) (File, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return File{}, err
	}

	var file File
	err = json.Unmarshal(resp.Result, &file)
	return file, err
}

// Serve fetches updates and send it to muxer.
//
// Offset, Limit, Timeout, and AllowedUpdates are optional.
// To avoId stale items, set Offset to one higher than the previous item.
// Set Timeout to a large number to reduce requests, so you can get updates
// instantly instead of having to wait between requests.
func (bot *BotAPI) Serve(ctx context.Context, config UpdatesConfig) error {
	mux, ok := bot.mux.Load().(Muxer)
	if !ok {
		return errors.New("bot serve error: muxer not initialized")
	}
	ctx = log.ToContext(ctx, log.Any("accountName", bot.Self.UserName))
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-bot.shutdownCh:
			return nil
		default:
		}
		resp, err := bot.Request(config)
		if err != nil {
			bot.logger.Error(bot.ctx, "failed to get updates",
				log.Any("error", err),
				log.Any("retryAfter", time.Now().Add(updatesRetryTimeout)),
			)
			time.Sleep(updatesRetryTimeout)
			continue
		}

		var updates []Update
		err = json.Unmarshal(resp.Result, &updates)
		if err != nil {
			bot.logger.Error(bot.ctx, "failed to unmarshal updates", log.Any("error", err))
			time.Sleep(updatesRetryTimeout)
			continue
		}
		for _, update := range updates {
			if update.UpdateId < config.Offset {
				continue
			}
			resp, err := mux.Handle(ctx, update)
			if err != nil {
				time.Sleep(updatesRetryTimeout)
				break
			}
			if resp != nil {
				err = bot.Send(resp)
				if err != nil {
					bot.logger.Error(ctx, "bot send failed", log.Any("error", err))
					time.Sleep(updatesRetryTimeout)
					break
				}
			}
			config.Offset = update.UpdateId + 1
		}
	}
}

// StopReceivingUpdates stops the go routine which receives updates
func (bot *BotAPI) StopReceivingUpdates() {
	bot.logger.Debug(bot.ctx, "bot stopping the update receiver routine")
	close(bot.shutdownCh)
}

// WriteToHTTPResponse writes the request to the HTTP ResponseWriter.
//
// It doesn't support uploading files.
func WriteToHTTPResponse(w http.ResponseWriter, c Chattable) error {
	params, err := c.params()
	if err != nil {
		return err
	}

	if t, ok := c.(Fileable); ok {
		if hasFilesNeedingUpload(t.files()) {
			return errors.New("unable to use http response to upload files")
		}
	}

	values := buildParams(params)
	values.Set("method", c.method())

	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	_, err = w.Write([]byte(values.Encode()))
	return err
}

// GetChat gets information about a chat.
func (bot *BotAPI) GetChat(config ChatInfoConfig) (Chat, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Chat{}, err
	}

	var chat Chat
	err = json.Unmarshal(resp.Result, &chat)
	return chat, err
}

// GetChatAdministrators gets a list of administrators in the chat.
//
// If none have been appointed, only the creator will be returned.
// Bots are not shown, even if they are an administrator.
func (bot *BotAPI) GetChatAdministrators(config ChatAdministratorsConfig) ([]ChatMember, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []ChatMember{}, err
	}

	var members []ChatMember
	err = json.Unmarshal(resp.Result, &members)
	return members, err
}

// GetChatMembersCount gets the number of users in a chat.
func (bot *BotAPI) GetChatMembersCount(config ChatMemberCountConfig) (int, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return -1, err
	}

	var count int
	err = json.Unmarshal(resp.Result, &count)
	return count, err
}

// GetChatMember gets a specific chat member.
func (bot *BotAPI) GetChatMember(config GetChatMemberConfig) (ChatMember, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return ChatMember{}, err
	}

	var member ChatMember
	err = json.Unmarshal(resp.Result, &member)
	return member, err
}

// GetGameHighScores allows you to get the high scores for a game.
func (bot *BotAPI) GetGameHighScores(config GetGameHighScoresConfig) ([]GameHighScore, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []GameHighScore{}, err
	}

	var highScores []GameHighScore
	err = json.Unmarshal(resp.Result, &highScores)
	return highScores, err
}

// GetInviteLink get InviteLink for a chat
func (bot *BotAPI) GetInviteLink(config ChatInviteLinkConfig) (string, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return "", err
	}

	var inviteLink string
	err = json.Unmarshal(resp.Result, &inviteLink)
	return inviteLink, err
}

// GetStickerSet returns a StickerSet.
func (bot *BotAPI) GetStickerSet(config GetStickerSetConfig) (StickerSet, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return StickerSet{}, err
	}

	var stickers StickerSet
	err = json.Unmarshal(resp.Result, &stickers)
	return stickers, err
}

// StopPoll stops a poll and returns the result.
func (bot *BotAPI) StopPoll(config StopPollConfig) (Poll, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Poll{}, err
	}

	var poll Poll
	err = json.Unmarshal(resp.Result, &poll)
	return poll, err
}

// GetMyCommands gets the currently registered commands.
func (bot *BotAPI) GetMyCommands() ([]BotCommand, error) {
	return bot.GetMyCommandsWithConfig(GetMyCommandsConfig{})
}

// GetMyCommandsWithConfig gets the currently registered commands with a config.
func (bot *BotAPI) GetMyCommandsWithConfig(config GetMyCommandsConfig) ([]BotCommand, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return nil, err
	}

	var commands []BotCommand
	err = json.Unmarshal(resp.Result, &commands)
	return commands, err
}

// CopyMessage copy messages of any kind. The method is analogous to the method
// forwardMessage, but the copied message doesn't have a link to the original
// message. Returns the MessageId of the sent message on success.
func (bot *BotAPI) CopyMessage(config CopyMessageConfig) (MessageId, error) {
	params, err := config.params()
	if err != nil {
		return MessageId{}, err
	}

	resp, err := bot.MakeRequest(config.method(), params)
	if err != nil {
		return MessageId{}, err
	}

	var messageId MessageId
	err = json.Unmarshal(resp.Result, &messageId)
	return messageId, err
}

// EscapeText takes an input text and escape Telegram markup symbols.
// In this way we can send a text without being afraId of having to escape the characters manually.
// Note that you don't have to include the formatting style in the input text, or it will be escaped too.
// If there is an error, an empty string will be returned.
//
// parseMode is the text formatting mode (ModeMarkdown, ModeMarkdownV2 or ModeHTML)
// text is the input string that will be escaped
func EscapeText(parseMode string, text string) string {
	var replacer *strings.Replacer

	if parseMode == ModeHTML {
		replacer = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	} else if parseMode == ModeMarkdown {
		replacer = strings.NewReplacer("_", "\\_", "*", "\\*", "`", "\\`", "[", "\\[")
	} else if parseMode == ModeMarkdownV2 {
		replacer = strings.NewReplacer(
			"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
			"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
			"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
			"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
		)
	} else {
		return ""
	}

	return replacer.Replace(text)
}
