package prodos

import (
	"time"
)

/*  49041 ($BF91)     49040 ($BF90)

        7 6 5 4 3 2 1 0   7 6 5 4 3 2 1 0
       +-+-+-+-+-+-+-+-+ +-+-+-+-+-+-+-+-+
DATE:  |    year     |  month  |   day   |
       +-+-+-+-+-+-+-+-+ +-+-+-+-+-+-+-+-+

        7 6 5 4 3 2 1 0   7 6 5 4 3 2 1 0
       +-+-+-+-+-+-+-+-+ +-+-+-+-+-+-+-+-+
TIME:  |    hour       | |    minute     |
       +-+-+-+-+-+-+-+-+ +-+-+-+-+-+-+-+-+

       49043 ($BF93)     49042 ($BF92)
*/

func DateTimeToProDOS(dateTime time.Time) []byte {
	year := dateTime.Year() % 100
	month := dateTime.Month()
	day := dateTime.Day()
	hour := dateTime.Hour()
	minute := dateTime.Minute()

	buffer := make([]byte, 4)
	buffer[0] = ((byte(month) & 15) << 5) + byte(day)
	buffer[1] = (byte(year) << 1) + (byte(month) >> 3)
	buffer[2] = byte(minute)
	buffer[3] = byte(hour)

	return buffer
}

func DateTimeFromProDOS(buffer []byte) time.Time {
	twoDigitYear := buffer[1] >> 1
	var year int
	if twoDigitYear < 76 {
		year = 2000 + int(twoDigitYear)
	} else {
		year = 1900 + int(twoDigitYear)
	}

	month := int(buffer[0]>>5 + buffer[1]&1)
	day := int(buffer[0] & 31)
	hour := int(buffer[3])
	minute := int(buffer[2])

	parsedTime := time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local)

	return parsedTime
}
