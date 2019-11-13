package cmd

import (
	"errors"
	"fmt"
)

type Weekday int

const (
	UNKNOWN_WEEKDAY Weekday = iota - 1
	SUNDAY          Weekday = iota
	MONDAY          Weekday = iota
	TUESDAY         Weekday = iota
	WEDNESDAY       Weekday = iota
	THURSDAY        Weekday = iota
	FRIDAY          Weekday = iota
	SATURDAY        Weekday = iota
)

func (w Weekday) String() string {
	return [...]string{
		"UNKNOWN_WEEKDAY",
		"SUNDAY",
		"MONDAY",
		"TUESDAY",
		"WEDNESDAY",
		"THURSDAY",
		"FRIDAY",
		"SATURDAY",
	}[w]
}

func ValueOf(str string) (Weekday, error) {
	switch str {
	case "UNKNOWN_WEEKDAY":
		return UNKNOWN_WEEKDAY, nil
	case "SUNDAY":
		return SUNDAY, nil
	case "MONDAY":
		return MONDAY, nil
	case "TUESDAY":
		return TUESDAY, nil
	case "WEDNESDAY":
		return WEDNESDAY, nil
	case "THURSDAY":
		return THURSDAY, nil
	case "FRIDAY":
		return FRIDAY, nil
	case "SATURDAY":
		return SATURDAY, nil
	default:
		return UNKNOWN_WEEKDAY, errors.New(fmt.Sprintf("Unknown Weekday: '%s'", str))
	}
}
