package tweet

import (
	"testing"

	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

func comparePlan(t *testing.T, tweet Tweet, expect model.Plan) {
	pp := planParser{
		actors: actors,
	}

	p, err := pp.parse(tweet)
	if err != nil {
		t.Errorf("Can not parse tweet: %v", err)
		return
	}

	if !p.Date.Equal(expect.Date) {
		t.Errorf("invalid Date, got: %v expect: %v", p.Date, expect.Date)
	}

	if len(p.Entries) != len(expect.Entries) {
		t.Errorf("different entry, got: %v expect: %v", len(p.Entries), len(expect.Entries))
	}

	for i, e := range p.Entries {
		expectEntry := expect.Entries[i]

		if e.ActorID != expectEntry.ActorID {
			t.Errorf("invalid ActorID, got: %v expect: %v", e.ActorID, expectEntry.ActorID)
			continue
		}

		if !e.StartAt.Equal(expectEntry.StartAt) {
			t.Errorf("invalid StartAt, got: %v expect: %v", e.StartAt, expectEntry.StartAt)
			continue
		}
	}
}

func TestPlanParser(t *testing.T) {

	text := `【どっとライブ】【アイドル部】
【生放送スケジュール2月26日】

19:00~: #ヤマトイオリ
21:00~: #カルロピノ
22:00~: #神楽すず
23:00~: #メリーミルク

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	tweet := Tweet{
		ID:   "tempID",
		Text: text,
		Date: jst.ShortDate(2020, 2, 25),
	}

	date := tweet.Date.AddOneDay()
	comparePlan(t, tweet, model.Plan{
		Date: date,
		Entries: []model.PlanEntry{
			createEntry(date, iori, 19, 00),
			createEntry(date, pino, 21, 00),
			createEntry(date, suzu, 22, 00),
			createEntry(date, milk, 23, 00),
		},
	})
}

func TestPlanParser2(t *testing.T) {
	text := `【どっとライブ】【アイドル部】
【生放送スケジュール2月28日】

20:00~: #シロ生放送 (bilibili)
22:00~: #北上双葉
23:00~: #神楽すず

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	tweet := Tweet{
		ID:   "tempID",
		Text: text,
		Date: jst.ShortDate(2020, 2, 27),
	}

	date := tweet.Date.AddOneDay()
	comparePlan(t, tweet, model.Plan{
		Date: date,
		Entries: []model.PlanEntry{
			createEntry(date, siro, 20, 00),
			createEntry(date, futaba, 22, 00),
			createEntry(date, suzu, 23, 00),
		},
	})
}

func TestPlanParser3(t *testing.T) {
	text := `【生放送スケジュール2月22日】

12:00~: #神楽すず
15:00~: #カルロピノ
18:00~: #北上双葉
19:00~: #花京院ちえり
20:00~: #Vに国境はいらない（出演時間21:30~を予定）
24:00~: #カルロピノ

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	tweet := Tweet{
		ID:   "tempID",
		Text: text,
		Date: jst.ShortDate(2020, 2, 21),
	}

	date := tweet.Date.AddOneDay()
	comparePlan(t, tweet, model.Plan{
		Date: date,
		Entries: []model.PlanEntry{
			createEntry(date, suzu, 12, 00),
			createEntry(date, pino, 15, 00),
			createEntry(date, futaba, 18, 00),
			createEntry(date, chieri, 19, 00),
			createEntry(date, pino, 24, 00),
		},
	})
}

func createEntry(date jst.Time, actor model.Actor, hour, min int) model.PlanEntry {
	return model.PlanEntry{
		ActorID: actor.ID,
		StartAt: jst.Date(date.Year(), date.Month(), date.Day(), hour, min),
	}
}

var iori = model.Actor{
	ID:      "iori",
	Hashtag: "#ヤマトイオリ",
}

var pino = model.Actor{
	ID:      "pino",
	Hashtag: "#カルロピノ",
}

var suzu = model.Actor{
	ID:      "suzu",
	Hashtag: "#神楽すず",
}

var chieri = model.Actor{
	ID:      "chieri",
	Hashtag: "#花京院ちえり",
}

var iroha = model.Actor{
	ID:      "iroha",
	Hashtag: "#金剛いろは",
}

var futaba = model.Actor{
	ID:      "futaba",
	Hashtag: "#北上双葉",
}

var mememe = model.Actor{
	ID:      "mememe",
	Hashtag: "#もこ田めめめ",
}

var siro = model.Actor{
	ID:      "siro",
	Hashtag: "#シロ生放送",
}

var milk = model.Actor{
	ID:      "milk",
	Hashtag: "#メリーミルク",
}

var actors = []model.Actor{
	iori,
	pino,
	suzu,
	chieri,
	iroha,
	futaba,
	mememe,
	siro,
	milk,
}
