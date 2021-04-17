package utils

import "time"

const (
	Simple_time_timeTemplate = "2006-01-02 15:04:05"
)

func tickerDone(interval int, function func()) {
	eventsTick := time.NewTicker(time.Duration(interval) * time.Second)
	defer eventsTick.Stop()
	for {
		select {
		case <-eventsTick.C:
			function()
		}
	}
}

func StartTimer(timePeriod int, startTime string, timeTemplate string, callFunc func()) error {
	if timeTemplate == "" {
		timeTemplate = Simple_time_timeTemplate
	}
	startTimeStamp, err := time.ParseInLocation(timeTemplate, startTime, time.Local)
	if err != nil {
		return err
	}
	if timePeriod <= 0 {
		return err
	}

	go func() {
		nowTimeStamp := time.Now()
		interval := startTimeStamp.Second() - (nowTimeStamp.Second())
		if interval < 0 {
			interval = timePeriod + interval%timePeriod
		}
		time.Sleep(time.Duration(interval) * time.Second)
		tickerDone(timePeriod, func() {
			callFunc()
		})

	}()

	return nil
}
