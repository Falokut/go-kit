package telegram_bot

import "time"

const (
	MessageUpdateType                                = "message"
	SuccessfulPaymentMessageUpdateType               = "successful_payment_message"
	EditedMessageUpdateType                          = "edited_message"
	SuccessfulPaymentEditedMessageUpdateType         = "successful_payment_edited_message"
	ChannelPostUpdateType                            = "channel_post"
	EditedChannelPostUpdateType                      = "edited_channel_post"
	BusinessConnectionUpdateType                     = "business_connection"
	BusinessMessageUpdateType                        = "business_message"
	SuccessfulPaymentBusinessMessageUpdateType       = "successful_payment_business_message"
	EditedBusinessMessageUpdateType                  = "edited_business_message"
	SuccessfulPaymentEditedBusinessMessageUpdateType = "successful_payment_edited_business_message"
	DeletedBusinessMessageUpdateType                 = "deleted_business_messages"
	MessageReactionUpdateType                        = "message_reaction"
	MessageReactionCountUpdateType                   = "message_reaction_count"
	InlineQueryUpdateType                            = "inline_query"
	ChosenInlineResultUpdateType                     = "chosen_inline_result"
	CallbackQueryUpdateType                          = "callback_query"
	ShippingQueryUpdateType                          = "shipping_query"
	PreCheckoutQueryUpdateType                       = "pre_checkout_query"
	PurchasedPaidMediaUpdateType                     = "purchased_paid_media"
	PollUpdateType                                   = "poll"
	PollAnswerUpdateType                             = "poll_answer"
	MyChatMemberUpdateType                           = "my_chat_member"
	ChatMemberUpdateType                             = "chat_member"
	ChatJoinRequestUpdateType                        = "chat_join_request"
	ChatBoostUpdateType                              = "chat_boost"
	RemovedChatBoostUpdateType                       = "removed_chat_boost"
)

// UpdatesConfig contains information about a GetUpdates request.
type UpdatesConfig struct {
	Offset                int
	Limit                 int
	Timeout               int
	AllowedUpdates        []string
	MaxMuxErrorRetryCount int           `json:"-"` // -1 = infinity retry count
	RetryDelay            time.Duration `json:"-"` // default = 3s, must be > 1s
}

func (UpdatesConfig) method() string {
	return "getUpdates"
}

// nolint:errcheck
func (config UpdatesConfig) params() (Params, error) {
	params := make(Params)

	params.AddNonZero("offset", config.Offset)
	params.AddNonZero("limit", config.Limit)
	params.AddNonZero("timeout", config.Timeout)
	params.AddInterface("allowed_updates", config.AllowedUpdates)
	return params, nil
}

