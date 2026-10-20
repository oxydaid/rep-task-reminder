package reminder

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// GoWeekdayToISO converts Go's time.Weekday (0=Sun..6=Sat) to ISO (1=Mon..7=Sun)
func GoWeekdayToISO(w time.Weekday) int {
	if w == time.Sunday {
		return 7
	}
	return int(w)
}

// ParseRecurrenceDays parses a string like "1,3,5" or "senin,rabu,jumat" into sorted 1..7 numbers
func ParseRecurrenceDays(input string) []int {
	dayMap := map[string]int{
		"senin": 1, "sen": 1, "mon": 1, "monday": 1,
		"selasa": 2, "sel": 2, "tue": 2, "tuesday": 2,
		"rabu": 3, "rab": 3, "wed": 3, "wednesday": 3,
		"kamis": 4, "kam": 4, "thu": 4, "thursday": 4,
		"jumat": 5, "jum": 5, "fri": 5, "friday": 5,
		"sabtu": 6, "sab": 6, "sat": 6, "saturday": 6,
		"minggu": 7, "min": 7, "sun": 7, "sunday": 7,
	}

	parts := strings.Split(input, ",")
	seen := make(map[int]bool)
	var days []int

	for _, p := range parts {
		clean := strings.ToLower(strings.TrimSpace(p))
		if clean == "" {
			continue
		}
		if d, ok := dayMap[clean]; ok {
			if !seen[d] {
				seen[d] = true
				days = append(days, d)
			}
			continue
		}
		if num, err := strconv.Atoi(clean); err == nil && num >= 1 && num <= 7 {
			if !seen[num] {
				seen[num] = true
				days = append(days, num)
			}
		}
	}

	sort.Ints(days)
	return days
}

// FormatRecurrenceDays formats a slice of 1..7 ints into comma separated string
func FormatRecurrenceDays(days []int) string {
	strs := make([]string, len(days))
	for i, d := range days {
		strs[i] = strconv.Itoa(d)
	}
	return strings.Join(strs, ",")
}

// CalculateNextRun returns the next scheduled time after currentScheduled
func CalculateNextRun(currentScheduled time.Time, recurrenceType RecurrenceType, recurrenceDays string, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	inLoc := currentScheduled.In(loc)

	switch recurrenceType {
	case RecurrenceOnce:
		return time.Time{}

	case RecurrenceDaily:
		return inLoc.AddDate(0, 0, 1)

	case RecurrenceWeekly:
		return inLoc.AddDate(0, 0, 7)

	case RecurrenceSpecificDays:
		days := ParseRecurrenceDays(recurrenceDays)
		if len(days) == 0 {
			// Fallback to daily if no days specified
			return inLoc.AddDate(0, 0, 1)
		}

		// Look ahead from 1 to 7 days
		for offset := 1; offset <= 7; offset++ {
			candidate := inLoc.AddDate(0, 0, offset)
			isoDay := GoWeekdayToISO(candidate.Weekday())
			for _, targetDay := range days {
				if isoDay == targetDay {
					return candidate
				}
			}
		}
		return inLoc.AddDate(0, 0, 1)

	default:
		return time.Time{}
	}
}

// CalculateInitialRun calculates the first run time for a given time of day
func CalculateInitialRun(hour, minute int, recurrenceType RecurrenceType, recurrenceDays string, loc *time.Location, now time.Time) time.Time {
	if loc == nil {
		loc = time.Local
	}
	nowInLoc := now.In(loc)

	// Today at target hour:minute
	todayTarget := time.Date(
		nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(),
		hour, minute, 0, 0, loc,
	)

	switch recurrenceType {
	case RecurrenceOnce:
		if todayTarget.After(nowInLoc) {
			return todayTarget
		}
		// If today's time passed, set for tomorrow
		return todayTarget.AddDate(0, 0, 1)

	case RecurrenceDaily:
		if todayTarget.After(nowInLoc) {
			return todayTarget
		}
		return todayTarget.AddDate(0, 0, 1)

	case RecurrenceWeekly:
		if todayTarget.After(nowInLoc) {
			return todayTarget
		}
		return todayTarget.AddDate(0, 0, 7)

	case RecurrenceSpecificDays:
		days := ParseRecurrenceDays(recurrenceDays)
		if len(days) == 0 {
			if todayTarget.After(nowInLoc) {
				return todayTarget
			}
			return todayTarget.AddDate(0, 0, 1)
		}

		todayISO := GoWeekdayToISO(nowInLoc.Weekday())
		// Check if today matches and is in future
		for _, d := range days {
			if d == todayISO && todayTarget.After(nowInLoc) {
				return todayTarget
			}
		}

		// Otherwise check next 7 days
		for offset := 1; offset <= 7; offset++ {
			candidate := todayTarget.AddDate(0, 0, offset)
			isoDay := GoWeekdayToISO(candidate.Weekday())
			for _, d := range days {
				if isoDay == d {
					return candidate
				}
			}
		}
		return todayTarget.AddDate(0, 0, 1)

	default:
		return todayTarget
	}
}
