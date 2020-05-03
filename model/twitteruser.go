package model

// TwitterUser ツイッターのユーザー
// Actor以外のツイッターユーザーで使用する
type TwitterUser struct {
	// ScreenName twitterのScreenName
	ScreenName string
	// LastTweetID 最後にツイートしたツイートのID
	LastTweetID string
}
