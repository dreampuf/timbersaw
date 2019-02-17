package dashboard

import (
	"container/ring"
	"context"
	"github.com/dreampuf/timbersaw/pkg/log"
	"sync"
	"sync/atomic"
	"time"
)

type Entity struct {
	hash      sync.Map
	timestamp int64
}

type Status struct {
	total   uint64
	buckets *ring.Ring
	p       uint32
	ctx     context.Context
	logger  log.Logger

	Interval time.Duration
	PeriodInSeconds uint32
}

func NewStatus(ctx context.Context, logger log.Logger, periodInSeconds uint32) *Status {
	r := ring.New(int(periodInSeconds))
	for i := uint32(0); i < periodInSeconds; i ++ {
		r.Value = &Entity{
			hash:      sync.Map{},
			timestamp: time.Now().Unix(),
		}
		r = r.Next()
	}
	return &Status{
		ctx:     ctx,
		total:   0,
		buckets: r,
		p:       0,
		logger:  logger,

		Interval: time.Second,
		PeriodInSeconds: periodInSeconds,
	}
}

func (s *Status) Run() {
	ticker := time.NewTicker(s.Interval)
forloop:
	for {
		select {
		case <-s.ctx.Done():
			ticker.Stop()
			break forloop
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *Status) tick() {
	s.buckets = s.buckets.Next()
	bucket := s.buckets.Value.(*Entity)
	bucket.hash.Range(func(key, value interface{}) bool {
		bucket.hash.Delete(key)
		return true
	})
	bucket.timestamp = time.Now().Unix()
}

func (s *Status) Enqueue(url string) {
	entity := s.buckets.Value.(*Entity)
	val := uint64(1)

	if actualVal, loaded := entity.hash.LoadOrStore(url, &val); loaded {
		atomic.AddUint64(actualVal.(*uint64), 1)
	}
}

func (s *Status) Total(lookback int) uint64 {
	n := 0
	r := s.buckets
	cur := r
	total := uint64(0)
	for true {
		cur.Value.(*Entity).hash.Range(func(key, value interface{}) bool {
			total += *value.(*uint64)
			return true
		})
		n ++
		cur = cur.Next()
		if (lookback == -1 && cur == r) || n == lookback {
			break
		}
	}
	return total
}
