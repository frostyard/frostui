package frostui

// LaunchDesktopApp starts the desktop app whose desktop file ID is id
// ("org.mozilla.firefox", without ".desktop") and does not wait for it.
//
// On Linux it looks up the desktop entry by ID and launches it through GIO.
// Looking up installed applications reads the disk; call it from a worker.
// It runs no shell and does not require MyGo to be running.
//
// Desktop file IDs are a freedesktop idea: on other systems it starts
// nothing and returns an error that wraps errors.ErrUnsupported.
func LaunchDesktopApp(id string) error { return launchDesktopApp(id) }
