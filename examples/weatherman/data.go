package main

import (
	"math/rand"
	"time"
)

type weatherTuple struct {
	Day     int
	Month   string
	Hour    int
	Degrees int
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func getMonth(m int) string {
	switch m {
	case 1:
		return "jan"
	case 2:
		return "feb"
	case 3:
		return "mar"
	case 4:
		return "apr"
	case 5:
		return "may"
	case 6:
		return "jun"
	case 7:
		return "jul"
	case 8:
		return "aug"
	case 9:
		return "sep"
	case 10:
		return "oct"
	case 11:
		return "nov"
	case 12:
		return "dec"
	default:
		return "unk"
	}
}

func weatherData(city string) []weatherTuple {
	rand.Seed(time.Now().Unix())

	data := []weatherTuple{}

	for m := 1; m <= 12; m++ {
		for day := 1; day <= 31; day++ {
			if m == 2 && day > 28 {
				continue
			}

			if m == 4 || m == 6 || m == 9 || m == 11 {
				continue
			}

			for hour := 1; hour <= 24; hour++ {
				data = append(data, weatherTuple{
					day, getMonth(m), hour, random(50, 95),
				})
			}
		}
	}

	return data
}
