package main

import (
	"testing"
)

var iori = Actor{
	id:      "iori",
	Name:    "ヤマト イオリ",
	Hashtag: "#ヤマトイオリ",
}

var pino = Actor{
	id:      "pino",
	Name:    "カルロ・ピノ",
	Hashtag: "#カルロピノ",
}

var suzu = Actor{
	id:      "suzu",
	Name:    "神楽すず",
	Hashtag: "#神楽すず",
}

var chieri = Actor{
	id:      "chieri",
	Name:    "花京院ちえり",
	Hashtag: "#花京院ちえり",
}

var iroha = Actor{
	id:      "iroha",
	Name:    "金剛いろは",
	Hashtag: "#金剛いろは",
}

var futaba = Actor{
	id:      "futaba",
	Name:    "北上双葉",
	Hashtag: "#北上双葉",
}

var mememe = Actor{
	id:      "mememe",
	Name:    "もこ田めめめ",
	Hashtag: "#もこ田めめめ",
}

var siro = Actor{
	id:      "siro",
	Name:    "電脳少女シロ",
	Hashtag: "#シロ生放送",
}

var milk = Actor{
	id:      "milk",
	Name:    "メリーミルク",
	Hashtag: "#メリーミルク",
}

var actors = []Actor{
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

const year = 2020

func testCreatePlan(tweet string, expect Plan, t *testing.T) {
	jstNow := createJSTTime(year, 1, 1, 0, 0)

	p, e := parsePlan(jstNow, "xxx", tweet, actors)
	if e != nil {
		t.Errorf("parse failed")
		return
	}

	if p.Date != expect.Date {
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

func TestCreatePlan1(t *testing.T) {
	tweet := `【どっとライブ】【アイドル部】
【生放送スケジュール2月26日】

19:00~: #ヤマトイオリ
21:00~: #カルロピノ
22:00~: #神楽すず
23:00~: #メリーミルク

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	plan := Plan{
		Date: createJSTTime(year, 2, 26, 0, 0),
		Entries: []PlanEntry{
			PlanEntry{
				ActorID: iori.id,
				StartAt: createJSTTime(year, 2, 26, 19, 00),
			},
			PlanEntry{
				ActorID: pino.id,
				StartAt: createJSTTime(year, 2, 26, 21, 00),
			},
			PlanEntry{
				ActorID: suzu.id,
				StartAt: createJSTTime(year, 2, 26, 22, 00),
			},
			PlanEntry{
				ActorID: milk.id,
				StartAt: createJSTTime(year, 2, 26, 23, 00),
			},
		},
	}

	testCreatePlan(tweet, plan, t)
}

func TestCreatePlan2(t *testing.T) {
	tweet := `【どっとライブ】【アイドル部】
【生放送スケジュール2月27日】

19:00~: #花京院ちえり
20:00~: #シロ生放送

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	plan := Plan{
		Date: createJSTTime(year, 2, 27, 0, 0),
		Entries: []PlanEntry{
			PlanEntry{
				ActorID: chieri.id,
				StartAt: createJSTTime(year, 2, 27, 19, 00),
			},
			PlanEntry{
				ActorID: siro.id,
				StartAt: createJSTTime(year, 2, 27, 20, 00),
			},
		},
	}

	testCreatePlan(tweet, plan, t)
}

func TestCreatePlan3(t *testing.T) {
	tweet := `【どっとライブ】【アイドル部】
【生放送スケジュール2月28日】

20:00~: #シロ生放送 (bilibili)
22:00~: #北上双葉
23:00~: #神楽すず

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	plan := Plan{
		Date: createJSTTime(year, 2, 28, 0, 0),
		Entries: []PlanEntry{
			PlanEntry{
				ActorID: siro.id,
				StartAt: createJSTTime(year, 2, 28, 20, 00),
			},
			PlanEntry{
				ActorID: futaba.id,
				StartAt: createJSTTime(year, 2, 28, 22, 00),
			},
			PlanEntry{
				ActorID: suzu.id,
				StartAt: createJSTTime(year, 2, 28, 23, 00),
			},
		},
	}

	testCreatePlan(tweet, plan, t)
}

func TestCreatePlan4(t *testing.T) {
	tweet := `訂正のお知らせ
【生放送スケジュール2月25日】

19:30~: #神楽すず (bilibili)
21:00~: #花京院ちえり

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	plan := Plan{
		Date: createJSTTime(year, 2, 25, 0, 0),
		Entries: []PlanEntry{
			PlanEntry{
				ActorID: suzu.id,
				StartAt: createJSTTime(year, 2, 25, 19, 30),
			},
			PlanEntry{
				ActorID: chieri.id,
				StartAt: createJSTTime(year, 2, 25, 21, 00),
			},
		},
	}

	testCreatePlan(tweet, plan, t)
}

func TestCreatePlan5(t *testing.T) {
	tweet := `【生放送スケジュール2月22日】

12:00~: #神楽すず
15:00~: #カルロピノ
18:00~: #北上双葉
19:00~: #花京院ちえり
20:00~: #Vに国境はいらない（出演時間21:30~を予定）
24:00~: #カルロピノ

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	plan := Plan{
		Date: createJSTTime(year, 2, 22, 0, 0),
		Entries: []PlanEntry{
			PlanEntry{
				ActorID: suzu.id,
				StartAt: createJSTTime(year, 2, 22, 12, 00),
			},
			PlanEntry{
				ActorID: pino.id,
				StartAt: createJSTTime(year, 2, 22, 15, 00),
			},
			PlanEntry{
				ActorID: futaba.id,
				StartAt: createJSTTime(year, 2, 22, 18, 00),
			},
			PlanEntry{
				ActorID: chieri.id,
				StartAt: createJSTTime(year, 2, 22, 19, 00),
			},
			PlanEntry{
				ActorID: pino.id,
				StartAt: createJSTTime(year, 2, 22, 24, 00),
			},
		},
	}

	testCreatePlan(tweet, plan, t)
}

func TestCreatePlan6(t *testing.T) {
	tweet := `【どっとライブ】【アイドル部】
【生放送スケジュール2月6日】

20:00~: #シロ生放送
21:00~: #花京院ちえり × #神楽すず

メンバーの動画、SNSのリンクはこちらから！
http://vrlive.party/member/

#アイドル部　#どっとライブ`

	plan := Plan{
		Date: createJSTTime(year, 2, 6, 0, 0),
		Entries: []PlanEntry{
			PlanEntry{
				ActorID: siro.id,
				StartAt: createJSTTime(year, 2, 6, 20, 00),
			},
			PlanEntry{
				ActorID: suzu.id,
				StartAt: createJSTTime(year, 2, 6, 21, 00),
			},
			PlanEntry{
				ActorID: chieri.id,
				StartAt: createJSTTime(year, 2, 6, 21, 00),
			},
		},
	}

	testCreatePlan(tweet, plan, t)
}
