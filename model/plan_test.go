package model

import (
	"testing"
	"time"

	"github.com/yaegaki/dotlive-schedule-server/jst"
)

func TestIsPlanned(t *testing.T) {
	p := CreatePlan(jst.ShortDate(2020, 6, 14), []EntryPart{
		CreateEntryPart(Futaba, 20, 0),
		CreateEntryPart(Suzu, 22, 0),
		CreateEntryPart(Chieri, 22, 0),
		CreateEntryPartMildom(Mememe, 24, 0),
	})

	videos := []Video{
		{
			ID:      "chieri",
			ActorID: Chieri.ID,
			Source:  VideoSourceYoutube,
			StartAt: jst.Date(2020, 6, 14, 22, 0),
		},
		{
			ID:      "mememe",
			ActorID: Mememe.ID,
			Source:  VideoSourceMildom,
			StartAt: jst.Date(2020, 6, 14, 24, 0),
		},
	}

	for _, v := range videos {
		if !p.IsPlanned(v) {
			t.Errorf("%v is planned. (%v)", v.ID, v.StartAt)
		}
	}

	notPlanned := []Video{
		{
			ID:      "suzu",
			ActorID: Suzu.ID,
			Source:  VideoSourceYoutube,
			StartAt: jst.Date(2020, 6, 14, 4, 0),
		},
		{
			ID:      "chieri",
			ActorID: Chieri.ID,
			Source:  VideoSourceYoutube,
			StartAt: jst.Date(2020, 6, 16, 0, 0),
		},
		{
			ID:      "mememe",
			ActorID: Mememe.ID,
			Source:  VideoSourceMildom,
			StartAt: jst.Date(2020, 6, 16, 3, 0),
		},
	}
	for _, v := range notPlanned {
		if p.IsPlanned(v) {
			t.Errorf("%v is not planned. (%v)", v.ID, v.StartAt)
		}
	}
}

func TestGetEntryIndex(t *testing.T) {
	p := CreatePlan(jst.ShortDate(2020, 6, 15), []EntryPart{
		CreateEntryPart(Suzu, 13, 0),
		CreateEntryPartMildom(Suzu, 19, 0),
	})

	index := p.GetEntryIndex(Video{
		ID:      "mildom",
		ActorID: Suzu.ID,
		Source:  VideoSourceMildom,
		StartAt: jst.Date(2020, 6, 15, 19, 0),
	})
	if index < 0 {
		t.Fatal("not found entry for mildom")
	}

	if p.Entries[index].Source != VideoSourceMildom {
		t.Fatal("invalid video source")
	}

	p = CreatePlan(jst.ShortDate(2020, 6, 17), []EntryPart{
		CreateEntryPartBilibili(Siro, 12, 0),
	})
	index = p.GetEntryIndex(Video{
		ID:      "bilibili",
		ActorID: Siro.ID,
		Source:  VideoSourceBilibili,
		StartAt: jst.Date(2020, 6, 16, 20, 0),
	})
	if index < 0 {
		t.Fatal("not found entry for bilibili")
	}

	if p.Entries[index].Source != VideoSourceBilibili {
		t.Fatal("invalid video source")
	}
}

func TestGetEntryIndexUnknownActor(t *testing.T) {
	d := jst.ShortDate(2020, 9, 24)
	p := Plan{
		Date: d,
		Entries: []PlanEntry{
			{
				ActorID: Natori.ID,
				Source:  VideoSourceYoutube,
				StartAt: d.Add(10 * time.Hour),
			},
			{
				ActorID: Suzu.ID,
				Source:  VideoSourceMildom,
				StartAt: d.Add(20 * time.Hour),
			},
			{
				ActorID: ActorIDUnknown,
				Source:  VideoSourceYoutube,
				StartAt: d.Add(20 * time.Hour),
				HashTag: "#電脳少女ガッチマンV (Siro Channel)",
			},
			{
				ActorID: ActorIDUnknown,
				Source:  VideoSourceYoutube,
				StartAt: d.Add(21 * time.Hour),
				HashTag: "#電脳少女ガッチマンV (ガッチマンVさんチャンネル)",
			},
			{
				ActorID: Iori.ID,
				Source:  VideoSourceYoutube,
				StartAt: d.Add(22 * time.Hour),
			},
			{
				ActorID: ActorIDUnknown,
				Source:  VideoSourceYoutube,
				StartAt: d.Add(23 * time.Hour),
				HashTag: "#Vのから騒ぎ",
			},
		},
	}

	test := func(got, expect int) {
		if got == expect {
			return
		}
		t.Fatalf("Can not get entry, got: %v expect: %v", got, expect)
	}

	index := p.GetEntryIndex(Video{
		ActorID: Natori.ID,
		Source:  VideoSourceYoutube,
		StartAt: d.Add(10 * time.Hour),
	})
	test(index, 0)

	index = p.GetEntryIndex(Video{
		ActorID: Suzu.ID,
		Source:  VideoSourceMildom,
		StartAt: d.Add(20 * time.Hour),
	})
	test(index, 1)

	index = p.GetEntryIndex(Video{
		ActorID:  Siro.ID,
		Source:   VideoSourceYoutube,
		StartAt:  d.Add(20 * time.Hour),
		HashTags: []string{"電脳少女ガッチマンV"},
	})
	test(index, 2)

	index = p.GetEntryIndex(Video{
		ActorID:        ActorIDUnknown,
		Source:         VideoSourceYoutube,
		StartAt:        d.Add(21 * time.Hour),
		RelatedActorID: Siro.ID,
		HashTags:       []string{"電脳少女ガッチマンV"},
	})
	test(index, 3)

	index = p.GetEntryIndex(Video{
		ActorID: Iori.ID,
		Source:  VideoSourceYoutube,
		StartAt: d.Add(22 * time.Hour),
	})
	test(index, 4)

	index = p.GetEntryIndex(Video{
		ActorID:        ActorIDUnknown,
		Source:         VideoSourceYoutube,
		StartAt:        d.Add(23 * time.Hour),
		RelatedActorID: Pino.ID,
		HashTags:       []string{"Vのから騒ぎ"},
	})
	test(index, 5)
}

