package log

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
		ann.StartLine = start
		ann.EndLine = end
	}

	return annotator(f)
}

// Span associates a span of columns on a single line with the annotation.
//
// If start or end < 1, column span information will be omitted from the annotation. Likewise
// if end < start.
//
// If the caller has previously used [Lines] and the start and end line of the span are
// different, column span information is omitted from the annotation. This is an internal
// GitHub constraint.
func Span(start, end uint) Annotation {
	if start < 1 {
		// Given a bogus start column, use 0 to have it
		// omitted from the final log
		start = 0
	}
	if end < 1 {
		// Given a bogus end line, same
		end = 0
	}

	if end < start {
		// End cannot be less than start, so just call it start
		// if given a rubbish value
		end = start
	}

	f := func(ann *annotation) {
		// If the annotation has line info and the start and end lines
		// are different, it cannot also have column information
		if ann.StartLine != ann.EndLine {
			start = 0
			end = 0
		}

		ann.StartColumn = start
		ann.EndColumn = end
	}

	return annotator(f)
}
