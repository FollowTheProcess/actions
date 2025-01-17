// Package log implements functionality to write workflow log commands from
// within a GitHub action.
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions
package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

// annotationTemplate is the text/template syntax for creating a log annotation.
//
// It's fairly simple, it proceeds in order through the annotation checking each field:
//   - If it is the zero value, it won't print the key=value for that field
//   - Whether the field before it was the zero value, in which case the previous field wasn't printed and it
//     wont insert a comma between them
//
// It also calls the two escape functions to sanitise the text for GitHub to render.
const annotationTemplate = `{{- if ne .Title "" }}title={{ .Title | escapeProperty }}{{ end -}}
{{- if ne .File "" }}{{ if ne .Title "" }},{{ end }}file={{ .File | escapeProperty }}{{ end -}}
{{- if ne .StartLine 0 }}{{ if ne .File "" }},{{ end }}line={{ .StartLine }}{{ end -}}
{{- if ne .EndLine 0 }}{{ if ne .StartLine 0 }},{{ end }}endLine={{ .EndLine }}{{ end -}}
{{- if ne .StartColumn 0 }}{{ if ne .EndLine 0 }},{{ end }}col={{ .StartColumn }}{{ end -}}
{{- if ne .EndColumn 0 }}{{ if ne .StartColumn 0}},{{ end }}endColumn={{ .EndColumn }}{{ end -}}`

var funcMap = template.FuncMap{
	"escapeProperty": propertyEscaper.Replace,
	"escapeData":     messageEscaper.Replace,
}

var templ = template.Must(template.New("annotation").Funcs(funcMap).Parse(annotationTemplate))

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
	// log messages must also be escaped.
	messageEscaper = strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
	)
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
// debug mode (${{ runner.debug }}) such that logs written with [Debug] will be visible.
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
// Additionally, the caller can configure source file annotation whereby the notice
// message will be associated with a particular file, line, column etc. of source. This is
// done by passing in one or more [Annotation] functions that configure this behaviour.
//
// The annotations are all optional, and will only be added to the log message if they
// are explicitly set by the caller. If no annotations are passed, the notice log
// will simply be the message string.
func (l Logger) Notice(message string, annotations ...Annotation) {
	if message == "" {
		return
	}

	// Escape the message
	message = messageEscaper.Replace(message)

	// If there are no annotations, this is just a straight message
	if len(annotations) == 0 {
		fmt.Fprintf(l.out, "::notice::%s\n", message)
		return
	}

	var ann annotation

	for _, annotation := range annotations {
		annotation.apply(&ann)
	}

	buf := &bytes.Buffer{}
	if err := templ.Execute(buf, ann); err != nil {
		// TODO(@FollowTheProcess): What do here? I guess just do the message with no annotations
		fmt.Println("Uh oh!", err)
		return
	}

	fmt.Fprintf(l.out, "::notice %s::%s\n", buf.String(), message)
}
