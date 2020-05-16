package cache

import (
	"context"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

var actors model.ActorSlice
var actorMutex sync.RWMutex

// FindActorsWithCache キャッシュかストアから配信者情報を取得する
func FindActorsWithCache(ctx context.Context, cli *firestore.Client) (model.ActorSlice, error) {
	a := GetActors()
	if a != nil {
		return a, nil
	}

	return store.FindActors(ctx, cli)
}

// GetActors キャッシュから配信者情報を取得する
func GetActors() model.ActorSlice {
	actorMutex.RLock()
	defer actorMutex.RUnlock()

	return actors
}

// SetActors 配信者情報をキャッシュする
func SetActors(a model.ActorSlice) {
	actorMutex.Lock()
	defer actorMutex.Unlock()

	actors = a
}