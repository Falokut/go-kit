// Package telegram_bot has functions and types used for interacting with
// the Telegram Bot API.
package tg_bot

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/Falokut/go-kit/log"

	"github.com/Falokut/go-kit/json"
)

// HttpClient is the type needed for the bot to perform HTTP requests.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// BotApi allows you to interact with the Telegram Bot API.
type BotApi struct {
	logger      log.Logger
	Self        User
	client      HttpClient
	logCtx      context.Context // nolint:containedctx
	token       string
	apiEndpoint string
}

// NewBotApi creates a new BotAPI instance.
//
// It requires a token, provIded by @BotFather on Telegram.
func NewBotApi(logCtx context.Context, token string, logger log.Logger) (*BotApi, error) {
	return NewBotApiWithClient(logCtx, token, ApiEndpoint, &http.Client{}, logger)
}

// NewBotApiWithApiEndpoint creates a new BotAPI instance
// and allows you to pass API endpoint.
//
// It requires a token, provIded by @BotFather on Telegram and API endpoint.
func NewBotApiWithApiEndpoint(logCtx context.Context, token string, apiEndpoint string, logger log.Logger) (*BotApi, error) {
	return NewBotApiWithClient(logCtx, token, apiEndpoint, &http.Client{}, logger)
}

// NewBotApiWithClient creates a new BotAPI instance
// and allows you to pass a http.Client.
//
// It requires a token, provIded by @BotFather on Telegram and API endpoint.
func NewBotApiWithClient(
	logCtx context.Context,
	token string,
	apiEndpoint string,
	client HttpClient,
	logger log.Logger,
) (*BotApi, error) {
	bot := &BotApi{
		client:      client,
		logger:      logger,
		logCtx:      logCtx,
		apiEndpoint: apiEndpoint,
		token:       token,
	}
	return bot, nil
}

// SetApiEndpoint changes the Telegram Bot API endpoint used by the instance.
func (bot *BotApi) SetApiEndpoint(apiEndpoint string) {
	bot.apiEndpoint = apiEndpoint
}

