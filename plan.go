// 生放送予定のツイートをパースする
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ChimeraCoder/anaconda"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func parsePlan(jstNow time.Time, tweetID string, tweet string, actors []Actor) (Plan, error) {
	lines := strings.Split(tweet, "\n")
	state := 0
	const liveScheduleStr = "生放送スケジュール"
	const liveScheduleLayout = "【生放送スケジュール1月2日】"

	// 放送時間のフォーマット
	// 24時などに対応するために時間部分にDayを使用する
	const dateLayout = "2:04~:"

	p := Plan{
		SourceID: tweetID,
	}

	for _, line := range lines {
		switch state {
		case 0:
			if !strings.Contains(line, liveScheduleStr) {
				continue
			}

			t, err := time.Parse(liveScheduleLayout, line)
			if err != nil {
				continue
			}

			p.Date = createJSTTime(jstNow.Year(), t.Month(), t.Day(), 0, 0)
			state = 1

		case 1:
			l := strings.Split(line, "#")
			if len(l) == 1 {
				continue
			}

			t, err := time.Parse(dateLayout, strings.TrimSpace(l[0]))
			if err != nil {
				continue
			}

			// 24時以上対応のためにDay部分に時間が入っている
			hour := t.Day()
			minute := t.Minute()

			startAt := createJSTTime(p.Date.Year(), p.Date.Month(), p.Date.Day(), hour, minute)

			for _, actor := range actors {
				if !strings.Contains(line, actor.Hashtag) {
					continue
				}

				p.Entries = append(p.Entries, PlanEntry{
					ActorID: actor.id,
					StartAt: startAt,
				})
			}
		}
	}

	if state == 0 {
		return Plan{}, ErrNotPlanTweet
	}

	return p, nil
}

func getPlanFromTweet(api *anaconda.TwitterApi, actors []Actor, lastTweetID string) ([]Plan, error) {
	timeline, err := getTimeline(api, "dotLIVEyoutuber", lastTweetID)

	if err != nil {
		return nil, err
	}

	jstNow := time.Now().In(jst)
	plans := []Plan{}
	for _, tweet := range timeline {
		p, err := parsePlan(jstNow, tweet.IdStr, tweet.FullText, actors)
		if err != nil {
			continue
		}

		plans = append(plans, p)
	}

	return plans, nil
}

func createDayKey(t time.Time) string {
	return fmt.Sprintf("%v-%v-%v", t.Year(), int(t.Month()), t.Day())
}

func storePlanToStore(ctx context.Context, api *anaconda.TwitterApi, client *firestore.Client, actors []Actor) error {
	planCollection := client.Collection("Plan")
	it := planCollection.OrderBy("date", firestore.Desc).Limit(1).Documents(ctx)
	latestPlanDocs, err := it.GetAll()
	if err != nil {
		return err
	}

	lastTweetID := ""
	if len(latestPlanDocs) > 0 {
		var latestPlan Plan
		latestPlanDocs[0].DataTo(&latestPlan)
		lastTweetID = latestPlan.SourceID
	}

	plans, err := getPlanFromTweet(api, actors, lastTweetID)
	if err != nil {
		return err
	}

	if len(plans) == 0 {
		return nil
	}

	err = client.RunTransaction(ctx, func(c context.Context, t *firestore.Transaction) error {
		writeFuncs := []func() error{}

		// notifiedを古いPlanから取得する
		for _, p := range plans {
			plan := p
			key := createDayKey(plan.Date)
			docRef := planCollection.Doc(key)
			doc, err := t.Get(docRef)
			if err != nil && status.Code(err) != codes.NotFound {
				return err
			}

			if err == nil {
				var oldPlan Plan
				doc.DataTo(&oldPlan)
				plan.Notified = oldPlan.Notified
			}

			writeFuncs = append(writeFuncs, func() error {
				return t.Set(docRef, plan)
			})
		}

		for _, f := range writeFuncs {
			err := f()
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func findPlan(ctx context.Context, c *firestore.Client, date time.Time) (Plan, error) {
	key := createDayKey(date.In(jst))
	doc, err := c.Collection("Plan").Doc(key).Get(ctx)
	if err != nil {
		return Plan{}, err
	}

	var p Plan
	doc.DataTo(&p)
	p.Date = p.Date.In(jst)
	return p, nil
}

func findLatestPlan(ctx context.Context, c *firestore.Client) (Plan, error) {
	it := c.Collection("Plan").OrderBy("date", firestore.Desc).Limit(1).Documents(ctx)
	latestPlanDocs, err := it.GetAll()
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return Plan{}, ErrNotFound
		}

		return Plan{}, err
	}

	var p Plan
	latestPlanDocs[0].DataTo(&p)
	p.Date = p.Date.In(jst)
	return p, nil
}

func (p Plan) getEntry(v Video) (PlanEntry, error) {
	return PlanEntry{}, nil
}

func (p Plan) update(ctx context.Context, c *firestore.Client) error {
	_, err := c.Collection("Plan").Doc(createDayKey(p.Date.In(jst))).Set(ctx, p)
	return err
}