// TODO: テスト用パッケージを使う(import cycleになってエラーになるためそのままは使用できない)

// EntryPart .
type EntryPart struct {
	Actor     Actor
	Hour      int
	Min       int
	Source    string
	CollaboID int
}

// CreatePlan
func CreatePlan(d jst.Time, parts []EntryPart) Plan {
	var entries []PlanEntry
	for _, p := range parts {
		entries = append(entries, PlanEntry{
			ActorID:   p.Actor.ID,
			StartAt:   jst.Date(d.Year(), d.Month(), d.Day(), p.Hour, p.Min),
			Source:    p.Source,
			CollaboID: p.CollaboID,
		})
	}

	return Plan{
		Date:    d,
		Entries: entries,
	}
}

// CreateEntryPart .
func CreateEntryPart(actor Actor, hour, min int) EntryPart {
	return EntryPart{
		Actor:     actor,
		Hour:      hour,
		Min:       min,
		Source:    VideoSourceYoutube,
		CollaboID: 0,
	}
}

// CreateEntryPartBilibili
func CreateEntryPartBilibili(actor Actor, hour, min int) EntryPart {
	return EntryPart{
		Actor:     actor,
		Hour:      hour,
		Min:       min,
		Source:    VideoSourceBilibili,
		CollaboID: 0,
	}
}

// CreateEntryPartMildom .
func CreateEntryPartMildom(actor Actor, hour, min int) EntryPart {
	return EntryPart{
		Actor:     actor,
		Hour:      hour,
		Min:       min,
		Source:    VideoSourceMildom,
		CollaboID: 0,
	}
}

// Iori .
var Iori = Actor{
	ID:                "iori",
	Hashtag:           "#ヤマトイオリ",
	Name:              "ヤマトイオリ",
	TwitterScreenName: "test-iori",
	Emoji:             "🍄",
}

// Pino .
var Pino = Actor{
	ID:                "pino",
	Hashtag:           "#カルロピノ",
	Name:              "カルロピノ",
	TwitterScreenName: "test-pino",
	Emoji:             "🐜",
}

// Suzu .
var Suzu = Actor{
	ID:                "suzu",
	Hashtag:           "#神楽すず",
	Name:              "神楽すず",
	TwitterScreenName: "test-suzu",
	Emoji:             "🍋",
}

// Chieri .
var Chieri = Actor{
	ID:                "chieri",
	Hashtag:           "#花京院ちえり",
	Name:              "花京院ちえり",
	TwitterScreenName: "test-chieri",
	Emoji:             "🍒",
}

// Iroha .
var Iroha = Actor{
	ID:                "iroha",
	Hashtag:           "#金剛いろは",
	Name:              "金剛いろは",
	TwitterScreenName: "test-iroha",
	Emoji:             "💎",
}

// Futaba .
var Futaba = Actor{
	ID:                "futaba",
	Hashtag:           "#北上双葉",
	Name:              "北上双葉",
	TwitterScreenName: "test-futaba",
	Emoji:             "🌱",
}

// Natori .
var Natori = Actor{
	ID:                "natori",
	Hashtag:           "#八重沢なとり",
	Name:              "八重沢なとり",
	TwitterScreenName: "test-natori",
	Emoji:             "🌾",
}

// Mememe .
var Mememe = Actor{
	ID:                "mememe",
	Hashtag:           "#もこ田めめめ",
	Name:              "もこ田めめめ",
	TwitterScreenName: "test-mememe",
	Emoji:             "🐏",
}

// Siro .
var Siro = Actor{
	ID:                "siro",
	Hashtag:           "#シロ生放送",
	Name:              "電脳少女シロ",
	TwitterScreenName: "test-siro",
	Emoji:             "🐬",
}

// Milk .
var Milk = Actor{
	ID:                "milk",
	Hashtag:           "#メリーミルク",
	Name:              "メリーミルク",
	TwitterScreenName: "test-milk",
	Emoji:             "🐑",
}

// All .
var All = []Actor{
	Iori,
	Pino,
	Suzu,
	Chieri,
	Iroha,
	Futaba,
	Natori,
	Mememe,
	Siro,
	Milk,
}
