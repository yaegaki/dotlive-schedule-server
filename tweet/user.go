package tweet

import (
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// GetProfileImageURL Twitterのアイコン画像のURLを取得する
func GetProfileImageURL(api *anaconda.TwitterApi, actor model.Actor) (string, error) {
	u, err := api.GetUsersShow(actor.TwitterScreenName, url.Values{})
	if err != nil {
		return "", nil
	}

	return u.ProfileImageUrlHttps, nil
}
