package model

// AwaiSenseiSchedule あわい先生のどっとライブスケジュールの情報
type AwaiSenseiSchedule struct {
	// TweetID 元ツイートのID
	TweetID string `json:"tweetId"`
	// Title タイトル
	Title string `json:"title"`
	// ImageURL 画像のURL
	ImageURL string `json:"imageURL"`
}
