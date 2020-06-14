package youtube

import "testing"

func TestIsYoutubeURL(t *testing.T) {
	urls := []string{
		"http://youtu.be/6bzVDa28dj4",
		"https://youtu.be/6bzVDa28dj4",
	}

	for _, u := range urls {
		if !IsYoutubeURL(u) {
			t.Fatalf("fail: %v", u)
		}
	}
}