// Update is an update response, from GetUpdates.
type Update struct {
	// UpdateID is the update's unique identifier.
	// Update identifiers start from a certain positive number and increase
	// sequentially.
	// This Id becomes especially handy if you're using Webhooks,
	// since it allows you to ignore repeated updates or to restore
	// the correct update sequence, should they get out of order.
	// If there are no new updates for at least a week, then identifier
	// of the next update will be chosen randomly instead of sequentially.
	UpdateId int `json:"update_id"`
	// Message new incoming message of any kind — text, photo, sticker, etc.
	//
	// optional
	Message *Message `json:"message,omitempty"`
	// EditedMessage new version of a message that is known to the bot and was
	// edited
	//
	// optional
	EditedMessage *Message `json:"edited_message,omitempty"`
	// ChannelPost new version of a message that is known to the bot and was
	// edited
	//
	// optional
	ChannelPost *Message `json:"channel_post,omitempty"`
	// EditedChannelPost new incoming channel post of any kind — text, photo,
	// sticker, etc.
	//
	// optional
	EditedChannelPost *Message `json:"edited_channel_post,omitempty"`
	// New message from a connected business account
	//
	// optional
	BusinessConnection *BusinessConnection `json:"business_connection,omitempty"`
	// New message from a connected business account
	//
	// optional
	BusinessMessage *Message `json:"business_message,omitempty"`
	// New version of a message from a connected business account
	//
	// optional
	EditedBusinessMessage *Message `json:"edited_business_message,omitempty"`
	// Messages were deleted from a connected business account
	//
	// optional
	DeletedBusinessMessage *BusinessMessageDeleted `json:"deleted_business_messages,omitempty"`
	// A reaction to a message was changed by a user. The bot must be an administrator in the chat and must explicitly specify "message_reaction" in the list of allowed_updates to receive these updates. The update isn't received for reactions set by bots.
	//
	// optional
	MessageReaction *MessageReaction `json:"message_reaction,omitempty"`
	// Reactions to a message with anonymous reactions were changed. The bot must be an administrator in the chat and must explicitly specify "message_reaction_count" in the list of allowed_updates to receive these updates. The updates are grouped and can be sent with delay up to a few minutes.
	//
	// optional
	MessageReactionCount *MessageReactionCount `json:"message_reaction_count,omitempty"`
	// InlineQuery new incoming inline query
	//
	// optional
	InlineQuery *InlineQuery `json:"inline_query,omitempty"`
	// ChosenInlineResult is the result of an inline query
	// that was chosen by a user and sent to their chat partner.
	// Please see our documentation on the feedback collecting
	// for details on how to enable these updates for your bot.
	//
	// optional
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result,omitempty"`
	// CallbackQuery new incoming callback query
	//
	// optional
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
	// ShippingQuery new incoming shipping query. Only for invoices with
	// flexible price
	//
	// optional
	ShippingQuery *ShippingQuery `json:"shipping_query,omitempty"`
	// PreCheckoutQuery new incoming pre-checkout query. Contains full
	// information about checkout
	//
	// optional
	PreCheckoutQuery *PreCheckoutQuery `json:"pre_checkout_query,omitempty"`
	// A user purchased paid media with a non-empty payload sent by the bot in a non-channel chat
	//
	// optional
	PurchasedPaidMedia *PurchasedPaidMedia `json:"purchased_paid_media,omitempty"`
	// Pool new poll state. Bots receive only updates about stopped polls and
	// polls, which are sent by the bot
	//
	// optional
	Poll *Poll `json:"poll,omitempty"`
	// PollAnswer user changed their answer in a non-anonymous poll. Bots
	// receive new votes only in polls that were sent by the bot itself.
	//
	// optional
	PollAnswer *PollAnswer `json:"poll_answer,omitempty"`
	// MyChatMember is the bot's chat member status was updated in a chat. For
	// private chats, this update is received only when the bot is blocked or
	// unblocked by the user.
	//
	// optional
	MyChatMember *ChatMemberUpdated `json:"my_chat_member,omitempty"`
	// ChatMember is a chat member's status was updated in a chat. The bot must
	// be an administrator in the chat and must explicitly specify "chat_member"
	// in the list of allowed_updates to receive these updates.
	//
	// optional
	ChatMember *ChatMemberUpdated `json:"chat_member,omitempty"`
	// ChatJoinRequest is a request to join the chat has been sent. The bot must
	// have the can_invite_users administrator right in the chat to receive
	// these updates.
	//
	// optional
	ChatJoinRequest *ChatJoinRequest `json:"chat_join_request"`
	// A chat boost was added or changed. The bot must be an administrator in the chat to receive these updates.
	//
	// optional
	ChatBoost *ChatBoostUpdated `json:"chat_boost,omitempty"`
	// A boost was removed from a chat. The bot must be an administrator in the chat to receive these updates.
	//
	// optional
	RemovedChatBoost *RemovedChatBoost `json:"removed_chat_boost,omitempty"`
}

func (u Update) GetCommand() string {
	switch {
	case u.Message != nil:
		return u.Message.Command()
	case u.EditedMessage != nil:
		return u.EditedMessage.Command()
	case u.ChannelPost != nil:
		return u.ChannelPost.Command()
	case u.EditedChannelPost != nil:
		return u.EditedChannelPost.Command()
	case u.BusinessMessage != nil:
		return u.BusinessMessage.Command()
	}
	return ""
}

