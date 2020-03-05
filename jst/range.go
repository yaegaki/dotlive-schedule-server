package jst

// Range 時間の範囲
type Range struct {
	// Begin 始まり
	Begin Time
	// End 終わり
	End Time
}

// In 時間が範囲内かどうか
func (r Range) In(t Time) bool {
	if r.Begin.Equal(t) || r.End.Equal(t) {
		return true
	}

	return r.Begin.Before(t) && r.End.After(t)
}
