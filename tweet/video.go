package tweet

import (
	"log"

	"github.com/ChimeraCoder/anaconda"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// VideoResolver 動画情報を解決する
type VideoResolver interface {
	// Except 当該URLがツイートに含まれている場合はResolveを呼ばない
	Except(url string) bool
	// Resolve URLから動画情報を取得して保存する
	Resolve(tweet Tweet, url string, actor model.Actor) error
	// Mark 読み取った最後のTweetIDを保存する
	Mark(tweetID string, actor model.Actor) error
}

// ResolveVideos Twitterから動画情報を取得する
func ResolveVideos(api *anaconda.TwitterApi, actors []model.Actor, r VideoResolver) {
	for _, actor := range actors {
		tl, err := getTimeline(api, actor.TwitterScreenName, actor.LastTweetID, "")
		if err != nil {
			log.Printf("Can not get tweet for %v: %v", actor.Name, err)
			continue
		}

		hasError := false
		lastTweetID := ""

		for _, tweet := range tl {
			if lastTweetID == "" {
				lastTweetID = tweet.ID
			}

			err := resolveVideoForTweet(r, actor, tweet)
			if err != nil {
				log.Printf("Can not resolve video for %v: %v", actor.Name, err)
				hasError = true
				break
			}

			if tweet.QuotedTweet != nil {
				err = resolveVideoForTweet(r, actor, *tweet.QuotedTweet)
				if err != nil {
					log.Printf("(QuatedTweet)Can not resolve video for %v: %v", actor.Name, err)
					hasError = true
					break
				}
			}
		}

		if hasError {
			continue
		}

		if lastTweetID == "" {
			continue
		}

		err = r.Mark(lastTweetID, actor)
		if err != nil {
			log.Printf("Can not mark last tweetID for %v: %v", actor.Name, err)
		}
	}
}

func resolveVideoForTweet(r VideoResolver, actor model.Actor, tweet Tweet) error {
	except := false
	for _, url := range tweet.URLs {
		except = r.Except(url)
		if except {
			break
		}
	}

	if except {
		return nil
	}

	for _, url := range tweet.URLs {
		err := r.Resolve(tweet, url, actor)
		if err != nil {
			return err
		}
	}

	return nil
}