func (u Update) UpdateType() string {
	switch {
	case u.Message != nil:
		if u.Message.SuccessfulPayment != nil {
			return SuccessfulPaymentMessageUpdateType
		}
		return MessageUpdateType
	case u.EditedMessage != nil:
		if u.EditedMessage.SuccessfulPayment != nil {
			return SuccessfulPaymentEditedMessageUpdateType
		}
		return EditedMessageUpdateType
	case u.ChannelPost != nil:
		return ChannelPostUpdateType
	case u.EditedChannelPost != nil:
		return EditedChannelPostUpdateType
	case u.BusinessConnection != nil:
		return BusinessConnectionUpdateType
	case u.BusinessMessage != nil:
		if u.BusinessMessage.SuccessfulPayment != nil {
			return SuccessfulPaymentBusinessMessageUpdateType
		}
		return BusinessMessageUpdateType
	case u.EditedBusinessMessage != nil:
		if u.EditedBusinessMessage.SuccessfulPayment != nil {
			return SuccessfulPaymentEditedBusinessMessageUpdateType
		}
		return EditedBusinessMessageUpdateType
	case u.DeletedBusinessMessage != nil:
		return DeletedBusinessMessageUpdateType
	case u.MessageReaction != nil:
		return MessageReactionUpdateType
	case u.MessageReactionCount != nil:
		return MessageReactionCountUpdateType
	case u.InlineQuery != nil:
		return InlineQueryUpdateType
	case u.ChosenInlineResult != nil:
		return ChosenInlineResultUpdateType
	case u.CallbackQuery != nil:
		return CallbackQueryUpdateType
	case u.ShippingQuery != nil:
		return ShippingQueryUpdateType
	case u.PreCheckoutQuery != nil:
		return PreCheckoutQueryUpdateType
	case u.PurchasedPaidMedia != nil:
		return PurchasedPaidMediaUpdateType
	case u.Poll != nil:
		return PollUpdateType
	case u.PollAnswer != nil:
		return PollAnswerUpdateType
	case u.MyChatMember != nil:
		return MyChatMemberUpdateType
	case u.ChatMember != nil:
		return ChatMemberUpdateType
	case u.ChatJoinRequest != nil:
		return ChatJoinRequestUpdateType
	case u.ChatBoost != nil:
		return ChatBoostUpdateType
	case u.RemovedChatBoost != nil:
		return RemovedChatBoostUpdateType
	}
	return ""
}

// SentFrom returns the user who sent an update. Can be nil, if Telegram did not provide information
// about the user in the update object.
func (u Update) SentFrom() *User {
	switch {
	case u.Message != nil:
		return u.Message.From
	case u.EditedMessage != nil:
		return u.EditedMessage.From
	case u.InlineQuery != nil:
		return u.InlineQuery.From
	case u.ChosenInlineResult != nil:
		return u.ChosenInlineResult.From
	case u.CallbackQuery != nil:
		return u.CallbackQuery.From
	case u.ShippingQuery != nil:
		return u.ShippingQuery.From
	case u.PreCheckoutQuery != nil:
		return u.PreCheckoutQuery.From
	default:
		return nil
	}
}

// CallbackData returns the callback query data, if it exists.
func (u Update) CallbackData() string {
	if u.CallbackQuery != nil {
		return u.CallbackQuery.Data
	}
	return ""
}

// FromChat returns the chat where an update occurred.
func (u Update) FromChat() *Chat {
	switch {
	case u.Message != nil:
		return u.Message.Chat
	case u.EditedMessage != nil:
		return u.EditedMessage.Chat
	case u.ChannelPost != nil:
		return u.ChannelPost.Chat
	case u.EditedChannelPost != nil:
		return u.EditedChannelPost.Chat
	case u.CallbackQuery != nil:
		return u.CallbackQuery.Message.Chat
	default:
		return nil
	}
}
