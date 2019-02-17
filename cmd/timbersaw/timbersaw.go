package main

import (
	"context"
	"flag"
	"github.com/dreampuf/timbersaw/pkg/dashboard"
	"github.com/dreampuf/timbersaw/pkg/log"
	"github.com/dreampuf/timbersaw/pkg/watcher"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

var (
	path        = flag.String("path", "/var/log/nginx/", "Path of Logs")
	refreshRate = flag.Uint("rate", 10, "Refresh Rate of dashboard")
	window      = flag.Uint("window", 120, "Monitor period")
	threshold   = flag.Uint("threshold", 1200, "Alert threshold")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.TODO())
	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapConfig),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	)
	logger := zap.New(zapCore)
	defer logger.Sync()
	var sugar log.Logger = logger.Sugar()
	ch := make(chan string)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		w := watcher.NewWatcher(sugar, *path, ch)
		return w.Watch(gCtx)
	})
	g.Go(func() error {
		var rule dashboard.Rule = &dashboard.ThresholdRule{
			Threshold: int(*threshold),
			Lookback:  int(*window),
			Logger:    sugar,
		}

		d := dashboard.NewDashboard(dashboard.DashboardOptions{
			Ctx:             ctx,
			Logger:          sugar,
			DataChannel:     ch,
			PeriodInSeconds: uint32(*window),
		})
		d.AddRule(rule)
		d.Run(gCtx)
		return nil
	})
	g.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigs:
			cancel()
		case <-gCtx.Done():
		}
		return nil
	})

	sugar.Info("Timbersaw started.")
	if err := g.Wait(); err != nil {
		sugar.Error(err)
	}
}
