package store

import (
	"context"

	"cloud.google.com/go/firestore"
)

var globalClient *firestore.Client

// Init キャッシュの初期化
func Init() {
	var err error
	globalClient, err = firestore.NewClient(context.Background(), firestore.DetectProjectID)
	if err != nil {
		panic(err)
	}
}

// GetClient キャッシュされたクライアントを取得する
func GetClient() *firestore.Client {
	return globalClient
}

// CloseClient キャッシュされたクライアントを閉じる
func CloseClient() {
	globalClient.Close()
}