// MakeRequest makes a request to a specific endpoint with our token.
func (bot *BotApi) MakeRequest(endpoint string, params Params) (*ApiResponse, error) {
	bot.logger.Debug(bot.logCtx, "bot request", log.Any("endpoint", endpoint))

	method := fmt.Sprintf(bot.apiEndpoint, bot.token, endpoint)
	values := buildParams(params)
	req, err := http.NewRequest(http.MethodPost, method, strings.NewReader(values.Encode()))
	if err != nil {
		return &ApiResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := bot.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	apiResp, err := bot.decodeApiResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	bot.logger.Debug(bot.logCtx, "bot response",
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

// decodeApiResponse decode http.Response.Body stream to APIResponse struct
// for efficient memory usage
func (bot *BotApi) decodeApiResponse(body io.Reader) (*ApiResponse, error) {
	var resp ApiResponse
	err := json.NewDecoder(body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UploadFiles makes a request to the API with files.
// nolint:gocognit,cyclop,funlen
func (bot *BotApi) UploadFiles(endpoint string, params Params, files []RequestFile) (*ApiResponse, error) {
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
			if !file.Data.NeedsUpload() {
				value := file.Data.SendData()

				err := m.WriteField(file.Name, value)
				if err != nil {
					w.CloseWithError(err)
					return
				}
				continue
			}
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

			_, err = io.Copy(part, reader)
			if err != nil {
				w.CloseWithError(err)
				return
			}

			closer, ok := reader.(io.ReadCloser)
			if ok {
				err = closer.Close()
				if err != nil {
					w.CloseWithError(err)
					return
				}
			}
		}
	}()

	bot.logger.Debug(bot.logCtx, "bot upload files",
		log.Any("endpoint", endpoint),
		log.Any("params", params),
		log.Any("numFiles", len(files)),
	)

	method := fmt.Sprintf(bot.apiEndpoint, bot.token, endpoint)
	req, err := http.NewRequest(http.MethodPost, method, r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", m.FormDataContentType())

	resp, err := bot.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	apiResp, err := bot.decodeApiResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	bot.logger.Debug(bot.logCtx, "bot upload files response",
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
func (bot *BotApi) GetFileDirectURL(fileId string) (string, error) {
	file, err := bot.GetFile(FileConfig{fileId})

	if err != nil {
		return "", err
	}

	return file.Link(bot.token), nil
}

// GetMe fetches the currently authenticated bot.
//
// This method is called upon creation to valIdate the token,
// and so you may get this data from BotAPI.Self without the need for
// another request.
func (bot *BotApi) GetMe() (User, error) {
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
func (bot *BotApi) IsMessageToMe(message Message) bool {
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
func (bot *BotApi) Request(c Chattable) (*ApiResponse, error) {
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
func (bot *BotApi) Send(c Chattable) error {
	_, err := bot.Request(c)
	if err != nil {
		return err
	}

	return nil
}

// SendMediaGroup sends a media group and returns the resulting messages.
func (bot *BotApi) SendMediaGroup(config MediaGroupConfig) ([]Message, error) {
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
func (bot *BotApi) GetUserProfilePhotos(config UserProfilePhotosConfig) (UserProfilePhotos, error) {
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
func (bot *BotApi) GetFile(config FileConfig) (File, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return File{}, err
	}

	var file File
	err = json.Unmarshal(resp.Result, &file)
	return file, err
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
func (bot *BotApi) GetChat(config ChatInfoConfig) (Chat, error) {
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
func (bot *BotApi) GetChatAdministrators(config ChatAdministratorsConfig) ([]ChatMember, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []ChatMember{}, err
	}

	var members []ChatMember
	err = json.Unmarshal(resp.Result, &members)
	return members, err
}

// GetChatMembersCount gets the number of users in a chat.
func (bot *BotApi) GetChatMembersCount(config ChatMemberCountConfig) (int, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return -1, err
	}

	var count int
	err = json.Unmarshal(resp.Result, &count)
	return count, err
}

// GetChatMember gets a specific chat member.
func (bot *BotApi) GetChatMember(config GetChatMemberConfig) (ChatMember, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return ChatMember{}, err
	}

	var member ChatMember
	err = json.Unmarshal(resp.Result, &member)
	return member, err
}

// GetGameHighScores allows you to get the high scores for a game.
func (bot *BotApi) GetGameHighScores(config GetGameHighScoresConfig) ([]GameHighScore, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return []GameHighScore{}, err
	}

	var highScores []GameHighScore
	err = json.Unmarshal(resp.Result, &highScores)
	return highScores, err
}

// GetInviteLink get InviteLink for a chat
func (bot *BotApi) GetInviteLink(config ChatInviteLinkConfig) (string, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return "", err
	}

	var inviteLink string
	err = json.Unmarshal(resp.Result, &inviteLink)
	return inviteLink, err
}

// GetStickerSet returns a StickerSet.
func (bot *BotApi) GetStickerSet(config GetStickerSetConfig) (StickerSet, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return StickerSet{}, err
	}

	var stickers StickerSet
	err = json.Unmarshal(resp.Result, &stickers)
	return stickers, err
}

// StopPoll stops a poll and returns the result.
func (bot *BotApi) StopPoll(config StopPollConfig) (Poll, error) {
	resp, err := bot.Request(config)
	if err != nil {
		return Poll{}, err
	}

	var poll Poll
	err = json.Unmarshal(resp.Result, &poll)
	return poll, err
}

// GetMyCommands gets the currently registered commands.
func (bot *BotApi) GetMyCommands() ([]BotCommand, error) {
	return bot.GetMyCommandsWithConfig(GetMyCommandsConfig{})
}

// GetMyCommandsWithConfig gets the currently registered commands with a config.
func (bot *BotApi) GetMyCommandsWithConfig(config GetMyCommandsConfig) ([]BotCommand, error) {
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
func (bot *BotApi) CopyMessage(config CopyMessageConfig) (MessageId, error) {
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
	switch parseMode {
	case ModeHTML:
		replacer = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	case ModeMarkdown:
		replacer = strings.NewReplacer("_", "\\_", "*", "\\*", "`", "\\`", "[", "\\[")
	case ModeMarkdownV2:
		replacer = strings.NewReplacer(
			"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
			"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
			"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
			"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
		)
	default:
		return ""
	}
	return replacer.Replace(text)
}
