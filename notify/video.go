package notify

import (
	"context"
	"fmt"
	"strings"

	"github.com/yaegaki/dotlive-schedule-server/model"
	"golang.org/x/xerrors"
)

const maxTopicCount = 5

// PushNotifyVideo 配信をプッシュ通知する
func PushNotifyVideo(ctx context.Context, cli Client, v model.Video, actors []model.Actor) error {
	count := len(actors)
	if count == 0 {
		return xerrors.Errorf("Actors are empty. video '%v'", v.URL)
	}

	if count == 1 {
		// ソロ
		actor := actors[0]
		condition := fmt.Sprintf("'%v' in topics", actor.TwitterScreenName)
		title := fmt.Sprintf("配信:%v", actor.Name)
		body := v.Text
		_, err := cli.Send(ctx, createMessageWithCondition(condition, title, body))

		return err
	}

	// コラボ
	emojis := []string{}
	conditions := []string{}
	for _, a := range actors {
		emojis = append(emojis, a.Emoji)
		conditions = append(conditions, fmt.Sprintf("'%v' in topics", a.TwitterScreenName))
	}

	title := fmt.Sprintf("コラボ配信:%v", strings.Join(emojis, ""))
	body := v.Text

	// 一度に指定できるトピックは5つまでなのでそれ以上の場合は分ける
	// トピックを全て購読している人には通知が二回行くが仕方ない
	// (多分5人以上のコラボはほとんどないので気にしない)
	for i := 0; i < count; i += maxTopicCount {
		end := i + maxTopicCount
		if end >= count {
			end = count
		}

		condition := strings.Join(conditions[i:end], " || ")
		_, err := cli.Send(ctx, createMessageWithCondition(condition, title, body))
		// 最初の送信でエラーになった場合はエラーにする
		// 2回目以降にエラーになった場合は1回目を取り消すことはできないので無視する
		if i == 0 && err != nil {
			return err
		}
	}

	return nil
}
