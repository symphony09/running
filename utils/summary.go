package utils

import (
	"time"

	"github.com/symphony09/running"
)

type RunSummary struct {
	Count int

	Cost time.Duration

	Logs map[string][]RunLog
}

type RunLog struct {
	Start, End time.Time

	Msg string

	Err error
}

func GetRunSummary(state running.State) RunSummary {
	var summary RunSummary

	value, exists := state.Query("run_summary")
	if !exists {
		summary = RunSummary{Logs: make(map[string][]RunLog)}
	} else {
		if s, ok := value.(*RunSummary); ok {
			summary = *s
		}
	}

	return summary
}

func AddLog(state running.State, name string, start, end time.Time, msg string, err error) {
	state.Transform("run_summary", func(from interface{}) interface{} {
		if from == nil {
			from = &RunSummary{Logs: make(map[string][]RunLog)}
		}

		if summary, ok := from.(*RunSummary); ok {
			summary.Count++
			summary.Cost += end.Sub(start)
			summary.Logs[name] = append(summary.Logs[name], RunLog{
				Start: start,
				End:   end,
				Msg:   msg,
				Err:   err,
			})
			return summary
		} else {
			return from
		}
	})
}
