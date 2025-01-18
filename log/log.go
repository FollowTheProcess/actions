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

// Logger is the actions logger, it maintains no state other than an [io.Writer]
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
// and message formatting. It is not necessary to append a final newline
// to format.
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

// StartGroup begins a new expandable group in the workflow log.
//
// Anything printed between the call to StartGroup and the call to [Logger.EndGroup] will
// be contained within this group.
//
// The caller is responsible for calling [Logger.EndGroup] after a group is started. Recommended usage is
// as follows:
//
//	func doInGroup(logger actions.Logger) {
//		logger.StartGroup()
//		defer logger.EndGroup()
//		// ...
//		// Do your grouped logic here, the group will be
//		// closed as the function returns
//	}
//
// If you want this to be handled automatically, use [Logger.WithGroup] and pass your logic as a closure.
//
// If title is the empty string "", nothing will be logged. Title will also be trimmed of
// all leading and trailing whitespace.
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#grouping-log-lines
func (l Logger) StartGroup(title string) {
	if title == "" {
		return
	}
	title = propertyEscaper.Replace(strings.TrimSpace(title))
	fmt.Fprintf(l.out, "::group::%s\n", title)
}

// EndGroup ends an expandable log group.
//
// Usage is typically deferred, see [Logger.StartGroup] for more info.
func (l Logger) EndGroup() {
	fmt.Fprintln(l.out, "::endgroup::")
}

// WithGroup executes the provided closure fn inside an expandable group in the workflow log.
//
// It automatically handles calling [Logger.StartGroup] and [Logger.EndGroup] to ensure the group
// is created and stopped correctly.
//
// Anything printed by fn will be contained within the created group. If title is the empty
// string "", nothing will be logged.
func (l Logger) WithGroup(title string, fn func()) {
	l.StartGroup(title)
	defer l.EndGroup()
	fn()
}

// Mask redacts a string or environment variable, preventing it from being printed in the workflow logs.
//
// When masked, the string or variable is replaced by `*` characters in subsequent logs. If str is
// the empty string "", nothing will be logged.
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#masking-a-value-in-a-log
//
//	logger := log.New(os.Stdout)
//	logger.Mask("$MY_SECRET") // Prevent the env var MY_SECRET from being logged
//	logger.Mask("a string") // Prevent any subsequent occurrences of "a string" from being logged
func (l Logger) Mask(str string) {
	if str == "" {
		return
	}
	str = messageEscaper.Replace(strings.TrimSpace(str))
	fmt.Fprintf(l.out, "::add-mask::%s\n", str)
}
