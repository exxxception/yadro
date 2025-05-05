package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	_ = iota
	EvRegister
	EvDrawStart
	EvOnStartLine
	EvStart
	EvOnRange
	EvHitTarget
	EvLeaveRange
	EvEnterPenalty
	EvLeavePenalty
	EvEndLap
	EvNotFinished
	EvDisqualified
	EvFinished
)

type Event struct {
	Time       time.Time
	ID         int
	Competitor string
	Extra      string
}

type Competitor struct {
	ID            string
	DrawStart     time.Time
	StartTime     time.Time
	EndTime       time.Time
	PenaltyStart  time.Time
	LapTimes      []time.Duration
	Disqualified  bool
	NotFinished   bool
	TotalPenalty  float64
	CurrentLap    int
	LapsCompleted int
	Comment       string
	ShootStats    map[int][]bool
}

type ResultsTable struct {
	CompetitorID string
	Status       string
	HitStat      string
	Laps         int
	LapTime      []string
	LapSpeed     []string
	PenaltyTime  []string
	PenaltySpeed []string
	Time         time.Duration
}

const TargetsOnLap = 5

func ProcessEvents(w io.Writer, cfg *Configuration, events []Event) {
	competitors := map[string]*Competitor{}
	table := make(map[string]*ResultsTable)

	for _, e := range events {
		comp, ok := competitors[e.Competitor]
		if !ok {
			comp = &Competitor{
				ID:         e.Competitor,
				ShootStats: make(map[int][]bool),
			}
			competitors[e.Competitor] = comp
		}

		if _, ok := table[e.Competitor]; !ok {
			table[e.Competitor] = &ResultsTable{CompetitorID: comp.ID, Laps: cfg.Laps}
		}

		switch e.ID {
		case EvRegister:
			fmt.Fprintf(w, "[%s] The competitor(%s) registered\n", e.Time.Format("15:04:05.000"), comp.ID)
		case EvDrawStart:
			t, _ := time.Parse("15:04:05.000", e.Extra)
			comp.DrawStart = t
			fmt.Fprintf(w, "[%s] The start time for the competitor(%s) was set by a draw to %s\n", e.Time.Format("15:04:05.000"), comp.ID, e.Extra)
		case EvOnStartLine:
			fmt.Fprintf(w, "[%s] The competitor(%s) is on the start line\n", e.Time.Format("15:04:05.000"), comp.ID)
		case EvStart:
			if e.Time.After(comp.DrawStart.Add(cfg.StartDelta)) {
				comp.Disqualified = true
				fmt.Fprintf(w, "[%s] The competitor(%s) is disqualified\n", e.Time.Format("15:04:05.000"), comp.ID)
				break
			}
			comp.StartTime = e.Time
			fmt.Fprintf(w, "[%s] The competitor(%s) has started\n", e.Time.Format("15:04:05.000"), comp.ID)
		case EvOnRange:
			fmt.Fprintf(w, "[%s] The competitor(%s) is on the firing range(%s)\n", e.Time.Format("15:04:05.000"), comp.ID, e.Extra)
		case EvHitTarget:
			comp.ShootStats[comp.CurrentLap] = append(comp.ShootStats[comp.CurrentLap], true)
			fmt.Fprintf(w, "[%s] The target(%s) has been hit by competitor(%s)\n", e.Time.Format("15:04:05.000"), comp.ID, e.Extra)
		case EvLeaveRange:
			fmt.Fprintf(w, "[%s] The competitor(%s) left the firing range\n", e.Time.Format("15:04:05.000"), comp.ID)
		case EvEnterPenalty:
			comp.PenaltyStart = e.Time
			fmt.Fprintf(w, "[%s] The competitor(%s) entered the penalty laps\n", e.Time.Format("15:04:05.000"), comp.ID)
		case EvLeavePenalty:
			timePenalty := e.Time.Sub(comp.PenaltyStart)
			speedPenalty := fmt.Sprintf("%.3f", float64(cfg.PenaltyLen)/timePenalty.Seconds())

			table[e.Competitor].PenaltyTime = append(table[e.Competitor].PenaltyTime, Duration2Time(timePenalty))
			table[e.Competitor].PenaltySpeed = append(table[e.Competitor].PenaltySpeed, speedPenalty)

			fmt.Fprintf(w, "[%s] The competitor(%s) left the penalty laps\n", e.Time.Format("15:04:05.000"), comp.ID)
		case EvEndLap:
			timeLap := e.Time.Sub(comp.StartTime)
			speedLap := fmt.Sprintf("%.3f", float64(cfg.LapLen)/timeLap.Seconds())

			table[e.Competitor].Time = timeLap
			table[e.Competitor].LapTime = append(table[e.Competitor].LapTime, Duration2Time(timeLap))
			table[e.Competitor].LapSpeed = append(table[e.Competitor].LapSpeed, speedLap)

			comp.LapTimes = append(comp.LapTimes, timeLap)
			comp.CurrentLap++

			fmt.Fprintf(w, "[%s] The competitor(%s) ended the main lap\n", e.Time.Format("15:04:05.000"), comp.ID)

			if comp.CurrentLap == cfg.Laps {
				comp.EndTime = e.Time
				fmt.Fprintf(w, "[%s] The competitor(%s) has finished\n", e.Time.Format("15:04:05.000"), comp.ID)
			}
		case EvNotFinished:
			comp.NotFinished = true
			comp.Comment = e.Extra
			fmt.Fprintf(w, "[%s] The competitor(%s) can`t continue: %s\n", e.Time.Format("15:04:05.000"), comp.ID, e.Extra)
		}
	}

	for _, comp := range competitors {
		t := table[comp.ID]

		switch {
		default:
			t.Status = comp.EndTime.Format("15:04:05.000")
		case comp.NotFinished:
			t.Status = "[NotFinished]"
		case comp.Disqualified:
			t.Status = "[NotStarted]"
		}

		hits := 0
		for _, lapHits := range comp.ShootStats {
			hits += len(lapHits)
		}
		t.HitStat = fmt.Sprintf("%d/%d", hits, cfg.Laps*TargetsOnLap)
	}

	CreateResultFile(table)
}

func ParseEvents(r io.Reader) ([]Event, error) {
	s := bufio.NewScanner(r)
	var evs []Event
	for s.Scan() {
		line := s.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var tstr, comp string
		var id int
		var extra string
		parts := strings.SplitN(line, "]", 2)
		tstr = strings.TrimPrefix(parts[0], "[")
		rest := strings.Fields(strings.TrimSpace(parts[1]))
		if len(rest) < 2 {
			return nil, fmt.Errorf("invalid event line: %s", line)
		}
		fmt.Sscanf(rest[0], "%d", &id)
		comp = rest[1]
		if len(rest) > 2 {
			extra = strings.Join(rest[2:], " ")
		}
		pt, err := time.Parse("15:04:05.000", tstr)
		if err != nil {
			return nil, err
		}

		evs = append(evs, Event{Time: pt, ID: id, Competitor: comp, Extra: extra})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return evs, nil
}
