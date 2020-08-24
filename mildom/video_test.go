package mildom

import (
	"testing"

	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

var urls = []string{
	"https://www.mildom.com/10596535",
	"https://mildom.com/10596535",
}

func TestIsMildomURL(t *testing.T) {
	for _, u := range urls {
		if !IsMildomURL(u) {
			t.Fatalf("fail: %v", u)
		}
	}
}

func TestFindVideo(t *testing.T) {
	url := urls[0]
	actor := model.Actor{
		ID:       "test",
		MildomID: "10596535",
	}
	date := jst.ShortDate(2020, 6, 3)
	v, err := FindVideo(url, actor, date)
	if err != nil {
		t.Fatalf("fail: %v", err)
	}

	expectID := "2020-6-3-mildom-" + actor.ID
	if v.ID != expectID {
		t.Fatalf("invalid id, got: %v expect: %v", v.ID, expectID)
	}

	if v.URL != url {
		t.Fatalf("invalid url, got: %v expect: %v", v.URL, url)
	}

	_, err = FindVideo("https://www.mildom.com/profile/10596535", actor, date)
	if err == nil {
		t.Fatalf("profile page")
	}
}
