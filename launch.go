package frostui

// LaunchDesktopApp starts the desktop app whose desktop file ID is id
// ("org.mozilla.firefox", without ".desktop") and does not wait for it.
//
// On Linux it runs gtk-launch, which finds the desktop file (Flatpak
// exports one per app) and starts the app as the desktop would; the app
// is a GUI app, and its end is reaped on a goroutine. It runs no shell;
// pass IDs from the app's configuration or a package manager's list, not
// typed input.
//
// Desktop file IDs are a freedesktop idea: on other systems it starts
// nothing and returns an error that wraps errors.ErrUnsupported.
func LaunchDesktopApp(id string) error { return launchDesktopApp(id) }
