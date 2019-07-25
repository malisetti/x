package bringer

import (
	"context"
	"time"

	"github.com/mseshachalam/x/app"
)

// PeriodicBringer sends any bringer periodically to maintainer
type PeriodicBringer struct {
	Ctx      context.Context
	Interval time.Duration
	Bringer  app.Bringer
}

// Bring gives a hn bringer periodically
func (pb *PeriodicBringer) Bring() <-chan app.Bringer {
	out := make(chan app.Bringer)
	go func() {
		defer close(out)
		pb.Bringer.SetContext(pb.Ctx)

		out <- pb.Bringer

		ticker := time.NewTicker(pb.Interval)
		for {
			select {
			case <-ticker.C:
				out <- pb.Bringer
			case <-pb.Ctx.Done():
				break
			}
		}
	}()

	return out
}
