package utility

import (
	"time"
	"walk-server/global"
)

// GetCurrentDate 获取当前的天数
func GetCurrentDate() uint8 {
	startTime, _ := time.ParseInLocation(
		time.DateTime,
		global.Config.GetString("startDate"),
		time.Local,
	)

	startTime = time.Date(
		startTime.Year(),
		startTime.Month(),
		startTime.Day(),
		0, 0, 0, 0,
		startTime.Location(),
	)

	return uint8(time.Since(startTime).Hours() / 24)
}

func CanOpenApi() bool {
	startTime, _ := time.ParseInLocation(
		time.DateTime,
		global.Config.GetString("startDate"),
		time.Local,
	)
	return time.Now().After(startTime) && time.Now().Hour() >= 6
}

func CanSubmit() bool {
	return time.Now().Hour() >= 12
}
