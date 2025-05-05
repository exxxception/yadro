package main

import (
	"encoding/json"
	"os"
	"time"
)

type Configuration struct {
	Laps        int           `json:"laps"`
	LapLen      float64       `json:"lapLen"`
	PenaltyLen  float64       `json:"penaltyLen"`
	FiringLines int           `json:"firingLines"`
	Start       time.Duration `json:"start"`
	StartDelta  time.Duration `json:"startDelta"`
}

func ParseConfig(path string) (*Configuration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var raw struct {
		Laps        int     `json:"laps"`
		LapLen      float64 `json:"lapLen"`
		PenaltyLen  float64 `json:"penaltyLen"`
		FiringLines int     `json:"firingLines"`
		Start       string  `json:"start"`
		StartDelta  string  `json:"startDelta"`
	}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, err
	}

	startT, err := time.Parse("15:04:05.000", raw.Start)
	if err != nil {
		return nil, err
	}
	parsedDelta, err := time.Parse("15:04:05", raw.StartDelta)
	delta := time.Duration(parsedDelta.Hour())*time.Hour + time.Duration(parsedDelta.Minute())*time.Minute + time.Duration(parsedDelta.Second())*time.Second
	return &Configuration{
		Laps:        raw.Laps,
		LapLen:      raw.LapLen,
		PenaltyLen:  raw.PenaltyLen,
		FiringLines: raw.FiringLines,
		Start:       time.Duration(startT.Hour())*time.Hour + time.Duration(startT.Minute())*time.Minute + time.Duration(startT.Second())*time.Second + time.Duration(startT.Nanosecond()),
		StartDelta:  delta,
	}, err
}
