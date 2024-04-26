package cmd

import (
	"calendar-sync/pkg"
	"calendar-sync/pkg/clients"
	"calendar-sync/pkg/web"
	"context"
	"fmt"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"

	"os/exec"
	"runtime"
	"time"
)

var authCmd = cobra.Command{
	Use: "auth",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("usage: calendar-sync auth <email-to-add>")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		emailToAdd := args[0]

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		state := uuid.New()
		results := make(chan web.Result)
		server := web.New(pkg.ListenPort, clients.Config, results)
		go server.Run(ctx, cancel, state)

		url := clients.Config.AuthCodeURL(state)
		if err := launchBrowser(url); err != nil {
			return errors.Wrap(err, "failed to launch browser")
		}

		var result web.Result
		select {
		case <-time.After(15 * time.Minute):
			return errors.New("timeout waiting for auth code")
		case <-ctx.Done():
			return errors.New("canceled")
		case result = <-results:
			break
		}

		c, err := client.Dial(client.Options{})
		if err != nil {
			cancel()
			return errors.Wrap(err, "failed to dial the temporal server")
		}
		defer c.Close()

		if err = triggerSchedule(ctx, c, result.SourceCalendarID, emailToAdd, result.Token, oneOff); err != nil {
			return errors.Wrap(err, "failed to trigger schedule")
		}

		return nil
	},
}

var oneOff bool

func init() {
	flags := authCmd.Flags()
	flags.BoolVarP(&oneOff, "one-off", "", false, "trigger a sync right away")
}

func launchBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func init() {
	rootCmd.AddCommand(&authCmd)
}
