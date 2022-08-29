package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseDuration parses an ISO 8601 string representing a duration,
// and returns the resultant golang time.Duration instance.
func ParseDuration(isoDuration string) (float64, error) {
	re := regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:.\d+)?)S)?$`)
	matches := re.FindStringSubmatch(isoDuration)
	if matches == nil {
		return 0, errors.New("input string is of incorrect format")
	}

	seconds := 0.0

	//skipping years and months

	//days
	if matches[3] != "" {
		f, err := strconv.ParseFloat(matches[3], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 24 * 60 * 60)
	}
	//hours
	if matches[4] != "" {
		f, err := strconv.ParseFloat(matches[4], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 60 * 60)
	}
	//minutes
	if matches[5] != "" {
		f, err := strconv.ParseFloat(matches[5], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 60)
	}
	//seconds & milliseconds
	if matches[6] != "" {
		f, err := strconv.ParseFloat(matches[6], 32)
		if err != nil {
			return 0, err
		}

		seconds += f
	}

	return seconds, nil
}

// FormatDuration returns an ISO 8601 duration string.
func FormatDuration(dur time.Duration) string {
	return "PT" + strings.ToUpper(dur.Truncate(time.Millisecond).String())
}

func ms2likeISOFormat(ms int) string {
	nano := ms * 1000000

	t := time.Date(1970, time.January, 1, 0, 0, 0, nano, time.UTC)
	format := "2006-01-02T15:04:05.999Z"
	iso := t.UTC().Format(format)

	if len([]rune(iso)) != len([]rune(format)) {
		idx := strings.Index(iso, ".")
		if idx == -1 {
			// 2006-01-02T15:04:05Z
			iso = iso[:len([]rune(iso))-1] + ".000Z"
		} else {
			// 2006-01-02T15:04:05.0Z
			// 2006-01-02T15:04:05.00Z
			// 2006-01-02T15:04:05.000Z
			mili_str := iso[idx+1 : len([]rune(iso))-1]
			for {
				if len(mili_str) == 3 {
					break
				}
				mili_str = mili_str + "0"
			}
			iso = iso[:idx] + "." + mili_str + "Z"
		}
	}

	trimmedIso := iso[8 : len([]rune(iso))-1]
	day_str := trimmedIso[0:2]
	day, _ := strconv.Atoi(day_str)
	dayStartFromZero := fmt.Sprintf("%02d", day-1)
	isoOnlyTime := trimmedIso[3:]
	return dayStartFromZero + ":" + isoOnlyTime
}

func likeIso2Float(likeIso string) float64 {
	splitted := strings.Split(likeIso, ":")
	days, err := strconv.ParseFloat(splitted[0], 64)
	if err != nil {
		panic(err)
	}

	hours, err := strconv.ParseFloat(splitted[1], 64)
	if err != nil {
		panic(err)
	}

	minutes, err := strconv.ParseFloat(splitted[2], 64)
	if err != nil {
		panic(err)
	}

	seconds, err := strconv.ParseFloat(splitted[3], 64)
	if err != nil {
		panic(err)
	}

	return days*24*60*60 + hours*60*60 + minutes*60 + seconds
}
