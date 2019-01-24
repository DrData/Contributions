package main

import "time"

// Utilities

// DaysBetween - (b - a) compute the number of days between the two dates
func DaysBetween(a, b time.Time) (daysdiff int) {
	daysdiff = int(b.Sub(a).Hours() / 24)
	return
}

func checkerror(e error) {
	if e != nil {
		panic(e)
	}
}
