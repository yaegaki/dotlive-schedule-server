package internal

import "os"

// IsDevelop 開発環境かどうか
var IsDevelop bool

func init() {
	IsDevelop = os.Getenv("DEVELOP") == "true"
}
