package tweet

import (
	"errors"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"golang.org/x/xerrors"
)

// planParser TweetからPlanを作成するParser
type planParser struct {
	actors []model.Actor
}

const liveScheduleStr = "生放送スケジュール"
const liveScheduleLayout = "【生放送スケジュール1月2日】"

// parse TweetからPlanを作成する
func (pp planParser) parse(t Tweet) (model.Plan, error) {
	lines := strings.Split(t.Text, "\n")
	state := 0

	// 放送時間のフォーマット
	// 24時などに対応するために時間部分にDayを使用する
	const dateLayout = "2:04~:"

	p := model.Plan{
		SourceID: t.ID,
	}

	tweetDate := t.Date

	collaboID := 1

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

			var year int
			if t.Month() == 1 && tweetDate.Month() != 1 {
				year = tweetDate.Year() + 1
			} else {
				year = tweetDate.Year()
			}

			p.Date = jst.ShortDate(year, t.Month(), t.Day())
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

			startAt := jst.Date(p.Date.Year(), p.Date.Month(), p.Date.Day(), hour, minute)

			actorCount := 0
			prevEntryCount := len(p.Entries)

			for _, actor := range pp.actors {
				if !strings.Contains(line, actor.Hashtag) {
					continue
				}

				var source string
				if strings.Contains(strings.ToLower(line), "bilibili") {
					source = model.VideoSourceBilibili
				} else {
					source = model.VideoSourceYoutube
				}

				p.Entries = append(p.Entries, model.PlanEntry{
					ActorID: actor.ID,
					StartAt: startAt,
					Source:  source,
				})

				actorCount++
			}

			// コラボ
			if actorCount > 1 {
				for i := range p.Entries {
					if i < prevEntryCount {
						continue
					}

					e := p.Entries[i]
					e.CollaboID = collaboID
					p.Entries[i] = e
				}

				collaboID++
			}
		}
	}

	if state == 0 {
		return model.Plan{}, errors.New("this tweet is not a plan")
	}

	return p, nil
}

// FindPlans どっとライブのアカウントからPlanを取得する
func FindPlans(api *anaconda.TwitterApi, lastTweetID string, actors []model.Actor) ([]model.Plan, error) {
	timeline, err := getTimeline(api, ScreenNameDotlive, lastTweetID)
	if err != nil {
		return nil, xerrors.Errorf("Can not get timeline: %w", err)
	}

	pp := planParser{
		actors: actors,
	}

	plans := []model.Plan{}
	for _, t := range timeline {
		p, err := pp.parse(t)
		if err != nil {
			continue
		}

		plans = append(plans, p)
	}

	return plans, nil
}
