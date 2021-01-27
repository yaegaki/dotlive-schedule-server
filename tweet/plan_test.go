package tweet

import (
	"sort"
	"strings"
	"testing"

	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/plan"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

func TestParsePlanText(t *testing.T) {
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
				CreateEntryPartCollaboHashTag(20, 00, "#Vに国境はいらない（出演時間21:30~を予定）"),
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
			"2020/4/19",
			jst.ShortDate(2020, 4, 18),
			`【どっとライブ】【アイドル部】
【生放送スケジュール4月19日】

20:00~: #花京院ちえり × #カルロピノ × #金剛いろは × #もこ田めめめ
22:00~: #神楽すず

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPartCollabo(Chieri, 20, 00, 1),
				CreateEntryPartCollabo(Pino, 20, 00, 1),
				CreateEntryPartCollabo(Iroha, 20, 00, 1),
				CreateEntryPartCollabo(Mememe, 20, 00, 1),
				CreateEntryPart(Suzu, 22, 00),
			},
		},
		{
			"2020/7/28",
			jst.ShortDate(2020, 7, 27),
			// 日付がバグっている場合
			`【どっとライブ】【アイドル部】
【生放送スケジュール7月2日】

20:00~: #北上双葉
21:00~: #八重沢なとり
22:00~: #花京院ちえり
23:00~: #ヤマトイオリ

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPart(Futaba, 20, 00),
				CreateEntryPart(Natori, 21, 00),
				CreateEntryPart(Chieri, 22, 00),
				CreateEntryPart(Iori, 23, 00),
			},
		},
		{
			"2020/10/27",
			jst.ShortDate(2020, 11, 26),
			// メンバー限定
			`【生放送スケジュール11月27日】

13:00~: #八重沢なとり
19:00~: #はんぱない文化祭 (前夜祭)
22:00~: #神楽すず
23:00~: #カルロピノ (メンバーシップ限定)
24:00~: #北上双葉

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPart(Natori, 13, 00),
				CreateEntryPartCollaboHashTag(19, 00, "#はんぱない文化祭 (前夜祭)"),
				CreateEntryPart(Suzu, 22, 00),
				CreateEntryPartMemberOnly(Pino, 23, 00),
				CreateEntryPart(Futaba, 24, 00),
			},
		},
		{
			"2020/11/29",
			jst.ShortDate(2020, 11, 28),
			`【どっとライブ】【アイドル部】
【生放送スケジュール11月29日】

 0:00~: #はんぱない文化祭

メンバーの動画、SNSのリンクはこちらから！
vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPartCollaboHashTag(00, 00, "#はんぱない文化祭"),
			},
		},
		{
			"2020/12/15",
			jst.ShortDate(2020, 12, 14),
			`【どっとライブ】【アイドル部】
【生放送スケジュール12月15日】

13:00~: #八重沢なとり
20:00~: #神楽すず
23:00~: ＃Vのから騒ぎ

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPart(Natori, 13, 00),
				CreateEntryPart(Suzu, 20, 00),
				CreateEntryPartCollaboHashTag(23, 00, "#Vのから騒ぎ"),
			},
		},
		{
			"2020/12/17(2)",
			jst.ShortDate(2020, 12, 16),
			// オリジナルは12月17日の後に②がついてる
			`【どっとライブ】【アイドル部】
【生放送スケジュール12月17日】

20:00~: #シロ生放送
21:00~: #神楽すず(Mildom) × #花京院ちえり(Mildom) × #ヤマトイオリ(Mildom) × #金剛いろは 

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPart(Siro, 20, 00),
				CreateEntryPartCollaboMildom(Suzu, 21, 00, 1),
				CreateEntryPartCollaboMildom(Chieri, 21, 00, 1),
				CreateEntryPartCollaboMildom(Iori, 21, 00, 1),
				CreateEntryPartCollabo(Iroha, 21, 00, 1),
			},
		},
		{
			"2099/4/19",
			jst.ShortDate(2099, 4, 18),
			`【どっとライブ】【アイドル部】
【生放送スケジュール4月19日】

20:00~: #花京院ちえり × #カルロピノ × #金剛いろは × #もこ田めめめ
21:00~: #神楽すず × #ヤマトイオリ

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
			[]EntryPart{
				CreateEntryPartCollabo(Chieri, 20, 00, 1),
				CreateEntryPartCollabo(Pino, 20, 00, 1),
				CreateEntryPartCollabo(Iroha, 20, 00, 1),
				CreateEntryPartCollabo(Mememe, 20, 00, 1),
				CreateEntryPartCollabo(Suzu, 21, 00, 2),
				CreateEntryPartCollabo(Iori, 21, 00, 2),
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
			}, CreatePlan(tt.tweetDate.AddOneDay(), tt.parts), false)
		})
	}

	t.Run("strict", func(t *testing.T) {
		tweet := Tweet{
			ID:   "temp",
			Date: jst.ShortDate(2020, 4, 18),
			Text: `【どっとライブ】【アイドル部】
【生放送スケジュール4月19日】

20:00~: #ヤマトイオリ × #hoge

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
		}

		_, err := ParsePlanTweet(tweet, All, false)
		if err != nil {
			t.Error("can not create plan")
			return
		}

		_, err = ParsePlanTweet(tweet, All, true)
		if err == nil {
			t.Error("invalid strict mode")
		}
	})

	t.Run("fix plan", func(t *testing.T) {
		tweet := Tweet{
			ID:   "temp",
			Date: jst.ShortDate(2020, 8, 24),
			Text: `変更のお知らせ
【生放送スケジュール8月24日】

13:00~: #八重沢なとり
19:00~: #もこ田めめめ (Mildom)
21:00~: #ヤマトイオリ
22:00~: #神楽すず × アキロゼさん
23:00~: #北上双葉

メンバーのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
		}

		parts := []EntryPart{
			CreateEntryPart(Natori, 13, 00),
			CreateEntryPartMildom(Mememe, 19, 00),
			CreateEntryPart(Iori, 21, 00),
			CreateEntryPart(Suzu, 22, 00),
			CreateEntryPart(Futaba, 23, 00),
		}

		comparePlan(t, tweet, CreatePlan(tweet.Date, parts), false)
	})

	t.Run("plan.Text", func(t *testing.T) {
		tweet := Tweet{
			ID:   "temp",
			Date: jst.ShortDate(2020, 8, 24),
			Text: `【生放送スケジュール8月24日】

13:00~: #八重沢なとり
19:00~: #もこ田めめめ (Mildom)
21:00~: #ヤマトイオリ
22:00~: #神楽すず × アキロゼさん
23:00~: #北上双葉
25:00~: #hogehoge

メンバーのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
		}

		p, err := ParsePlanTweet(tweet, All, false)
		if err != nil {
			t.Errorf("Can not parse tweet: %v", err)
			return
		}

		expect := `13:00~: #八重沢なとり
19:00~: #もこ田めめめ (Mildom)
21:00~: #ヤマトイオリ
22:00~: #神楽すず × アキロゼさん
23:00~: #北上双葉
25:00~: #hogehoge`

		if p.Text() != expect {
			t.Errorf("Invalid text: %v", err)
		}
	})

	t.Run("Additional", func(t *testing.T) {
		tweet := Tweet{
			ID:   "temp",
			Date: jst.ShortDate(2020, 9, 24),
			Text: `【
【どっとライブ】【アイドル部】
【生放送スケジュール9月24日】

10:00~: #八重沢なとり
20:00~: #神楽すず (Mildom)
20:00~: #電脳少女ガッチマンV (Siro Channel)

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
		}

		p, err := ParsePlanTweet(tweet, All, false)
		if err != nil {
			t.Fatalf("Can not parse tweet: %v", err)
			return
		}

		if p.PlanTag != "" {
			t.Fatal("invalid plantag")
		}

		tweet = Tweet{
			ID:   "temp",
			Date: jst.ShortDate(2020, 9, 24),
			Text: `【
【どっとライブ】【アイドル部】
【生放送スケジュール9月24日①】

10:00~: #八重沢なとり
20:00~: #神楽すず (Mildom)
20:00~: #電脳少女ガッチマンV (Siro Channel)

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
		}

		p, err = ParsePlanTweet(tweet, All, false)
		if err != nil {
			t.Fatalf("Can not parse tweet: %v", err)
			return
		}

		if p.PlanTag != "①" {
			t.Fatalf("invalid plantag, got:%v expect:①", p.PlanTag)
		}

		for _, e := range p.Entries {
			if e.PlanTag != "①" {
				t.Fatalf("invalid plantag, got:%v expect:①", e.PlanTag)
			}
		}

		tweet = Tweet{
			ID:   "temp",
			Date: jst.ShortDate(2020, 9, 24),
			Text: `【どっとライブ】【アイドル部】
【生放送スケジュール9月24日②】

21:00~: #電脳少女ガッチマンV (ガッチマンVさんチャンネル)
22:00~: #ヤマトイオリ
23:00~: #Vのから騒ぎ

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`,
		}

		p, err = ParsePlanTweet(tweet, All, false)
		if err != nil {
			t.Fatalf("Can not parse tweet: %v", err)
			return
		}

		if p.PlanTag != "②" {
			t.Fatalf("invalid plantag, got:%v expect:②", p.PlanTag)
		}

		for _, e := range p.Entries {
			if e.PlanTag != "②" {
				t.Fatalf("invalid plantag, got:%v expect:②", e.PlanTag)
			}
		}
	})
}

