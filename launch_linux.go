package frostui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ebitengine/purego"
)

var desktopGIO struct {
	once    sync.Once
	err     error
	newInfo func(*byte) uintptr
	launch  func(uintptr, uintptr, uintptr, uintptr) bool
	unref   func(uintptr)
}

func loadDesktopGIO() error {
	desktopGIO.once.Do(func() {
		lib, err := purego.Dlopen("libgio-2.0.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			desktopGIO.err = err
			return
		}
		purego.RegisterLibFunc(&desktopGIO.newInfo, lib, "g_desktop_app_info_new")
		purego.RegisterLibFunc(&desktopGIO.launch, lib, "g_app_info_launch")
		purego.RegisterLibFunc(&desktopGIO.unref, lib, "g_object_unref")
	})
	return desktopGIO.err
}

func launchDesktopApp(id string) error {
	if id == "" || strings.ContainsRune(id, '\x00') {
		return fmt.Errorf("invalid desktop application ID %q", id)
	}
	if err := loadDesktopGIO(); err != nil {
		return fmt.Errorf("load GIO: %w", err)
	}
	info := desktopGIO.newInfo(cString(strings.TrimSuffix(id, ".desktop") + ".desktop"))
	if info == 0 {
		return fmt.Errorf("desktop application %q not found", id)
	}
	defer desktopGIO.unref(info)
	if !desktopGIO.launch(info, 0, 0, 0) {
		return fmt.Errorf("launch desktop application %q failed", id)
	}
	return nil
}
