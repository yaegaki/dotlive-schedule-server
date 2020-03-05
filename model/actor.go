package model

// Actor 配信者
type Actor struct {
	// ID 配信者ID
	ID string
	// Name 名前
	Name string
	// Hashtag ハッシュタグ
	Hashtag string
	// TwitterScreenName Twitterのスクリーンネーム
	TwitterScreenName string
	// Emoji 推しアイコン
	Emoji string
	// YoutubeChannelID YoutubeのチャンネルID
	YoutubeChannelID string
	// BilibiliID BilibiliのID
	BilibiliID string
	// LastTweetID 最後に取得したTweetのID
	LastTweetID string
}
