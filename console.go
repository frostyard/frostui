package frostui

import (
	"strings"
	"sync"
	"time"
)

// Level is the kind of a console line, which the console's filter and
// colors follow.
type Level uint8

const (
	// LevelDebug is the app's own diagnostic log.
	LevelDebug Level = iota
	// LevelOutput is a line a command printed.
	LevelOutput
	// LevelInfo is a step the app took or finished.
	LevelInfo
	// LevelCommand is a command the app started, shown as "$ argv".
	LevelCommand
	// LevelDryRun is a command dry-run mode did not start.
	LevelDryRun
	// LevelWarn is something that went partly wrong.
	LevelWarn
	// LevelError is a failure.
	LevelError
	levelCount
)

// LevelNames are the names the console's filter buttons show, by Level.
var LevelNames = [levelCount]string{"Debug", "Output", "Info", "Command", "Dry run", "Warn", "Error"}

// Line is one console line.
type Line struct {
	Seq    uint64
	At     time.Time
	Level  Level
	Source string
	Text   string
}

// Console holds the newest lines of everything an app ran and logged, up
// to a capacity. Add is safe from any goroutine; the rest is for the UI
// thread too, but locks, so a worker may read.
type Console struct {
	mu       sync.Mutex
	lines    []Line // ring buffer
	start    int
	count    int
	seq      uint64
	notify   func()
	now      func() time.Time
	cleared  uint64
	warnings int
	errors   int
}

// NewConsole returns a console keeping the newest capacity lines.
func NewConsole(capacity int) *Console {
	if capacity < 16 {
		capacity = 16
	}
	return &Console{lines: make([]Line, capacity), now: time.Now}
}

// SetNotify sets the function called after lines are added, from the
// goroutine that added them: usually one that wakes the UI.
func (c *Console) SetNotify(fn func()) {
	c.mu.Lock()
	c.notify = fn
	c.mu.Unlock()
}

// Add appends a line. Text with several lines becomes several lines, all
// of level.
func (c *Console) Add(level Level, source, text string) {
	c.add(source, splitLines(text), func(int) Level { return level })
}

// splitLines splits text into the lines Add appends: a final newline ends
// the last line rather than starting an empty one, and each line loses a
// trailing carriage return.
func splitLines(text string) []string {
	parts := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, part := range parts {
		parts[i] = strings.TrimRight(part, "\r")
	}
	return parts
}

// add appends texts, the ith of level levelOf(i), under one lock, and
// notifies once.
func (c *Console) add(source string, texts []string, levelOf func(i int) Level) {
	c.mu.Lock()
	for i, text := range texts {
		level := levelOf(i)
		c.seq++
		line := Line{Seq: c.seq, At: c.now(), Level: level, Source: source, Text: text}
		if c.count < len(c.lines) {
			c.lines[(c.start+c.count)%len(c.lines)] = line
			c.count++
		} else {
			c.lines[c.start] = line
			c.start = (c.start + 1) % len(c.lines)
		}
		switch level {
		case LevelWarn:
			c.warnings++
		case LevelError:
			c.errors++
		}
	}
	notify := c.notify
	c.mu.Unlock()
	if notify != nil {
		notify()
	}
}

// Addf is Add with a format.
func (c *Console) Addf(level Level, source, format string, args ...any) {
	c.Add(level, source, sprintf(format, args...))
}

// Lines returns the lines kept, oldest first, that match: a line matches
// when its level is in levels (nil for all) and its text or source holds
// query, ignoring case.
func (c *Console) Lines(levels *LevelSet, query string) []Line {
	c.mu.Lock()
	defer c.mu.Unlock()
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]Line, 0, c.count)
	for i := 0; i < c.count; i++ {
		l := c.lines[(c.start+i)%len(c.lines)]
		if levels != nil && !levels.Has(l.Level) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(l.Text), q) && !strings.Contains(strings.ToLower(l.Source), q) {
			continue
		}
		out = append(out, l)
	}
	return out
}

// Seq is the sequence number of the newest line, 0 when none was added.
func (c *Console) Seq() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.seq
}

// Counts returns how many warnings and errors were added since the last
// Clear.
func (c *Console) Counts() (warnings, errors int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.warnings, c.errors
}

// Clear removes every line.
func (c *Console) Clear() {
	c.mu.Lock()
	c.start, c.count, c.warnings, c.errors = 0, 0, 0, 0
	c.cleared = c.seq
	c.mu.Unlock()
}

// Text renders lines as plain text, one per line, for the clipboard.
func Text(lines []Line) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.At.Format("15:04:05"))
		b.WriteString("  ")
		b.WriteString(LevelNames[l.Level])
		b.WriteString("  ")
		if l.Source != "" {
			b.WriteString("[" + l.Source + "] ")
		}
		b.WriteString(l.Text)
		b.WriteByte('\n')
	}
	return b.String()
}

// LevelSet is a set of levels.
type LevelSet uint16

// AllLevels is every level.
const AllLevels = LevelSet(1<<levelCount - 1)

// DefaultLevels hides the app's debug log.
const DefaultLevels = AllLevels &^ (1 << LevelDebug)

// Has reports whether l is in the set.
func (s *LevelSet) Has(l Level) bool { return *s&(1<<l) != 0 }

// Toggle adds l to the set or takes it out.
func (s *LevelSet) Toggle(l Level) { *s ^= 1 << l }

// Writer returns an io.Writer whose writes become lines of level from
// source, as a log.Logger's output: each write is one or more lines, split
// at its newlines. The writer keeps no partial line between writes, so a
// write that does not end in a newline still ends in a line of its own; a
// log.Logger's writes always end in one.
func (c *Console) Writer(level Level, source string) *ConsoleWriter {
	return &ConsoleWriter{c: c, level: level, source: source}
}

// ConsoleWriter is the io.Writer Console.Writer returns. It does not strip
// the timestamp prefix the standard logger adds: for lines without it, set
// the logger's flags to 0 or strip it before the write.
type ConsoleWriter struct {
	c      *Console
	level  Level
	source string
	// Classify, when set, picks the level of each line from its text: it
	// runs once per line, so the lines of one write can differ in level.
	// It runs on the writing goroutine, outside the console's lock.
	Classify func(text string) Level
}

func (w *ConsoleWriter) Write(p []byte) (int, error) {
	texts := splitLines(string(p))
	if w.Classify == nil {
		w.c.add(w.source, texts, func(int) Level { return w.level })
		return len(p), nil
	}
	levels := make([]Level, len(texts))
	for i, text := range texts {
		levels[i] = w.Classify(text)
	}
	w.c.add(w.source, texts, func(i int) Level { return levels[i] })
	return len(p), nil
}
