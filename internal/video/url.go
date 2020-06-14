package video

import (
	"net/url"
	"strings"
)

// SourceURLSlice 動画ソースのURLのprefix
type SourceURLSlice []string

// IsTargetVideoSource ターゲットURLが動画ソースのURLかどうかをチェックする
func IsTargetVideoSource(sourceURLs SourceURLSlice, targetURL string) bool {
	for _, u := range sourceURLs {
		u1, err := url.Parse(u)
		if err != nil {
			return false
		}

		u2, err := url.Parse(targetURL)
		if err != nil {
			return false
		}

		if u1.Host != u2.Host {
			continue
		}

		if strings.HasPrefix(u2.Path, u1.Path) {
			return true
		}
	}

	return false
}
