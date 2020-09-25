package tweet

import "github.com/yaegaki/dotlive-schedule-server/jst"

// Tweet ツイート
type Tweet struct {
	// ID ツイートID
	ID string
	// UserName ユーザー名
	UserName string
	// Text ツイート内容
	Text string
	// QuotedTweet 引用リツイートの引用先
	QuotedTweet *Tweet
	// Date ツイート時刻
	Date jst.Time
	// URLs ツイートに含まれるURL
	URLs []string
	// MediaURLs ツイートに含まれるメディアのURL
	MediaURLs []string
	// HashTags ツイートに含まれるハッシュタグ
	HashTags []string
}
