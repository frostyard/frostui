package frostui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenerOpensAllowedURLs(t *testing.T) {
	var opened []string
	o := Opener{Launch: func(u string) error { opened = append(opened, u); return nil }}
	for _, raw := range []string{"http://example.org/a", "https://example.org/a?b=c#d", "HTTPS://example.org", "mailto:a@example.org", "mailto:?subject=hi"} {
		if err := o.OpenURL(raw); err != nil {
			t.Errorf("OpenURL(%q) = %v", raw, err)
		}
	}
	if len(opened) != 5 {
		t.Errorf("opened %q", opened)
	}
}

func TestOpenerRefusesOtherURLs(t *testing.T) {
	o := Opener{Launch: func(u string) error { t.Errorf("launched %q", u); return nil }}
	for _, raw := range []string{
		"", "example.org", "/etc/passwd", "file:///etc/passwd", "FILE:///etc/passwd", "javascript:alert(1)",
		"ftp://example.org", "smb://host/share", "data:text/html,x", "https:", "https://", "https:example.org",
		"https:\\\\example.org", "mailto:", "http://exa mple.org", "http://example.org/\nmore", "http://example.org/\x7f",
	} {
		err := o.OpenURL(raw)
		if !errors.Is(err, ErrNotAllowed) {
			t.Errorf("OpenURL(%q) = %v, want a refusal", raw, err)
		}
	}
}

func TestOpenerSchemesReplaceTheDefaultsButNeverAddFiles(t *testing.T) {
	o := Opener{Schemes: []string{"obsidian:", "file"}, Launch: func(string) error { return nil }}
	if err := o.OpenURL("obsidian://open?vault=a"); err != nil {
		t.Errorf("a listed scheme: %v", err)
	}
	if err := o.OpenURL("https://example.org"); !errors.Is(err, ErrNotAllowed) {
		t.Errorf("a default scheme not listed: %v", err)
	}
	if err := o.OpenURL("file:///etc/passwd"); !errors.Is(err, ErrNotAllowed) {
		t.Errorf("file listed as a scheme: %v", err)
	}
}

func TestOpenerReturnsTheSystemsError(t *testing.T) {
	boom := errors.New("no handler")
	o := Opener{Launch: func(string) error { return boom }}
	if err := o.OpenURL("https://example.org"); !errors.Is(err, boom) || errors.Is(err, ErrNotAllowed) {
		t.Errorf("err = %v", err)
	}
}

// vault makes a directory holding a note, and returns its path.
func vault(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "note.md"), 0o644)
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func write(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func TestOpenerOpensFilesUnderARoot(t *testing.T) {
	dir := vault(t)
	write(t, filepath.Join(dir, "sub", "photo.PNG"), 0o644)
	var opened []string
	o := Opener{Roots: []string{dir}, LaunchFile: func(p string) error { opened = append(opened, p); return nil }}
	for _, p := range []string{filepath.Join(dir, "note.md"), filepath.Join(dir, "sub", "photo.PNG"), filepath.Join(dir, "sub", "..", "note.md")} {
		if err := o.OpenFile(p); err != nil {
			t.Errorf("OpenFile(%q) = %v", p, err)
		}
	}
	if len(opened) != 3 || opened[2] != filepath.Join(dir, "note.md") {
		t.Errorf("opened %q", opened)
	}
}

func TestOpenerRefusesOtherFiles(t *testing.T) {
	dir, outside := vault(t), vault(t)
	write(t, filepath.Join(dir, "run.bin"), 0o755)
	write(t, filepath.Join(dir, "launcher.desktop"), 0o644)
	write(t, filepath.Join(dir, "setup.EXE"), 0o644)
	if err := os.Symlink(filepath.Join(outside, "note.md"), filepath.Join(dir, "away.md")); err != nil {
		t.Fatal(err)
	}
	o := Opener{Roots: []string{dir}, LaunchFile: func(p string) error { t.Errorf("launched %q", p); return nil }}
	for _, p := range []string{
		filepath.Join(outside, "note.md"),
		filepath.Join(dir, "..", filepath.Base(outside), "note.md"),
		filepath.Join(dir, "away.md"),
		filepath.Join(dir, "missing.md"),
		filepath.Join(dir, "run.bin"),
		filepath.Join(dir, "launcher.desktop"),
		filepath.Join(dir, "setup.EXE"),
		filepath.Join(dir, "sub"),
		"note.md",
		"",
	} {
		if err := o.OpenFile(p); !errors.Is(err, ErrNotAllowed) {
			t.Errorf("OpenFile(%q) = %v, want a refusal", p, err)
		}
	}
	if err := (Opener{}).OpenFile(filepath.Join(dir, "note.md")); !errors.Is(err, ErrNotAllowed) {
		t.Errorf("no roots opened a file: %v", err)
	}
}

func TestOpenerRootsFollowSymlinksAndDoNotMatchPrefixes(t *testing.T) {
	base := vault(t)
	target := filepath.Join(base, "real")
	sibling := filepath.Join(base, "real-other")
	for _, d := range []string{target, sibling} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(d, "a.md"), 0o644)
	}
	link := filepath.Join(base, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	o := Opener{Roots: []string{link}, LaunchFile: func(string) error { return nil }}
	if err := o.OpenFile(filepath.Join(target, "a.md")); err != nil {
		t.Errorf("a file under a root given through a link: %v", err)
	}
	o.Roots = []string{target}
	if err := o.OpenFile(filepath.Join(sibling, "a.md")); !errors.Is(err, ErrNotAllowed) {
		t.Errorf("a directory sharing the root's prefix: %v", err)
	}
}

func TestOpenerExtensionsAndFolders(t *testing.T) {
	dir := vault(t)
	write(t, filepath.Join(dir, "a.PDF"), 0o644)
	o := Opener{Roots: []string{dir}, Extensions: []string{".pdf"}, LaunchFile: func(string) error { return nil }}
	if err := o.OpenFile(filepath.Join(dir, "a.PDF")); err != nil {
		t.Errorf("a listed extension: %v", err)
	}
	if err := o.OpenFile(filepath.Join(dir, "note.md")); !errors.Is(err, ErrNotAllowed) {
		t.Errorf("an unlisted extension: %v", err)
	}
	o.Extensions = []string{".sh"}
	write(t, filepath.Join(dir, "a.sh"), 0o644)
	if err := o.OpenFile(filepath.Join(dir, "a.sh")); !errors.Is(err, ErrNotAllowed) {
		t.Errorf("a runnable type listed: %v", err)
	}
	o = Opener{Roots: []string{dir}, Folders: true, LaunchFile: func(string) error { return nil }}
	if err := o.OpenFile(filepath.Join(dir, "sub")); err != nil {
		t.Errorf("a folder with Folders: %v", err)
	}
}
