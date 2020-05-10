package service

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/notify"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

// PushNotify プッシュ通知を実行する
func PushNotify(ctx context.Context, c *firestore.Client, actors model.ActorSlice) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Printf("Can not create firebase app: %v", err)
		return
	}
	msgCli, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("Can not create firebase messaging client: %v", err)
		return
	}

	pushNotifyLatestPlan(ctx, c, msgCli, actors)
	pushNotifyVideo(ctx, c, msgCli, actors)
}

func pushNotifyLatestPlan(ctx context.Context, c *firestore.Client, msgCli *messaging.Client, actors model.ActorSlice) {
	plan, err := store.FindLatestPlan(ctx, c)
	if err != nil {
		log.Printf("Can not get latest plan: %v", err)
		return
	}

	if plan.Notified {
		return
	}

	_, updated, err := store.MarkPlanAsNotified(ctx, c, plan)
	if err != nil {
		log.Printf("Can not mark plan as notified: %v", err)
		return
	}

	// 基本的にはありえないがタスクを同時実行したときなどは起こりえる
	if !updated {
		return
	}

	log.Printf("push notify plan: %v", plan.Date)
	err = notify.PushNotifyPlan(ctx, msgCli, plan, actors)
	if err != nil {
		log.Printf("Can not send push notification: %v", err)
		return
	}
}

type markVideoAsNotifiedFunc func(ctx context.Context, video model.Video) (model.Video, bool, error)

func pushNotifyVideo(ctx context.Context, c *firestore.Client, msgCli notify.Client, actors model.ActorSlice) {
	now := jst.Now()
	r := jst.Range{
		Begin: now.AddDay(-2),
		End:   now.AddOneDay(),
	}

	// 昨日、今日、明日の計画を取得する
	plans, err := store.FindPlans(ctx, c, r)
	if err != nil {
		log.Printf("Can not get plans: %v", err)
		return
	}

	videos, err := store.FindNotNotifiedVideos(ctx, c)
	if err != nil {
		log.Printf("Can not get videos: %v", err)
		return
	}

	pushNotifyVideoInternal(ctx, msgCli, plans, videos, actors, now, func(ctx context.Context, v model.Video) (model.Video, bool, error) {
		return store.MarkVideoAsNotified(ctx, c, v)
	})
}

func pushNotifyVideoInternal(ctx context.Context, msgCli notify.Client, plans []model.Plan, videos []model.Video, actors model.ActorSlice, now jst.Time, markAsNotified markVideoAsNotifiedFunc) {
	// 現在時間より2時間前の場合は古いので通知しない
	notifyLimit := now.Add(-2 * time.Hour)

	// 既にプッシュ通知を送った配信者
	// 同時に複数のプッシュ通知を送らないように制御する
	notifiedActor := map[string]bool{}

	for _, v := range videos {
		isPlanned := false
		startAt := v.StartAt
		var targetPlan model.Plan
		collaboID := 0

		for _, p := range plans {
			e, err := p.GetEntry(v)
			if err != nil {
				continue
			}

			// Entryが見つかった場合は計画配信
			isPlanned = true
			targetPlan = p

			// Bilibiliの場合は開始時刻を正しく取得できないので開始時刻に補正する
			if v.Source == model.VideoSourceBilibili {
				startAt = e.StartAt
			}

			collaboID = e.CollaboID

			break
		}

		// まだ時間になっていない
		if startAt.After(now) {
			continue
		}

		v, updated, err := markAsNotified(ctx, v)
		if err != nil {
			log.Printf("Can not mark video as notified: %v", err)
			continue
		}

		if !updated {
			continue
		}

		if _, ok := notifiedActor[v.ActorID]; ok {
			log.Printf("Skip notify video because duplicate notification. video:%v acthor:%v", v.ID, v.ActorID)
			continue
		}

		notifiedActor[v.ActorID] = true

		if startAt.Before(notifyLimit) {
			log.Printf("Skip notify video because old. video:%v startAt:%v now:%v", v.ID, startAt, now)
			continue
		}

		// シロちゃんの動画は常に計画されているとする
		const siroID = "lLhToxu1Kyxuwwygh0FK"
		if v.ActorID == siroID && !v.IsLive {
			isPlanned = true
		}

		var relatedActors []model.Actor
		if collaboID > 0 {
			for _, e := range targetPlan.Entries {
				if e.CollaboID != collaboID {
					continue
				}
				actor, err := actors.FindActor(e.ActorID)
				if err != nil {
					log.Printf("Unknown actor %v", actor.ID)
					continue
				}

				relatedActors = append(relatedActors, actor)
			}
		} else {
			actor, err := actors.FindActor(v.ActorID)
			if err != nil {
				log.Printf("Unknown actor %v", actor.ID)
				continue
			}

			relatedActors = append(relatedActors, actor)
		}

		log.Printf("push notify video: %v, %v, isPlanned:%v, isLive:%v isCollabo:%v", v.ID, v.Text, isPlanned, v.IsLive, collaboID > 0)
		var baseDate jst.Time
		if isPlanned {
			baseDate = targetPlan.Date
		} else {
			baseDate = v.StartAt
		}
		err = notify.PushNotifyVideo(ctx, msgCli, baseDate, v, relatedActors)
		if err != nil {
			log.Printf("Can not send push notification: %v", err)
			return
		}
	}
}
