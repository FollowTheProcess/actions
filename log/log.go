// Package log implements functionality to write workflow log commands from
// within a GitHub action.
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions
package log

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	// properties in workflow commands (e.g. file, startLine etc.) must be
	// escaped according to these rules.
	propertyEscaper = strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
		":", "%3A",
		",", "%2C",
	)
	// data (e.g. file=<data>) must also be escaped, but less strict here.
	dataEscaper = strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
	)
)

// IsDebug reports whether the actions runner is running in
// debug mode (${{ runner.debug }}) such that logs written with [Debug] will be visible.
//
// See https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows/troubleshooting-workflows/enabling-debug-logging
func IsDebug() bool {
	return os.Getenv("RUNNER_DEBUG") == "1"
}

// Logger is the actions logger, it maintains no state other than and [io.Writer]
// which is where the logs will be printed.
type Logger struct {
	out io.Writer
}

// New returns a new [Logger] configured to write to out.
//
// Correct usage in GitHub Actions sets out to [os.Stdout], but specifying
// the writer can be handy for unit tests in your action code.
//
//	logger := log.New(os.Stdout)
func New(out io.Writer) Logger {
	return Logger{out: out}
}

// Debug writes a formatted debug message to the workflow log.
//
// The signature is analogous to [fmt.Printf] allowing format verbs
// and message formatting.
//
// If the format arguments are omitted, format will be treated as
// a verbatim string and passed straight through.
//
// This will only be seen if $RUNNER_DEBUG (or ${{ runner.debug }}) is set
// which can be controlled by the person running the action.
//
// Generally, this is done after a failed run with the "run with debug logs"
// option enabled.
//
// See https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows/troubleshooting-workflows/enabling-debug-logging
func (l Logger) Debug(format string, a ...any) {
	// GitHub gets confused when workflow commands are empty
	if format == "" {
		return
	}
	var message string
	if len(a) == 0 {
		// No data to format, treat it as a simple string
		message = format
	} else {
		// We need to do formatting
		message = fmt.Sprintf(format, a...)
	}

	fmt.Fprintf(l.out, "::debug::%s\n", dataEscaper.Replace(message))
}
