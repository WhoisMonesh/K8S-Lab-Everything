package cli

import (
	"os"
	"strings"
)

type Theme struct {
	Reset     string
	Bold      string
	Dim       string
	Red       string
	Green     string
	Yellow    string
	Blue      string
	Magenta   string
	Cyan      string
	White     string
	BgRed     string
	BgGreen   string
	BgYellow  string
	BgBlue    string
	BgMagenta string
	BgCyan    string
	BrRed     string
	BrGreen   string
	BrYellow  string
	BrBlue    string
	BrMagenta string
	BrCyan    string
	BrWhite   string
	DimW      string
}

var (
	DarkTheme = &Theme{
		Reset:     "\033[0m",
		Bold:      "\033[1m",
		Dim:       "\033[2m",
		Red:       "\033[31m",
		Green:     "\033[32m",
		Yellow:    "\033[33m",
		Blue:      "\033[34m",
		Magenta:   "\033[35m",
		Cyan:      "\033[36m",
		White:     "\033[37m",
		BgRed:     "\033[41m",
		BgGreen:   "\033[42m",
		BgYellow:  "\033[43m",
		BgBlue:    "\033[44m",
		BgMagenta: "\033[45m",
		BgCyan:    "\033[46m",
		BrRed:     "\033[91m",
		BrGreen:   "\033[92m",
		BrYellow:  "\033[93m",
		BrBlue:    "\033[94m",
		BrMagenta: "\033[95m",
		BrCyan:    "\033[96m",
		BrWhite:   "\033[97m",
		DimW:      "\033[90m",
	}

	LightTheme = &Theme{
		Reset:     "\033[0m",
		Bold:      "\033[1m",
		Dim:       "\033[2m",
		Red:       "\033[31m",
		Green:     "\033[32m",
		Yellow:    "\033[33m",
		Blue:      "\033[34m",
		Magenta:   "\033[35m",
		Cyan:      "\033[36m",
		White:     "\033[37m",
		BgRed:     "\033[41m",
		BgGreen:   "\033[42m",
		BgYellow:  "\033[43m",
		BgBlue:    "\033[44m",
		BgMagenta: "\033[45m",
		BgCyan:    "\033[46m",
		BrRed:     "\033[91m",
		BrGreen:   "\033[92m",
		BrYellow:  "\033[93m",
		BrBlue:    "\033[94m",
		BrMagenta: "\033[95m",
		BrCyan:    "\033[36m",
		BrWhite:   "\033[37m",
		DimW:      "\033[90m",
	}

	activeTheme *Theme
)

func init() {
	activeTheme = DetectTheme()
}

func DetectTheme() *Theme {
	fgBg := os.Getenv("COLORFGBG")
	if fgBg != "" {
		parts := strings.Split(fgBg, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			if bg == "white" || bg == "15" || bg == "0;15" {
				return LightTheme
			}
		}
	}

	term := os.Getenv("TERM")
	if strings.Contains(term, "light") {
		return LightTheme
	}

	return DarkTheme
}

func SetTheme(name string) {
	switch strings.ToLower(name) {
	case "light":
		activeTheme = LightTheme
	case "dark":
		activeTheme = DarkTheme
	default:
		activeTheme = DetectTheme()
	}
}

func GetTheme() *Theme {
	if activeTheme == nil {
		activeTheme = DetectTheme()
	}
	return activeTheme
}
