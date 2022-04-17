package logrus

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unicode/utf8"
)

const JWDefaultTimeFormat = "2006-01-02 15:04:05,000"

type JWFormatter struct {
	ForceColors bool

	// Force disabling colors.
	DisableColors bool

	// Override coloring based on CLICOLOR and CLICOLOR_FORCE. - https://bixense.com/clicolors/
	EnvironmentOverrideColors bool

	// TimestampFormat to use for display when a full timestamp is printed.
	// The format to use is the same than for time.Format or time.Parse from the standard
	// library.
	// The standard Library already provides a set of predefined format.
	TimestampFormat string

	// PadLevelText Adds padding the level text so that all the levels output at the same length
	// PadLevelText is a superset of the DisableLevelTruncation option
	PadLevelText bool

	// Whether the logger's out is to a terminal
	isTerminal bool

	// CallerPrettyfier can be set by the user to modify the content
	// of the function and file keys in the data when ReportCaller is
	// activated. If any of the returned value is the empty string the
	// corresponding key will be removed from fields.
	CallerPrettyfier func(*runtime.Frame) (function string, file string)

	terminalInitOnce sync.Once

	// The max length of the level text, generated dynamically on init
	levelTextMaxLength int
}

func (f *JWFormatter) init(entry *Entry) {
	if entry.Logger != nil {
		f.isTerminal = checkIfTerminal(entry.Logger.Out) || checkStdOut(entry.Logger.Out)
	}
	// Get the max length of the level text
	for _, level := range AllLevels {
		levelTextLength := utf8.RuneCount([]byte(level.String()))
		if levelTextLength > f.levelTextMaxLength {
			f.levelTextMaxLength = levelTextLength
		}
	}
}

func (f *JWFormatter) Format(entry *Entry) ([]byte, error) {

	f.terminalInitOnce.Do(func() {
		f.init(entry)
	})
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.formatOutput(b, entry, f.isColored())

	b.WriteString("\n")
	return b.Bytes(), nil

}

func (f *JWFormatter) isColored() bool {
	isColored := f.ForceColors || (f.isTerminal && (runtime.GOOS != "windows"))

	if f.EnvironmentOverrideColors {
		switch force, ok := os.LookupEnv("CLICOLOR_FORCE"); {
		case ok && force != "0":
			isColored = true
		case ok && force == "0", os.Getenv("CLICOLOR") == "0":
			isColored = false
		}
	}

	return isColored && !f.DisableColors
}

func checkStdOut(w io.Writer) bool {
	switch f := w.(type) {
	case *os.File:
		return f == os.Stdout || f == os.Stderr
	default:
		return false
	}
}

func (f *JWFormatter) formatOutput(b *bytes.Buffer, entry *Entry, withColor bool) {
	var levelColor int
	switch entry.Level {
	case DebugLevel, TraceLevel:
		levelColor = blue
	case WarnLevel:
		levelColor = yellow
	case ErrorLevel, FatalLevel, PanicLevel:
		levelColor = red
	case InfoLevel:
		//levelColor = green
	default:
		levelColor = blue
	}
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = JWDefaultTimeFormat
	}

	caller := ""
	if entry.HasCaller() {

		var (
			funcVal, fileVal string
		)

		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		} else {
			strs := strings.Split(entry.Caller.Function, ".")
			funcVal = strs[len(strs)-1]
			funcVal = fmt.Sprintf("%s", funcVal)
			fileVal = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
		}
		caller = fmt.Sprintf("%s [%s]", fileVal, funcVal)
	}
	if withColor {
		fmt.Fprintf(b, "\x1b[%dm[%s] %s %s %s", levelColor, entry.Level, entry.Time.Format(timestampFormat), caller, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] %s %s %s", entry.Level, entry.Time.Format(timestampFormat), caller, entry.Message)
	}

}
