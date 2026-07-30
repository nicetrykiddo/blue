package models

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	ChannelPost   *Message       `json:"channel_post,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageID       int             `json:"message_id"`
	MessageThreadID int             `json:"message_thread_id,omitempty"`
	From            *User           `json:"from,omitempty"`
	Chat            *Chat           `json:"chat"`
	ReplyToMessage  *Message        `json:"reply_to_message,omitempty"`
	Date            int             `json:"date"`
	Text            string          `json:"text,omitempty"`
	Entities        []MessageEntity `json:"entities,omitempty"`
	Document        *Document       `json:"document,omitempty"`
}

type MessageEntity struct {
	Type          string `json:"type"`
	Offset        int    `json:"offset"`
	Length        int    `json:"length"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type SendMessageResponse struct {
	OK          bool    `json:"ok"`
	Result      Message `json:"result"`
	Description string  `json:"description,omitempty"`
}

type GetChatResponse struct {
	OK          bool   `json:"ok"`
	Result      Chat   `json:"result"`
	Description string `json:"description,omitempty"`
}

type File struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     int    `json:"file_size,omitempty"`
	FilePath     string `json:"file_path,omitempty"`
}

type Document struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileName     string `json:"file_name,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	FileSize     int    `json:"file_size,omitempty"`
}

type Sticker struct {
	FileID        string `json:"file_id"`
	FileUniqueID  string `json:"file_unique_id"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	IsAnimated    bool   `json:"is_animated"`
	IsVideo       bool   `json:"is_video"`
	Emoji         string `json:"emoji,omitempty"`
	SetName       string `json:"set_name,omitempty"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

type StickerSet struct {
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Stickers []Sticker `json:"stickers"`
}

type GetStickerSetResponse struct {
	OK          bool       `json:"ok"`
	Result      StickerSet `json:"result"`
	Description string     `json:"description,omitempty"`
}

type GetStickersResponse struct {
	OK          bool      `json:"ok"`
	Result      []Sticker `json:"result"`
	Description string    `json:"description,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data"`
}

type ForumTopic struct {
	MessageThreadID   int    `json:"message_thread_id"`
	Name              string `json:"name"`
	IconColor         int    `json:"icon_color,omitempty"`
	IconCustomEmojiID string `json:"icon_custom_emoji_id,omitempty"`
}

type CreateForumTopicResponse struct {
	OK          bool       `json:"ok"`
	Result      ForumTopic `json:"result"`
	Description string     `json:"description,omitempty"`
}
