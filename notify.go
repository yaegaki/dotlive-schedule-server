package main

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
)

type notification struct {
}

func pushNotify(ctx context.Context, c *firestore.Client) {
	// 新着予定の通知
	notifyLatestPlan(ctx, c)
	// 配信開始した動画の通知
	notifyVideo(ctx, c)
}

func notifyLatestPlan(ctx context.Context, c *firestore.Client) {
	latestPlan, err := findLatestPlan(ctx, c)
	if err != nil {
		log.Printf("Can not get latest plan: %v", err)
		return
	}

	if latestPlan.Notified {
		return
	}

	// 通知を送ったフラグは先に立てておいて複数回通知を送らいないようにする
	latestPlan.Notified = true
	err = latestPlan.update(ctx, c)
	if err != nil {
		log.Printf("Can not update plan's flag: %v", err)
		return
	}

	// TODO:プッシュ通知を送る
	log.Printf("push notify plan: %v", latestPlan.Date.In(jst))
}

// videoResolver 動画が計画されたものかどうか、いつ開始なのかを解決する
type videoResolver struct {
	ctx     context.Context
	c       *firestore.Client
	planMap map[string]Plan
}

type videoResolveResult struct {
}

func (r *videoResolver) getPlan(t time.Time) (Plan, error) {
	key := createDayKey(t)
	if s, ok := r.planMap[key]; ok {
		return s, nil
	}

	p, err := findPlan(r.ctx, r.c, t)
	if err != nil {
		return Plan{}, err
	}
	r.planMap[key] = p
	return p, nil
}

func (r *videoResolver) resolveByTime(v Video, t time.Time) (videoResolveResult, error) {
	p, err := r.getPlan(t)
	if err != nil {
		return videoResolveResult{}, err
	}

	e, err := p.getEntry(v)
	if err != nil {
		return videoResolveResult{}, err
	}

	log.Print(e.ActorID)
	return videoResolveResult{}, nil
}

func (r *videoResolver) resolve(v Video) (videoResolveResult, error) {
	today := getDay(v.StartAt)

	// 5時前なら前日の計画も参照する
	if v.StartAt.Before(today.Add(5 * time.Hour)) {
		yesterday := getDay(v.StartAt.Add(-1 * 24 * time.Hour))
		res, err := r.resolveByTime(v, yesterday)
		if err == nil {
			return res, nil
		} else if err != ErrNotFound {
			return videoResolveResult{}, err
		}
	}

	res, err := r.resolveByTime(v, today)
	if err == nil {
		return res, nil
	} else if err != ErrNotFound {
		return videoResolveResult{}, err
	}

	// 22時降なら明日の計画も参照する
	if v.StartAt.After(today.Add(22 * time.Hour)) {
		tommorow := getDay(v.StartAt.Add(1 * 24 * time.Hour))
		res, err := r.resolveByTime(v, tommorow)
		if err == nil {
			return res, nil
		} else if err != ErrNotFound {
			return videoResolveResult{}, err
		}
	}

	// ゲリラ配信
	return videoResolveResult{}, nil
}

func notifyVideo(ctx context.Context, c *firestore.Client) {
}
