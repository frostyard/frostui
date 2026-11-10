package frostui

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/egoist/mygo"
)

// ErrNotAllowed is what an Opener's refusals wrap: the target is not on
// its allow-lists.
var ErrNotAllowed = errors.New("not allowed")

// DefaultOpenSchemes are the URL schemes an Opener with no Schemes opens.
func DefaultOpenSchemes() []string { return []string{"http", "https", "mailto"} }

// runnableExtensions is a deny list of file types that run, install or
// redirect when the system opens them, or mount and hide what they hold:
// an Opener never opens them. It is a backstop, and cannot be complete:
// Opener.Extensions is the allow-list for files from outside the app.
var runnableExtensions = []string{
	".app", ".appimage", ".appinstaller", ".application", ".appref-ms", ".appx", ".appxbundle",
	".bash", ".bat", ".chm", ".cmd", ".com", ".command", ".cpl", ".deb", ".desktop", ".dmg",
	".exe", ".fileloc", ".flatpak", ".flatpakref", ".gadget", ".hta", ".img", ".inetloc", ".inf",
	".iso", ".jar", ".js", ".jse", ".library-ms", ".lnk", ".mpkg", ".msc", ".msi", ".msix",
	".msixbundle", ".msp", ".mst", ".pif", ".pkg", ".pl", ".ps1", ".ps1xml", ".psc1", ".psd1",
	".psm1", ".py", ".pyw", ".pyz", ".rb", ".reg", ".rpm", ".run", ".scf", ".scr", ".sct",
	".search-ms", ".settingcontent-ms", ".sh", ".terminal", ".url", ".vb", ".vbe", ".vbs",
	".vhd", ".vhdx", ".webloc", ".workflow", ".ws", ".wsc", ".wsf", ".wsh", ".xbap", ".xll",
}

// Opener opens links and files with the system's handlers, behind
// allow-lists, for what comes from outside the app: a link in a message
// or a note, an attachment. Its zero value opens http, https and mailto
// links, and no file. Anything else is refused with an error wrapping
// ErrNotAllowed. It holds no state, and is safe from any goroutine.
type Opener struct {
	// Schemes are the URL schemes OpenURL opens, in any case; nil is
	// DefaultOpenSchemes. A "file" entry opens nothing: files go through
	// OpenFile, under Roots.
	Schemes []string
	// Roots are the directories OpenFile opens files under; none opens
	// no file. Symbolic links are followed before the check, so one
	// under a root that leads out of it is refused.
	Roots []string
	// Extensions, when set, are the only file extensions OpenFile opens,
	// as ".pdf", in any case. Set it for files from outside the app, as
	// attachments: without it OpenFile relies on a deny list of types that
	// run, install or redirect, a backstop that cannot be complete. Types
	// on the deny list are refused whether listed here or not.
	Extensions []string
	// Folders lets OpenFile open directories under Roots, in the file
	// manager.
	Folders bool
	// Launch and LaunchFile open what passed the checks; nil uses MyGo's
	// Shell.OpenExternal and Shell.OpenPath. A test sets them.
	Launch     func(url string) error
	LaunchFile func(path string) error
}

// refused is an error saying why target is not opened.
func refused(target, why string) error {
	return fmt.Errorf("frostui: refused to open %q: %s: %w", target, why, ErrNotAllowed)
}

// CheckURL reports whether OpenURL would open raw: the URL as it would be
// opened, or why not.
func (o Opener) CheckURL(raw string) (string, error) {
	for _, r := range raw {
		if r <= ' ' || r == 0x7f {
			return "", refused(raw, "it holds spaces or control characters")
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", refused(raw, "it is not a URL")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme == "" {
		return "", refused(raw, "it has no scheme")
	}
	if scheme == "file" {
		return "", refused(raw, "a file goes through OpenFile")
	}
	allowed := o.Schemes
	if allowed == nil {
		allowed = DefaultOpenSchemes()
	}
	if !slices.ContainsFunc(allowed, func(s string) bool { return strings.EqualFold(strings.TrimSuffix(s, ":"), scheme) }) {
		return "", refused(raw, "the scheme "+scheme+" is not allowed")
	}
	switch {
	case (scheme == "http" || scheme == "https") && u.Hostname() == "":
		return "", refused(raw, "it has no host")
	case u.Opaque == "" && u.Host == "" && u.Path == "" && u.RawQuery == "":
		return "", refused(raw, "it names nothing")
	}
	return u.String(), nil
}

// OpenURL opens raw with the system's handler for its scheme, when it is
// on the allow-list. It waits for the system, which may take a while:
// call it from a worker rather than a view.
func (o Opener) OpenURL(raw string) error {
	target, err := o.CheckURL(raw)
	if err != nil {
		return err
	}
	if o.Launch != nil {
		return o.Launch(target)
	}
	return mygo.Shell.OpenExternal(target)
}

// CheckFile reports whether OpenFile would open path, which must be
// absolute: its real path, symbolic links followed, or why not.
func (o Opener) CheckFile(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", refused(path, "the path is not absolute")
	}
	if len(o.Roots) == 0 {
		return "", refused(path, "no folder is allowed")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", refused(path, "it cannot be read: "+err.Error())
	}
	inside := false
	for _, root := range o.Roots {
		r, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		if rel, err := filepath.Rel(r, resolved); err == nil && filepath.IsLocal(rel) {
			inside = true
			break
		}
	}
	if !inside {
		return "", refused(path, "it is outside the allowed folders")
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", refused(path, "it cannot be read: "+err.Error())
	}
	name := filepath.Base(resolved)
	if runtime.GOOS == "windows" {
		// Windows drops trailing dots and spaces, and reads what follows
		// a colon as a stream of the file.
		if strings.Contains(name, ":") {
			return "", refused(path, "the name holds a colon")
		}
		name = strings.TrimRight(name, ". ")
	}
	ext := strings.ToLower(filepath.Ext(name))
	if slices.Contains(runnableExtensions, ext) {
		return "", refused(path, "files of the type "+ext+" can run, install or redirect")
	}
	switch {
	case info.IsDir():
		if !o.Folders {
			return "", refused(path, "it is a folder")
		}
		return resolved, nil
	case !info.Mode().IsRegular():
		return "", refused(path, "it is not a regular file")
	case info.Mode().Perm()&0o111 != 0:
		return "", refused(path, "it is executable")
	}
	if o.Extensions != nil && !slices.ContainsFunc(o.Extensions, func(e string) bool { return strings.EqualFold(e, ext) }) {
		return "", refused(path, "the type "+ext+" is not allowed")
	}
	return resolved, nil
}

// OpenFile opens the file at path with the system's handler for its
// type, when it is under an allowed root and passes the checks. It waits
// for the system: call it from a worker rather than a view.
func (o Opener) OpenFile(path string) error {
	target, err := o.CheckFile(path)
	if err != nil {
		return err
	}
	if o.LaunchFile != nil {
		return o.LaunchFile(target)
	}
	return mygo.Shell.OpenPath(target)
}
