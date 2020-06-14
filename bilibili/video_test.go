package bilibili

import "testing"

func TestIsBilibiliURL(t *testing.T) {
	urls := []string{
		"https://live.bilibili.com/21307497",
	}

	for _, u := range urls {
		if !IsBilibiliURL(u) {
			t.Fatalf("fail: %v", u)
		}
	}
}
