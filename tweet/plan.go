package tweet

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"golang.org/x/xerrors"
)

const liveScheduleStr = "生放送スケジュール"
const liveScheduleLayout = "【生放送スケジュール1月2日】"

// ParsePlanTweet TweetからPlanを作成する
func ParsePlanTweet(t Tweet, actors model.ActorSlice, strict bool) (model.Plan, error) {
	lines := strings.Split(t.Text, "\n")
	state := 0

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

			// 計画が複数に分けてツイートされる場合、
			// 日付の後に①などがついている
			// パースするときに邪魔なので①の部分は消す
			daySplits := strings.Split(line, "日")
			if len(daySplits) > 1 {
				// 複数に分かれている場合は常に追記
				// TODO: 訂正かつ複数に分かれている場合、
				// ずっと追記になってしまうのでどうにかする
				if !strings.HasPrefix(daySplits[1], "】") {
					p.Additional = true
				}

				line = daySplits[0] + "日】"
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
			// 過去の計画がツイートされるのはおかしいのでその場合は翌日の計画とする
			if !tweetDate.FloorToDay().Equal(p.Date) && p.Date.Before(tweetDate) {
				p.Date = tweetDate.AddOneDay().FloorToDay()
			}
			state = 1

		case 1:
			l := strings.Split(line, "#")
			if len(l) == 1 {
				continue
			}

			timeStr := strings.TrimSpace(strings.Split(l[0], "~:")[0])
			startAt, err := parseEntryTime(p.Date, timeStr)
			if err != nil {
				continue
			}

			actorCount := 0
			prevEntryCount := len(p.Entries)

			for _, actor := range actors {
				if !strings.Contains(line, actor.Hashtag) {
					continue
				}

				var source string
				memberOnly := false
				if strings.Contains(strings.ToLower(line), "bilibili") {
					source = model.VideoSourceBilibili
				} else if strings.Contains(strings.ToLower(line), "mildom") {
					source = model.VideoSourceMildom
				} else {
					source = model.VideoSourceYoutube
					memberOnly = strings.Contains(line, "メンバーシップ限定")
				}

				p.Entries = append(p.Entries, model.PlanEntry{
					ActorID:    actor.ID,
					StartAt:    startAt,
					Source:     source,
					MemberOnly: memberOnly,
				})

				actorCount++
			}

			// strictの場合は知らないハッシュタグがあるとエラー扱い
			if strict && actorCount != (len(l)-1) {
				return model.Plan{}, xerrors.Errorf("invalid line: %v", line)
			}

			if actorCount == 0 {
				hashTagIndex := strings.Index(line, "#")
				if hashTagIndex < 0 {
					panic("hashTagIndex")
				}
				hashTag := string([]rune(line)[hashTagIndex:])
				// コラボやイベントなどの特殊なハッシュタグ
				p.Entries = append(p.Entries, model.PlanEntry{
					ActorID: model.ActorIDUnknown,
					HashTag: hashTag,
					StartAt: startAt,
					// とりあえずYoutubeにしておく
					// (コラボはYoutubeではない可能性があるが
					//	そもそも配信ページのリンクを取得できないので
					//	Youtubeにしておいても問題ないはず)
					Source: model.VideoSourceYoutube,
				})
			} else if actorCount > 1 {
				// コラボ
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

			if p.Text == "" {
				p.Text = line
			} else {
				p.Text = p.Text + "\n" + line
			}
		}
	}

	if state == 0 {
		return model.Plan{}, errors.New("this tweet is not a plan")
	}

	return p, nil
}

var errInvalidTimeFormat = errors.New("invalid time format")

func parseEntryTime(base jst.Time, timeStr string) (jst.Time, error) {
	xs := strings.Split(timeStr, ":")
	if len(xs) != 2 {
		return jst.Time{}, errInvalidTimeFormat
	}

	hour, err := strconv.Atoi(xs[0])
	if err != nil || hour < 0 || hour > 47 {
		return jst.Time{}, errInvalidTimeFormat
	}

	minute, err := strconv.Atoi(xs[1])
	if err != nil || minute < 0 || minute > 59 {
		return jst.Time{}, errInvalidTimeFormat
	}

	return jst.Date(base.Year(), base.Month(), base.Day(), hour, minute), nil
}

// FindPlans どっとライブのアカウントからPlanを取得する
func FindPlans(api *anaconda.TwitterApi, user model.TwitterUser, actors []model.Actor) (model.TwitterUser, []model.Plan, error) {
	timeline, err := getTimeline(api, user.ScreenName, user.LastTweetID)
	if err != nil {
		return model.TwitterUser{}, nil, xerrors.Errorf("Can not get timeline: %w", err)
	}

	plans := []model.Plan{}
	for i, t := range timeline {
		if i == 0 {
			user.LastTweetID = t.ID
		}
		p, err := ParsePlanTweet(t, actors, false)
		if err != nil {
			continue
		}

		plans = append(plans, p)
	}

	return user, plans, nil
}
