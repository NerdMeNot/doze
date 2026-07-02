package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run -- <command> [args...]",
		Short: "Ensure the daemon is up, then run a command",
		Long: "run ensures the doze daemon is up (so instances boot on first connect and\n" +
			"reap when idle) and then executes the command — useful as a wrapper around a\n" +
			"test or dev-server command so the backends are awake before it connects. doze\n" +
			"injects nothing into the environment: because every instance has an explicit\n" +
			"port, your connection strings are stable — put them in your app config, or\n" +
			"declare the app as a `process` block to have its dependencies' URLs injected.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			// Ensure the daemon is up so connections boot instances on demand.
			if !daemonRunning(cfg) {
				if err := startDaemon(cfg); err != nil {
					return err
				}
			}
			c := exec.Command(args[0], args[1:]...)
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			err = c.Run()
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				os.Exit(ee.ExitCode())
			}
			return err
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}
