package log

import "text/template"

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

// annotation is an optional attachment to certain workflow commands (notice, error, and warning)
// that adds additional metadata to the log and/or associates it with a range of source code.
//
// The fields must be exported for text/template to be able to use them.
type annotation struct {
	Title       string // The title for the annotation
	File        string // Path of the file to associate the annotation to
	StartLine   uint   // The line number (start at 1) of the start of the source range to annotate
	EndLine     uint   // The line number corresponding with the end of the annotated source
	StartColumn uint   // Start column of the annotated source range
	EndColumn   uint   // End column of the annotated source range
}

// Annotation is a log command annotation.
type Annotation interface {
	// Apply the annotation to the log message.
	//
	// Note: having the Annotation be an opaque interface completely hides all internal
	// workings from the end user, all they see is the Annotation type and the exported
	// functions. They are also nicely grouped in the godoc.
	apply(*annotation)
}

// annotator is a function that implements the Annotation interface, kind of like
// how http.HandlerFunc implements http.Handler.
type annotator func(*annotation)

// apply applies the option, implementing the Option interface for our option
// functional adapter by calling itself.
func (a annotator) apply(ann *annotation) {
	a(ann)
}

// Title adds a title to the log command annotation.
func Title(title string) Annotation {
	f := func(ann *annotation) {
		ann.Title = title
	}
	return annotator(f)
}

// File associates a source file with the annotation.
func File(file string) Annotation {
	f := func(ann *annotation) {
		ann.File = file
	}
	return annotator(f)
}

// Lines associates a span of lines in a source file with the annotation.
//
// If the annotation does not already have [File] information when Lines is called
// the line information is omitted from the annotation i.e. you can't have lines but no file.
//
// If start or end < 1, the default of 1 is used. Likewise if end < start, start
// is used as the end.
func Lines(start, end uint) Annotation {
	if start < 1 {
		// Given a bogus start line
		start = 1
	}
	if end < 1 {
		// Given a bogus end line
		end = 1
	}

	if end < start {
		// End cannot be less than start, so just call it start
		// if given a rubbish value
		end = start
	}

	f := func(ann *annotation) {
		// If there's no file, it doesn't make sense to add line info
		if ann.File == "" {
			start = 0
			end = 0
		}
		ann.StartLine = start
		ann.EndLine = end
	}

	return annotator(f)
}

// Span associates a span of columns on a single line with the annotation.
//
// If the annotation does not already have [File] information when Span is called
// the span information is omitted from the annotation i.e. you can't have span but no file.
//
// If start or end < 1, column span information will be omitted from the annotation. Likewise
// if end < start.
//
// If the caller has previously used [Lines] and the start and end line of the span are
// different, column span information is omitted from the annotation. This is an internal
// GitHub constraint.
func Span(start, end uint) Annotation {
	if start < 1 || end < 1 || end < start {
		// Given a bogus start/end column, use 0 to have span info
		// omitted from the final log
		start = 0
		end = 0
	}

	f := func(ann *annotation) {
		// If there's no file, it doesn't make sense to add span info, likewise
		// if start and end lines are different, it cannot also have column info
		if ann.File == "" || ann.StartLine != ann.EndLine {
			start = 0
			end = 0
		}

		ann.StartColumn = start
		ann.EndColumn = end
	}

	return annotator(f)
}
