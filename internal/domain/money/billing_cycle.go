package money

import (
	"fmt"
	"time"
)

type BillingCycle struct {
	startDay int
}

func NewBillingCycle(startDay int) (BillingCycle, error) {
	if startDay < 1 || startDay > 31 {
		return BillingCycle{}, fmt.Errorf("billing cycle day must be between 1 and 31")
	}
	return BillingCycle{startDay: startDay}, nil
}

func DefaultBillingCycle() BillingCycle {
	return BillingCycle{startDay: 1}
}

func (b BillingCycle) StartDay() int { return b.startDay }

func (b BillingCycle) Period(now time.Time) (time.Time, time.Time) {
	start := cycleDate(now.Year(), now.Month(), b.startDay, now.Location())
	if now.Before(start) {
		start = cycleDate(now.Year(), now.Month()-1, b.startDay, now.Location())
	}
	end := cycleDate(start.Year(), start.Month()+1, b.startDay, now.Location())
	return start, end
}

func cycleDate(year int, month time.Month, day int, location *time.Location) time.Time {
	first := time.Date(year, month, 1, 0, 0, 0, 0, location)
	lastDay := first.AddDate(0, 1, -1).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(first.Year(), first.Month(), day, 0, 0, 0, 0, location)
}
