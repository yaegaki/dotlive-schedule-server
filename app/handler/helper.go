package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/yaegaki/dotlive-schedule-server/jst"
	"golang.org/x/xerrors"
)

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