func comparePlan(t *testing.T, tweet Tweet, expect model.Plan, testText bool) {
	p, err := ParsePlanTweet(tweet, All, false)
	if err != nil {
		t.Errorf("Can not parse tweet: %v", err)
		return
	}

	if !p.Date.Equal(expect.Date) {
		t.Errorf("invalid Date, got: %v expect: %v", p.Date, expect.Date)
	}

	if p.PlanTag != expect.PlanTag {
		t.Errorf("invalid PlanTag, got: %v expect: %v", p.PlanTag, expect.PlanTag)
	}

	if len(p.Entries) != len(expect.Entries) {
		t.Errorf("different entry, got: %v expect: %v", len(p.Entries), len(expect.Entries))
		return
	}

	createSortedEntries := func(entries []model.PlanEntry) []model.PlanEntry {
		c := append([]model.PlanEntry{}, entries...)
		sort.Slice(c, func(i, j int) bool {
			if c[i].StartAt.Equal(c[j].StartAt) {
				// 開始時間が同じ場合はid順
				return strings.Compare(c[i].ActorID, c[j].ActorID) < 0
			}

			// 開始時間でソート
			return c[i].StartAt.Before(c[j].StartAt)
		})
		return c
	}

	expectEntries := createSortedEntries(expect.Entries)

	for i, e := range createSortedEntries(p.Entries) {
		expectEntry := expectEntries[i]

		if e.ActorID != expectEntry.ActorID {
			t.Errorf("invalid ActorID, got: %v expect: %v", e.ActorID, expectEntry.ActorID)
			continue
		}

		if e.PlanTag != expectEntry.PlanTag {
			t.Errorf("invalid PlanTag, got: %v expect: %v", e.PlanTag, expectEntry.PlanTag)
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

		if e.CollaboID != expectEntry.CollaboID {
			t.Errorf("invalid CollaboID, got: %v expect: %v", e.CollaboID, expectEntry.CollaboID)
		}

		if e.MemberOnly != expectEntry.MemberOnly {
			t.Errorf("invalid MemberOnly, got: %v expect: %v", e.MemberOnly, expectEntry.MemberOnly)
		}
	}

	if testText {
		if len(p.Texts) != len(expect.Texts) {
			t.Errorf("different text, got: %v expect: %v", len(p.Texts), len(expect.Texts))
			return
		}

		for i, text := range p.Texts {
			expectText := expect.Texts[i]
			if !text.Date.Equal(expectText.Date) {
				t.Errorf("invalid Text.Date, got: %v expect: %v", text.Date, expectText.Date)
				continue
			}

			if text.PlanTag != expectText.PlanTag {
				t.Errorf("invalid Text.PlanTag, got: %v expect: %v", text.PlanTag, expectText.PlanTag)
				continue
			}

			if text.Text != expectText.Text {
				t.Errorf("invalid Text.Text, got: %v expect: %v", text.Text, expectText.Text)
				continue
			}
		}
	}
}
