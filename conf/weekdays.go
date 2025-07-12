package conf

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const DAYS_IN_WEEK = 7

type WeekdaysSet struct {
	days [DAYS_IN_WEEK]bool
}

func (s *WeekdaysSet) IsEmpty() bool {
	for _, present := range s.days {
		if present {
			return false
		}
	}
	return true
}

func (s *WeekdaysSet) Contains(day time.Time) bool {
	return s.days[day.Weekday()]
}

func (s *WeekdaysSet) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node got %s", node.Value)
	}
	return s.ParseString(node.Value)
}

func (s WeekdaysSet) MarshalYAML() (interface{}, error) {
	out := ""
	for idx, present := range s.days {
		if present {
			// space between words
			if out != "" {
				out += " "
			}
			out += time.Weekday(idx).String()
		}
	}
	return out, nil
}

func (s *WeekdaysSet) ParseString(value string) error {
	days := [DAYS_IN_WEEK]bool{}
	daysStr := strings.Split(value, " ")
	for _, dayStr := range daysStr {
		dayStr = strings.ToLower(dayStr)
		switch dayStr {
		case "sunday":
			fallthrough
		case "sun":
			days[time.Sunday] = true
		case "monday":
			fallthrough
		case "mon":
			days[time.Monday] = true
		case "tuesday":
			fallthrough
		case "tue":
			days[time.Tuesday] = true
		case "wednesday":
			fallthrough
		case "wed":
			days[time.Wednesday] = true
		case "thursday":
			fallthrough
		case "thu":
			days[time.Thursday] = true
		case "friday":
			fallthrough
		case "fri":
			days[time.Friday] = true
		case "saturday":
			fallthrough
		case "sat":
			days[time.Saturday] = true
		default:
			return fmt.Errorf("unexpected value %q - expecting Mon/Tue/... or Monday/..., failed to parse %q", dayStr, value)
		}
	}
	s.days = days
	return nil
}
