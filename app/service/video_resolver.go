package service

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/bilibili"
	"github.com/yaegaki/dotlive-schedule-server/common"
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

// Resolve impl tweet.VideoResolver
func (r *VideoResolver) Resolve(tweet tweet.Tweet, url string, actor model.Actor) error {
	if youtube.IsYoutubeURL(url) {
		return r.resolveYoutubeVideo(tweet, url, actor)
	} else if bilibili.IsBilibiliURL(url) {
		return r.resolveBilibiliVideo(tweet, url, actor)
	}

	return nil
}

func (r *VideoResolver) resolveYoutubeVideo(tweet tweet.Tweet, url string, actor model.Actor) error {
	v, err := youtube.FindVideo(r.ctx, r.youtubeService, url, actor)
	if err == common.ErrInvalidChannel {
		return nil
	}

	if err != nil {
		return xerrors.Errorf("Can not get youtube video(%v): %w", url, err)
	}
	v.Text = tweet.Text

	err = r.save(v)
	if err != nil {
		return xerrors.Errorf("Can not save youtube video(%v): %w", v.ID, err)
	}

	return nil
}

func (r *VideoResolver) resolveBilibiliVideo(tweet tweet.Tweet, url string, actor model.Actor) error {
	v, err := bilibili.FindVideo(url, actor, tweet.ID, tweet.Date)
	if err == common.ErrInvalidChannel {
		return nil
	}

	if err != nil {
		return xerrors.Errorf("Can not get bilibili video(%v) info: %w", url, err)
	}
	v.Text = tweet.Text

	err = r.save(v)
	if err != nil {
		return xerrors.Errorf("Can not save bilibili video(%v): %w", v.ID, err)
	}

	return nil
}

func (r *VideoResolver) save(v model.Video) error {
	return store.SaveVideo(r.ctx, r.c, v)
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
