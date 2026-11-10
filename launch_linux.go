package frostui

import (
	"os"
	"os/exec"
)

func launchDesktopApp(id string) error {
	cmd := exec.Command("gtk-launch", id) //nolint:gosec // see LaunchDesktopApp
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
