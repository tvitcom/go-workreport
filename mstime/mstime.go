// https://stackoverflow.com/questions/17140652/read-time-from-excel-sheet-using-xlrd-in-time-format-and-not-in-float
package mstime

//package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	PRODMODE bool = false
)

// The function return hours, minutes, seconds
func GetStringTime(str string, hour24 bool, ns bool) (s string, e error) {
	var hours, minutes, seconds float64
	var ost float64
	var alfabetWithoutE string = "abcdefghijklmnopqrstuvwxyzABCDFGHIJKLMNOPQRSTUVWXYZ"
	l := len(str)
	if l < 1 || strings.ContainsAny(str, alfabetWithoutE) {
		return "0d", nil
	}

	t, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatal(err)
	}

	if t >= 2 || t < 0 {
		s, e = timeFormat(0, 0, 0, ns)
		return s, errors.New("Error in incoming time number")
	} else if t > 1 {
		t = t - 1
	}

	seconds = t * 86400
	seconds = math.Ceil(seconds)
	hours, ost = math.Modf(seconds / 3600)
	minutes, ost = math.Modf(ost * 60)
	seconds, ost = math.Modf(ost * 60)

	if hour24 {
		if hours > 12 {
			hours -= 12
			s, e = timeFormat(hours, minutes, seconds, ns)
			return s, e
		} else {
			s, e = timeFormat(hours, minutes, seconds, ns)
			return s, e
		}
	}
	s, e = timeFormat(hours, minutes, seconds, ns)
	return s, e
}

func timeFormat(h, m, s float64, nosec bool) (o string, e error) {
	if !nosec {
		o = fmt.Sprintf("%02.0f:%02.0f", h, m)
	} else {
		o = fmt.Sprintf("%02.0f:%02.0f:%02.0f", h, m, s)
	}
	return o, e
}

func GetDurateInMinutes(start, finish string) (int, error) {
	hm_format := "15:04"
	rexp := `([2]{1}[0-3]{1})|([01]{1}[0-9]{1})[\:]{1}[0-5]{1}[0-9]{1}`
	re := regexp.MustCompile(rexp)
	validT := re.MatchString(start)
	if validT {
		t1, _ := time.Parse(hm_format, start)
		t2, _ := time.Parse(hm_format, finish)
		return int(math.Round(t2.Sub(t1).Minutes())), nil
	}
	return 0, errors.New("Time dont match")
}

// In 1900 mode, Excel takes dates in floating point numbers of days starting with Jan 1 1900.
// The days are not zero indexed, so Jan 1 1900 would be 1.
// Except that Excel pretends that Feb 29, 1900 occurred to be compatible with a bug in Lotus 123.
// So, this constant uses Dec 30, 1899 instead of Jan 1, 1900, so the diff will be correct.
// http://www.cpearson.com/excel/datetime.htm
// Info from: github.com/tealeg/xlsx data.go file
func GetDateByDayNumber(days int) (year int, month time.Month, day int) {
	startYear := 1899
	startMonth := time.December
	startDay := 30
	loc, _ := time.LoadLocation("Local")
	start := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, loc)
	resultTime := start.AddDate(0, 0, days)
	year, month, day = resultTime.Date()
	return year, month, day
}
