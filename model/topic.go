package model

// Topic メッセージングのトピックを表す
type Topic struct {
	// Name トピックの名前
	Name string `json:"name"`
	// DisplayName トピックの表示用の名前
	DisplayName string `json:"displayName"`
	// Subscribed 購読しているかどうか
	Subscribed bool `json:"subscribed"`
}
