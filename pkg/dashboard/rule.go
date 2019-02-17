package dashboard

import (
	"github.com/dreampuf/timbersaw/pkg/log"
	"time"
)

type Rule interface {
	Verify(status *Status) bool
}

type ThresholdRule struct {
	Threshold, Lookback int
	Logger log.Logger
	fired bool
}

func (tr *ThresholdRule) Verify(s *Status) bool {
	total := s.Total(tr.Lookback)
	if int(total) >= tr.Threshold {
		tr.Logger.Warnf("High traffic generated an alert - hits = %d, triggered at %s", total, time.Now().Format(time.RFC3339))
		tr.fired = true
	} else if (tr.fired) {
		tr.Logger.Info("Threshold Alert recovered")
		tr.fired = false
	}
	return int(total) >= tr.Threshold
}
