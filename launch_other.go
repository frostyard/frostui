//go:build !linux

package frostui

import (
	"errors"
	"fmt"
)

func launchDesktopApp(id string) error {
	return fmt.Errorf("frostui: launch %q: desktop file IDs need a freedesktop desktop: %w", id, errors.ErrUnsupported)
}
