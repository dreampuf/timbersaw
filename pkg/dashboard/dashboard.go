package dashboard

import (
	"context"
	"github.com/dreampuf/timbersaw/pkg/formater"
	"github.com/dreampuf/timbersaw/pkg/log"
	"sync"
	"time"
)

type Dashboard struct {
	logger log.Logger
	ch <-chan string
	status *Status
	rules []Rule
}

type DashboardOptions struct {
	Ctx context.Context
	Logger log.Logger
	DataChannel <-chan string

	PeriodInSeconds uint32

}

func NewDashboard(opt DashboardOptions) *Dashboard {
	status := NewStatus(opt.Ctx, opt.Logger, opt.PeriodInSeconds)
	return &Dashboard{
		logger: opt.Logger,
		ch: opt.DataChannel,
		status: status,
	}
}

func (d *Dashboard) Run(ctx context.Context) {
	go d.handleData()
	go d.status.Run()
	go d.CheckRules(ctx)

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			r := d.status.buckets
			curRing := r.Next()
			maps := []*sync.Map{&r.Value.(*Entity).hash}
			for curRing != r {
				maps = append(maps, &curRing.Value.(*Entity).hash)
				curRing = curRing.Next()
			}
			mergedMap := MergeMap(maps...)
			for _, pair := range SortedMapByValue(mergedMap) {
				d.logger.Infof("%s: %d", pair.key, pair.val)
			}
		}
	}
}

func (d *Dashboard) handleData() {
	logFormater := formater.NewHTTPCommanLogFormatter()
	for line := range d.ch {
		if entity := logFormater.Format(line); entity != nil {
			d.status.Enqueue(URLExtraSection(entity.Request))
			logFormater.Put(entity)
		}
	}
}

func (d *Dashboard) AddRule(rule Rule) {
	d.rules = append(d.rules, rule)
}

func (d *Dashboard) CheckRules(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <- ctx.Done():
			ticker.Stop()
			return
		case <- ticker.C:
			for _, rule := range d.rules {
				rule.Verify(d.status)
			}
		}
	}
}
