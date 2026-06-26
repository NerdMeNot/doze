package main

import (
	"github.com/spf13/cobra"

	"github.com/doze-dev/doze/internal/daemon"
	"github.com/doze-dev/doze/internal/tui"
)

func dashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dash",
		Short: "Launch the live TUI dashboard",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			// The dashboard reflects live daemon state; start it if needed.
			if !daemonRunning(cfg) {
				if err := startDaemon(cfg); err != nil {
					return err
				}
			}
			return tui.Run(daemon.ControlSocketPath(cfg))
		},
	}
}
