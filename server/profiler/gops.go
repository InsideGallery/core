package profiler

import (
	"log/slog"
	"os"

	"github.com/google/gops/agent"
)

// GOPS run gops instance and use GOPS_ADDR if it present as listener addr
func GOPS() func() {
	gopsAddr := os.Getenv("GOPS_ADDR")
	if gopsAddr == "" {
		return func() {}
	}

	if err := agent.Listen(agent.Options{Addr: gopsAddr}); err != nil {
		slog.Default().Error("Error start gops agent", "err", err)

		return func() {}
	}

	return func() {
		agent.Close()
	}
}
