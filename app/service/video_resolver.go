package service

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/bilibili"
	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/mildom"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
	"github.com/yaegaki/dotlive-schedule-server/tweet"
	"github.com/yaegaki/dotlive-schedule-server/youtube"
	"golang.org/x/oauth2/google"
	"golang.org/x/xerrors"
	y "google.golang.org/api/youtube/v3"
)

// VideoResolver ビデオ情報の解決をする
type VideoResolver struct {
	ctx            context.Context
	c              *firestore.Client
	youtubeService *y.Service
}

// NewVideoResolver videoResolverを作成する
func NewVideoResolver(ctx context.Context, c *firestore.Client) (*VideoResolver, error) {
	httpClient, err := google.DefaultClient(ctx, y.YoutubeReadonlyScope)
	if err != nil {
		return nil, xerrors.Errorf("Can not create http client: %w", err)
	}

	youtubeService, err := y.New(httpClient)
	if err != nil {
		return nil, xerrors.Errorf("Can not create youtube service:%w", err)
	}

	return &VideoResolver{
		ctx:            ctx,
		c:              c,
		youtubeService: youtubeService,
	}, nil
}

// Except impl tweet.VideoResolver
func (r *VideoResolver) Except(url string) bool {
	// チャンネルURLが含まれる場合は配信の告知ではない
	return youtube.IsYoutubeChannelURL(url)
}

// Resolve impl tweet.VideoResolver
func (r *VideoResolver) Resolve(tweet tweet.Tweet, url string, actor model.Actor) error {
	var v model.Video
	var err error

	if youtube.IsYoutubeURL(url) {
		v, err = youtube.FindVideo(r.ctx, r.youtubeService, url, actor, tweet.Date)
	} else if bilibili.IsBilibiliURL(url) {
		v, err = bilibili.FindVideo(url, actor, tweet.Date)
	} else if mildom.IsMildomURL(url) {
		v, err = mildom.FindVideo(url, actor, tweet.Date)
	} else {
		return nil
	}

	if err == common.ErrInvalidChannel {
		return nil
	}

	if err != nil {
		return xerrors.Errorf("Can not get video(%v): %w", url, err)
	}
	v.Text = tweet.Text
	v.HashTags = tweet.HashTags

	err = r.save(v, tweet)
	if err != nil {
		return xerrors.Errorf("Can not save video(%v): %w", v.ID, err)
	}

	return nil
}

func (r *VideoResolver) save(v model.Video, tweet tweet.Tweet) error {
	return store.SaveVideo(r.ctx, r.c, v, func(oldVideo model.Video) bool {
		// 過去の動画についてツイートしたときに上書きされると微妙なので
		// 動画の開始時間から1日後より以前の時間のツイートなら情報を更新する
		return tweet.Date.Before(oldVideo.StartAt.AddOneDay())
	})
}

// Mark impl tweet.VideoResolver
func (r *VideoResolver) Mark(tweetID string, actor model.Actor) error {
	actor.LastTweetID = tweetID
	return store.SaveActor(r.ctx, r.c, actor)
}

// YoutubeService YoutubeServiceを取得する
func (r *VideoResolver) YoutubeService() *y.Service {
	return r.youtubeService
}
