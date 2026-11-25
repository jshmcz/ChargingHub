package app

import (
	"testing"
	"time"
)

func TestNextStartTimeRollsToTomorrow(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	hhmm := "00:00"
	got := nextStartTime(hhmm)
	if !got.After(now) {
		t.Fatalf("expected future start, got %s now %s", got, now)
	}
	if got.Sub(now) > 24*time.Hour+time.Minute {
		t.Fatalf("expected at most one day ahead, got %s", got.Sub(now))
	}
}

func TestMoneyText(t *testing.T) {
	if got := moneyText(5050); got != "50.50 元" {
		t.Fatalf("unexpected money text: %s", got)
	}
}
