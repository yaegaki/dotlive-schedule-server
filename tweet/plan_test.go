package tweet

import (
	"testing"

	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/plan"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

func TestPlanParser(t *testing.T) {
	tests := []struct {
		name      string
		tweetDate jst.Time
		tweet     string
		parts     []EntryPart
	}{
		{
			"2020/2/26",
			jst.ShortDate(2020, 2, 25),
			`【どっとライブ】【アイドル部】
【生放送スケジュール2月26日】

19:00~: #ヤマトイオリ
21:00~: #カルロピノ
22:00~: #神楽すず
23:00~: #メリーミルク

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPart(Iori, 19, 00),
				CreateEntryPart(Pino, 21, 00),
				CreateEntryPart(Suzu, 22, 00),
				CreateEntryPart(Milk, 23, 00),
			},
		},
		{
			"2020/2/28",
			jst.ShortDate(2020, 2, 27),
			`【どっとライブ】【アイドル部】
【生放送スケジュール2月28日】

20:00~: #シロ生放送 (bilibili)
22:00~: #北上双葉
23:00~: #神楽すず

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPartBilibili(Siro, 20, 00),
				CreateEntryPart(Futaba, 22, 00),
				CreateEntryPart(Suzu, 23, 00),
			},
		},
		{
			"2020/2/22",
			jst.ShortDate(2020, 2, 21),
			`【生放送スケジュール2月22日】

12:00~: #神楽すず
15:00~: #カルロピノ
18:00~: #北上双葉
19:00~: #花京院ちえり
20:00~: #Vに国境はいらない（出演時間21:30~を予定）
24:00~: #カルロピノ

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPart(Suzu, 12, 00),
				CreateEntryPart(Pino, 15, 00),
				CreateEntryPart(Futaba, 18, 00),
				CreateEntryPart(Chieri, 19, 00),
				CreateEntryPart(Pino, 24, 00),
			},
		},
		{
			"2020/4/24",
			jst.ShortDate(2020, 4, 23),
			`【どっとライブ】【アイドル部】
【生放送スケジュール4月24日】

19:00~: #シロ生放送 (bilibili)
22:00~: #神楽すず

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPartBilibili(Siro, 19, 00),
				CreateEntryPart(Suzu, 22, 00),
			},
		},
		{
			"Empty",
			jst.ShortDate(2090, 4, 1),
			`【どっとライブ】【アイドル部】
【生放送スケジュール4月2日】


メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparePlan(t, Tweet{
				ID:   "temp",
				Date: tt.tweetDate,
				Text: tt.tweet,
			}, CreatePlan(tt.tweetDate.AddOneDay(), tt.parts))
		})
	}
}

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
