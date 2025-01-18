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

// messageEscaper escapes disallowed characters in workflow log messages.
var messageEscaper = strings.NewReplacer(
	"%", "%25",
	"\r", "%0D",
	"\n", "%0A",
)

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

// IsDebug reports whether the actions runner is running in
// debug mode (${{ runner.debug }}) such that logs written with [Logger.Debug] will be visible.
//
// See https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows/troubleshooting-workflows/enabling-debug-logging
func IsDebug() bool {
	return os.Getenv("RUNNER_DEBUG") == "1"
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

	fmt.Fprintf(l.out, "::debug::%s\n", messageEscaper.Replace(message))
}

// Notice writes a notice message to the workflow log.
//
// If message is the empty string "", nothing will be logged.
//
// Additionally, the caller can configure source file annotation whereby the log
// will be associated with a particular file, line, column etc. of source. This is
// done by passing in one or more [Annotation] functions that configure this behaviour.
//
// The annotations are all optional, and will only be added to the log message if they
// are explicitly set by the caller. If no annotations are passed, the log
// will simply be the message string.
func (l Logger) Notice(message string, annotations ...Annotation) {
	l.log("notice", message, annotations...)
}

// Warning writes a warning message to the workflow log.
//
// If message is the empty string "", nothing will be logged.
//
// Additionally, the caller can configure source file annotation whereby the log
// will be associated with a particular file, line, column etc. of source. This is
// done by passing in one or more [Annotation] functions that configure this behaviour.
//
// The annotations are all optional, and will only be added to the log message if they
// are explicitly set by the caller. If no annotations are passed, the log
// will simply be the message string.
func (l Logger) Warning(message string, annotations ...Annotation) {
	l.log("warning", message, annotations...)
}

// Error writes a error message to the workflow log.
//
// If message is the empty string "", nothing will be logged.
//
// Additionally, the caller can configure source file annotation whereby the log
// will be associated with a particular file, line, column etc. of source. This is
// done by passing in one or more [Annotation] functions that configure this behaviour.
//
// The annotations are all optional, and will only be added to the log message if they
// are explicitly set by the caller. If no annotations are passed, the log
// will simply be the message string.
func (l Logger) Error(message string, annotations ...Annotation) {
	l.log("error", message, annotations...)
}

// log renders an annotated message (cmd = notice | warning | error).
//
// It's behaviour is common to all annotations.
func (l Logger) log(cmd, message string, annotations ...Annotation) {
	if message == "" {
		return
	}

	// Escape the message
	message = messageEscaper.Replace(message)

	// If there are no annotations, this is just a straight message
	if len(annotations) == 0 {
		fmt.Fprintf(l.out, "::%s::%s\n", cmd, message)
		return
	}

	var ann annotation

	for _, annotation := range annotations {
		annotation.apply(&ann)
	}

	annotation := ann.String()

	// If there were no annotations after stringifying it (due to the rules for the annotations)
	// then we can just do the raw message again
	if len(annotation) == 0 {
		fmt.Fprintf(l.out, "::%s::%s\n", cmd, message)
		return
	}

	// Otherwise we need a space after ::<cmd> and the first annotation
	fmt.Fprintf(l.out, "::%s %s::%s\n", cmd, annotation, message)
}
