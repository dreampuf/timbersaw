package dashboard

import (
	"context"
	"go.uber.org/zap"
	"testing"
)

type statusTestCase struct {
	period uint32
	data []string
	expectations []uint64
}

func TestStatus_Enqueue(t *testing.T) {
	for _, i := range []statusTestCase{
		{3, []string{"/", "/posts", "/profile"}, []uint64{1, 2, 3}},
		{3, []string{"/", "/posts", ""}, []uint64{1, 2, 2}},
		{3, []string{"/", "/posts", "", "/app"}, []uint64{1, 2, 2, 2}},
		{3, []string{"/", "/posts", "", "/app", "/profile"}, []uint64{1, 2, 2, 2, 2}},
	} {
		singleTest(t, i.period, i.data, i.expectations)
	}
}

func singleTest(t *testing.T, period uint32, data []string, expectations []uint64) {
	ctx, cancel := context.WithCancel(context.TODO())
	logger := zap.NewNop()
	sugar := logger.Sugar()

	s := NewStatus(ctx, sugar, period)
	for n, line := range data {
		if line != "" {
			s.Enqueue(line)
		}
		d := s.Total(-1)
		if d != expectations[n] {
			t.Errorf("NO.%d expectation: %d, but get %d", n, expectations[n], d)
		}
		s.tick()
	}
	cancel()
}
