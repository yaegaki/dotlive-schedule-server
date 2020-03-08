package jst

import "time"

var jstLocation = time.FixedZone("Asia/Tokyo", 9*60*60)

// Time JSTのタイム
type Time struct {
	t time.Time
}

// From time.TimeからTimeを作成する
func From(t time.Time) Time {
	return Time{
		t: t.In(jstLocation),
	}
}

// Now JSTの現在時刻を取得する
func Now() Time {
	return Time{
		t: time.Now().In(jstLocation),
	}
}

// Time time.Timeを取得する
func (t Time) Time() time.Time {
	return t.t
}

// FloorToDay 時間を切り捨てる
func (t Time) FloorToDay() Time {
	return Time{
		t: time.Date(t.t.Year(), t.t.Month(), t.t.Day(), 0, 0, 0, 0, jstLocation),
	}
}

// ShortDate 日付まで指定してTimeを作成する
func ShortDate(year int, month time.Month, day int) Time {
	return Time{
		t: time.Date(year, month, day, 0, 0, 0, 0, jstLocation),
	}
}

// Date 分まで指定してTimeを作成する
func Date(year int, month time.Month, day int, hour int, min int) Time {
	return Time{
		t: time.Date(year, month, day, hour, min, 0, 0, jstLocation),
	}
}

// Year 年
func (t Time) Year() int {
	return t.t.Year()
}

// Month 月
func (t Time) Month() time.Month {
	return t.t.Month()
}

// Day 日
func (t Time) Day() int {
	return t.t.Day()
}

// Hour 時
func (t Time) Hour() int {
	return t.t.Hour()
}

// Minute 分
func (t Time) Minute() int {
	return t.t.Minute()
}

// Equal 比較する
func (t Time) Equal(other Time) bool {
	return t.t.Equal(other.t)
}

// After 指定した時刻より後か
func (t Time) After(other Time) bool {
	return t.t.After(other.t)
}

// Before 指定した時刻より前か
func (t Time) Before(other Time) bool {
	return t.t.Before(other.t)
}

// Add 加算する
func (t Time) Add(d time.Duration) Time {
	return Time{
		t: t.t.Add(d),
	}
}

// AddOneDay 1日加算する
func (t Time) AddOneDay() Time {
	return t.Add(1 * 24 * time.Hour)
}

// AddDay n日加算する
func (t Time) AddDay(n int) Time {
	return t.Add(time.Duration(n) * 24 * time.Hour)
}

func (t Time) String() string {
	return t.t.String()
}

// UnmarshalJSON .
func (t *Time) UnmarshalJSON(data []byte) error {
	err := t.t.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	t.t = t.t.In(jstLocation)
	return nil
}

// MarshalJSON .
func (t Time) MarshalJSON() ([]byte, error) {
	return t.t.MarshalJSON()
}
