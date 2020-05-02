package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
	"github.com/yaegaki/dotlive-schedule-server/tweet"
	"golang.org/x/net/context"
	"golang.org/x/xerrors"
)

func main() {
	// 計画をjsonファイルから登録する

	args := os.Args
	if len(args) != 2 {
		log.Fatal("usage: insertplan path/to/json")
	}

	p := args[1]
	// p := "plans/2020GW.json"
	bytes, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatalf("Can not open %v", p)
	}

	var plans []plan
	err = json.Unmarshal(bytes, &plans)
	if err != nil {
		log.Fatalf("Can not parse json")
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "dotlive-schedule")
	if err != nil {
		log.Fatalf("Can not create a firestore client: %v", err)
	}
	defer client.Close()

	actors, err := store.FindActors(ctx, client)
	if err != nil {
		log.Fatalf("Can not get actors: %v", err)
	}

	for _, temp := range plans {
		p, err := temp.Plan(actors)
		if err != nil {
			log.Fatalf("Can not create a plan: %v", err)
		}

		// このツールで追加する場合は最初から通知済みとする
		p.Notified = true

		err = store.SavePlanWithExplicitID(ctx, client, p, p.SourceID)
		if err != nil {
			log.Fatalf("Can not save plan: %v err: %v", p.SourceID, err)
		}
		log.Printf("Insert plan: %v", p.SourceID)
	}
}

type plan struct {
	Date    string   `json:"date"`
	Entries []string `json:"entries"`
}

func (p plan) Plan(actors model.ActorSlice) (model.Plan, error) {
	d, err := parseYearMonthDayQuery(p.Date)
	if err != nil {
		return model.Plan{}, err
	}

	text := fmt.Sprintf("【生放送スケジュール%v月%v日】\n", int(d.Month()), d.Day()) + strings.Join(p.Entries, "\n")

	return tweet.ParsePlanTweet(tweet.Tweet{
		ID:   fmt.Sprintf("insertplan-%v", p.Date),
		Text: text,
		Date: d.AddDay(-1),
	}, actors, true)
}

// parseYearMonthDayQuery '2022-2-22'形式の文字列をパースする
func parseYearMonthDayQuery(s string) (jst.Time, error) {
	if s != "" {
		xs := strings.Split(s, "-")
		if len(xs) == 3 {
			year, err1 := strconv.Atoi(xs[0])
			month, err2 := strconv.Atoi(xs[1])
			day, err3 := strconv.Atoi(xs[2])
			if err1 == nil && err2 == nil && err3 == nil {
				return jst.ShortDate(year, time.Month(month), day), nil
			}
		}
	}

	return jst.Time{}, xerrors.Errorf("Can not parse: %v", s)
}
