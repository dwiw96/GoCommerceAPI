package converter

import (
	"log"
	"strconv"
	"time"
)

func ConvertStrToDate(s string) time.Time {
	date, err := time.Parse("2006-1-2", s)
	if err != nil {
		log.Println("error ConvertStrToDate(), msg:", err)
	}

	return date
}

func ConvertStrToInt(s string) (int, error) {
	res, err := strconv.Atoi(s)
	return res, err
}

func ConvertInt32ToString(n int32) string {
	return strconv.Itoa(int(n))
}
