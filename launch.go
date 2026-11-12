package frostui

// LaunchDesktopApp starts the desktop app whose desktop file ID is id
// ("org.mozilla.firefox", without ".desktop") and does not wait for it.
//
// On Linux it looks up the desktop entry by ID and launches it through GIO.
// Missing, hidden or unlaunchable entries fail before the call returns.
// Lookup and activation may block: call it from a worker. It is safe from
// any goroutine and does not require MyGo to be running. The application
// inherits neither stdout nor stderr; GIO is called without a display launch
// context, so startup notification and Wayland activation are not supplied.
// Pass IDs from configuration or a package manager's list, not typed input:
// even entries hidden from the desktop menu can be launched by ID.
//
// Desktop file IDs are a freedesktop idea: on other systems it starts
// nothing and returns an error that wraps errors.ErrUnsupported.
func LaunchDesktopApp(id string) error { return launchDesktopApp(id) }
