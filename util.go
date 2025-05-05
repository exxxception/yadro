package main

import (
	"fmt"
	"strings"
	"time"
)

func Duration2Time(d time.Duration) string {
	const (
		msPerSecond = 1000
		msPerMinute = 60 * msPerSecond
		msPerHour   = 60 * msPerMinute
	)

	totalMilliseconds := d.Milliseconds()

	h := totalMilliseconds / msPerHour
	m := (totalMilliseconds % msPerHour) / msPerMinute
	s := (totalMilliseconds % msPerMinute) / msPerSecond
	ms := totalMilliseconds % msPerSecond

	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func TransformResultTable(table *ResultsTable) string {
	output := fmt.Sprintf("[%s] %s [", table.Status, table.CompetitorID)

	output += transformLapData(table.LapTime, table.LapSpeed, table.Laps)
	output += "] ["
	output += transformPenaltyData(table.PenaltyTime, table.PenaltySpeed)
	output += "] " + table.HitStat

	return output
}

func transformLapData(lapTime, lapSpeed []string, totalLaps int) string {
	var parts []string
	for i := 0; i < totalLaps; i++ {
		if i < len(lapTime) {
			parts = append(parts, fmt.Sprintf("{%s, %s}", lapTime[i], lapSpeed[i]))
		} else {
			parts = append(parts, "{,}")
		}
	}
	return strings.Join(parts, ", ")
}

func transformPenaltyData(penaltyTime, penaltySpeed []string) string {
	var parts []string
	for i := 0; i < len(penaltyTime); i++ {
		parts = append(parts, fmt.Sprintf("{%s, %s}", penaltyTime[i], penaltySpeed[i]))
	}
	return strings.Join(parts, ", ")
}
