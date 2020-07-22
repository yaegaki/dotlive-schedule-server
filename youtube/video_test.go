package youtube

import "testing"

func TestIsYoutubeChannelURL(t *testing.T) {
	if !IsYoutubeChannelURL("https://www.youtube.com/channel/UCP9ZgeIJ3Ri9En69R0kJc9Q") {
		t.Errorf("Channel URL")
	}

	if IsYoutubeChannelURL("https://www.youtube.com/watch?v=bVsei7pIrbk") {
		t.Errorf("not Channel URL")
	}
}

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
