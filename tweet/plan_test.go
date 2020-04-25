package tweet

import (
	"testing"

	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/plan"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

func comparePlan(t *testing.T, tweet Tweet, expect model.Plan) {
	pp := planParser{
		actors: All,
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

		if e.Source != expectEntry.Source {
			t.Errorf("Invalid source, got: %v expect: %v", e.Source, expectEntry.Source)
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
			CreateEntry(date, Iori, 19, 00),
			CreateEntry(date, Pino, 21, 00),
			CreateEntry(date, Suzu, 22, 00),
			CreateEntry(date, Milk, 23, 00),
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
			CreateEntryBilibili(date, Siro, 20, 00),
			CreateEntry(date, Futaba, 22, 00),
			CreateEntry(date, Suzu, 23, 00),
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
			CreateEntry(date, Suzu, 12, 00),
			CreateEntry(date, Pino, 15, 00),
			CreateEntry(date, Futaba, 18, 00),
			CreateEntry(date, Chieri, 19, 00),
			CreateEntry(date, Pino, 24, 00),
		},
	})
}

func TestPlanParser4(t *testing.T) {
	text := `【どっとライブ】【アイドル部】
【生放送スケジュール4月24日】

19:00~: #シロ生放送 (bilibili)
22:00~: #神楽すず

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	tweet := Tweet{
		ID:   "tempID",
		Text: text,
		Date: jst.ShortDate(2020, 4, 23),
	}

	date := tweet.Date.AddOneDay()
	comparePlan(t, tweet, model.Plan{
		Date: date,
		Entries: []model.PlanEntry{
			CreateEntryBilibili(date, Siro, 19, 00),
			CreateEntry(date, Suzu, 22, 00),
		},
	})
}
