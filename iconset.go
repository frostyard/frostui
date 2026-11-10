package frostui

import (
	"embed"
	"strings"
)

// The kit's own icon set: symbolic icons of Lucide (https://lucide.dev),
// ISC-licensed (icons/LICENSE, NOTICE), embedded so that an app shows an
// icon where the desktop has none. Icons.Get falls back to it after the
// app's embedded icons and the desktop's theme: off Linux it is the only
// set, and on Linux it fills the names the theme lacks. Only the files
// that kitIconNames maps are in the directory, so the binary carries no
// icon nobody names (TestKitIconSetHoldsOnlyMappedIcons).
//
//go:embed icons/*.svg
var kitIconFiles embed.FS

// kitIconNames maps a freedesktop icon name, without "-symbolic", to the
// Lucide icon of the kit's set that stands for it. To add a name, copy
// the Lucide icon to icons/<lucide name>.svg and add a line here
// (docs/design/frostui.md, "The kit's icon set").
var kitIconNames = map[string]string{
	// Names the kit draws itself.
	"changes-prevent":           "lock",
	"content-loading":           "loader-circle",
	"dialog-error":              "circle-x",
	"dialog-password":           "key-round",
	"dialog-warning":            "triangle-alert",
	"edit-clear-all":            "brush-cleaning",
	"edit-copy":                 "copy",
	"edit-find":                 "search",
	"go-bottom":                 "arrow-down-to-line",
	"go-next":                   "chevron-right",
	"media-skip-forward":        "skip-forward",
	"network-server":            "server",
	"object-select":             "check",
	"open-menu":                 "menu",
	"pan-down":                  "chevron-down",
	"pan-up":                    "chevron-up",
	"process-stop":              "circle-stop",
	"sidebar-show":              "panel-left",
	"software-update-available": "circle-arrow-down",
	"system-search":             "search",
	"utilities-terminal":        "square-terminal",
	"view-refresh":              "refresh-cw",
	"window-close":              "x",

	// Navigation and views.
	"go-down":                  "arrow-down",
	"go-home":                  "house",
	"go-previous":              "chevron-left",
	"go-top":                   "arrow-up-to-line",
	"go-up":                    "arrow-up",
	"pan-end":                  "chevron-right",
	"pan-start":                "chevron-left",
	"sidebar-show-right":       "panel-right",
	"view-grid":                "layout-grid",
	"view-list":                "list",
	"view-more":                "ellipsis-vertical",
	"view-more-horizontal":     "ellipsis",
	"view-pin":                 "pin",
	"view-sort-ascending":      "arrow-up-narrow-wide",
	"view-sort-descending":     "arrow-down-wide-narrow",
	"zoom-in":                  "zoom-in",
	"zoom-out":                 "zoom-out",
	"network-offline":          "cloud-off",
	"network-transmit-receive": "arrow-down-up",

	// Editing and documents.
	"document-edit":             "pencil",
	"document-new":              "file-plus",
	"document-open":             "folder-open",
	"document-save":             "save",
	"document-send":             "send",
	"edit-cut":                  "scissors",
	"edit-delete":               "trash",
	"edit-paste":                "clipboard-paste",
	"edit-redo":                 "redo-2",
	"edit-undo":                 "undo-2",
	"folder":                    "folder",
	"folder-new":                "folder-plus",
	"folder-open":               "folder-open",
	"format-text-bold":          "bold",
	"format-text-italic":        "italic",
	"format-text-strikethrough": "strikethrough",
	"format-text-underline":     "underline",
	"image-x-generic":           "file-image",
	"insert-image":              "image",
	"insert-link":               "link",
	"list-add":                  "plus",
	"list-remove":               "minus",
	"tag":                       "tag",
	"text-x-generic":            "file-text",
	"user-trash":                "trash",

	// Mail, calendar and people.
	"alarm":                 "alarm-clock",
	"appointment-new":       "calendar-plus",
	"avatar-default":        "circle-user",
	"contact-new":           "user-plus",
	"emblem-important":      "circle-alert",
	"mail-archive":          "archive",
	"mail-attachment":       "paperclip",
	"mail-forward":          "forward",
	"mail-mark-important":   "flag",
	"mail-mark-junk":        "shield-alert",
	"mail-mark-read":        "mail-open",
	"mail-mark-unread":      "mail",
	"mail-message-new":      "square-pen",
	"mail-read":             "mail-open",
	"mail-reply-all":        "reply-all",
	"mail-reply-sender":     "reply",
	"mail-send":             "send",
	"mail-send-receive":     "refresh-cw",
	"mail-unread":           "mail",
	"system-users":          "users",
	"x-office-address-book": "book-user",
	"x-office-calendar":     "calendar",

	// The app and the system.
	"application-exit":   "log-out",
	"changes-allow":      "lock-open",
	"dialog-information": "info",
	"dialog-question":    "circle-question-mark",
	"drive-harddisk":     "hard-drive",
	"emblem-system":      "settings",
	"help-about":         "info",
	"help-browser":       "circle-question-mark",
	"input-keyboard":     "keyboard",
	"preferences-system": "settings",
}

// kitIcon returns the kit's icon for a freedesktop name, with or without
// "-symbolic", or nil when the set has none.
func kitIcon(name string) []byte {
	file, ok := kitIconNames[strings.TrimSuffix(name, "-symbolic")]
	if !ok {
		return nil
	}
	data, err := kitIconFiles.ReadFile("icons/" + file + ".svg")
	if err != nil {
		return nil
	}
	return data
}
