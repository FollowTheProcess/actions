package log_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/FollowTheProcess/actions/log"
	"github.com/FollowTheProcess/test"
)

func TestIsDebug(t *testing.T) {
	tests := []struct {
		env  map[string]string // Env vars to set for the test
		name string            // Name of the test case
		want bool              // Expected return value
	}{
		{
			name: "unset",
			env:  map[string]string{},
			want: false,
		},
		{
			name: "off",
			env: map[string]string{
				"RUNNER_DEBUG": "0", // Presumably this is what it's set to when off
			},
			want: false,
		},
		{
			name: "on",
			env: map[string]string{
				"RUNNER_DEBUG": "1",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got := log.IsDebug()
			test.Equal(t, got, tt.want)
		})
	}
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name   string // Name of the test case
		format string // Format template
		want   string // Expected output
		args   []any  // Args to be formatted
	}{
		{
			name:   "empty",
			format: "",
			args:   nil,
			want:   "",
		},
		{
			name:   "simple message",
			format: "debug log here",
			args:   nil,
			want:   "::debug::debug log here\n",
		},
		{
			name:   "formatted",
			format: "reading file: %s",
			args:   []any{"some/file.txt"},
			want:   "::debug::reading file: some/file.txt\n",
		},
		{
			name:   "formatted with escaped chars",
			format: "stuff \r\n happening here %d%% complete",
			args:   []any{42},
			want:   "::debug::stuff %0D%0A happening here 42%25 complete\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := log.New(buf)

			logger.Debug(tt.format, tt.args...)

			got := buf.String()
			test.Equal(t, got, tt.want)
		})
	}
}

func TestAnnotations(t *testing.T) {
	tests := []struct {
		name        string           // Name of the test case
		message     string           // The message to be logged
		want        string           // Expected log message
		annotations []log.Annotation // Annotations to apply to the log
	}{
		{
			name:        "empty",
			message:     "",
			annotations: nil,
			want:        "",
		},
		{
			name:        "just message",
			message:     "notice meeee",
			annotations: nil,
			want:        "::notice::notice meeee\n",
		},
		{
			name:        "message escaped",
			message:     "percent % percent % cr \r cr \r lf \n lf \n",
			annotations: nil,
			want:        "::notice::percent %25 percent %25 cr %0D cr %0D lf %0A lf %0A\n",
		},
		{
			name:    "just title",
			message: "notice meeee",
			annotations: []log.Annotation{
				log.Title("My Title"),
			},
			want: "::notice title=My Title::notice meeee\n",
		},
		{
			name:    "title escaped",
			message: "this is a notice",
			annotations: []log.Annotation{
				log.Title("Percent % crlf \r\n colon : comma ,"),
			},
			want: "::notice title=Percent %25 crlf %0D%0A colon %3A comma %2C::this is a notice\n",
		},
		{
			name:    "just file",
			message: "notice meeee",
			annotations: []log.Annotation{
				log.File("cmd/tool/main.go"),
			},
			want: "::notice file=cmd/tool/main.go::notice meeee\n",
		},
		{
			name:    "file escaped",
			message: "oh look, another notice",
			annotations: []log.Annotation{
				log.File("src/some%thing/wei\rd/who:has/colonsinfilenames"),
			},
			want: "::notice file=src/some%25thing/wei%0Dd/who%3Ahas/colonsinfilenames::oh look, another notice\n",
		},
		{
			name:    "just lines", // Doesn't make sense but I'm sure people would try it
			message: "notice meeee",
			annotations: []log.Annotation{
				log.Lines(1, 32),
			},
			want: "::notice::notice meeee\n", // Should omit line info and just do the message
		},
		{
			name:    "just span", // Also makes no sense but...
			message: "notice meeee",
			annotations: []log.Annotation{
				log.Span(1, 32),
			},
			want: "::notice::notice meeee\n", // Should omit span info and just do the message
		},
		{
			name:    "title and file",
			message: "Unexpected token '<'",
			annotations: []log.Annotation{
				log.Title("Syntax Error"),
				log.File("src/lib.rs"),
			},
			want: "::notice title=Syntax Error,file=src/lib.rs::Unexpected token '<'\n",
		},
		{
			name:    "title file and lines",
			message: "Unused import 'fmt'",
			annotations: []log.Annotation{
				log.Title("Syntax Error"),
				log.File("http/handler.go"),
				log.Lines(1, 1),
			},
			want: "::notice title=Syntax Error,file=http/handler.go,line=1,endLine=1::Unused import 'fmt'\n",
		},
		{
			name:    "full",
			message: "Your code is bad",
			annotations: []log.Annotation{
				log.Title("Look Here!"),
				log.File("src/app/handler.py"),
				log.Lines(184, 184),
				log.Span(27, 32),
			},
			want: "::notice title=Look Here!,file=src/app/handler.py,line=184,endLine=184,col=27,endColumn=32::Your code is bad\n",
		},
		{
			name:    "lines bad start",
			message: "Uh oh",
			annotations: []log.Annotation{
				log.Title("A Creative Title"),
				log.File("log/logger.go"),
				log.Lines(0, 12), // There is no line 0
			},
			want: "::notice title=A Creative Title,file=log/logger.go,line=1,endLine=12::Uh oh\n", // Should just use 1 as the start
		},
		{
			name:    "lines bad end",
			message: "Oh no!",
			annotations: []log.Annotation{
				log.Title("A Better Title"),
				log.File("cmd/dingle/main.go"),
				log.Lines(1, 0), // There is no line 0
			},
			want: "::notice title=A Better Title,file=cmd/dingle/main.go,line=1,endLine=1::Oh no!\n", // Should just use 1 as the end
		},
		{
			name:    "lines end gt start",
			message: "insert message here plz",
			annotations: []log.Annotation{
				log.Title("Star Wars"),
				log.File("src/cli.py"),
				log.Lines(37, 12), // End cannot be < Start
			},
			want: "::notice title=Star Wars,file=src/cli.py,line=37,endLine=37::insert message here plz\n", // Should just use start as end
		},
		{
			name:    "lines but no file",
			message: "where file?",
			annotations: []log.Annotation{
				log.Title("WTF"),
				log.Lines(1, 4), // Lines but where's the file!?
			},
			want: "::notice title=WTF::where file?\n", // Line info should be omitted
		},
		{
			name:    "span bad start",
			message: "naughty span",
			annotations: []log.Annotation{
				log.Title("You span me right round"),
				log.File("span/span_test.go"),
				log.Lines(1, 1), // It must have a single line for a span
				log.Span(0, 12), // There is no column 0
			},
			want: "::notice title=You span me right round,file=span/span_test.go,line=1,endLine=1::naughty span\n", // Should omit column information completely
		},
		{
			name:    "span bad end",
			message: "When will the span end",
			annotations: []log.Annotation{
				log.Title("Span? What Span?"),
				log.File("my/super/code.js"),
				log.Lines(42, 42),
				log.Span(12, 0), // Can't end on column zero
			},
			want: "::notice title=Span? What Span?,file=my/super/code.js,line=42,endLine=42::When will the span end\n", // No column info
		},
		{
			name:    "span end gt start",
			message: "wow such message",
			annotations: []log.Annotation{
				log.Title("You no math good"),
				log.File("src/request/builder.py"),
				log.Lines(128, 128),
				log.Span(17, 15), // End cannot be < Start
			},
			want: "::notice title=You no math good,file=src/request/builder.py,line=128,endLine=128::wow such message\n", // No column info
		},
		{
			name:    "span but no file",
			message: "where file?",
			annotations: []log.Annotation{
				log.Title("WTF"),
				log.Span(1, 4), // Span of what

			},
			want: "::notice title=WTF::where file?\n", // Column info should be omitted
		},
		{
			name:    "span with multiple lines",
			message: "are you mad!?",
			annotations: []log.Annotation{
				log.Title("Too Many Lines"),
				log.File("script.py"),
				log.Lines(15, 19), // Multiple lines of code - can't have a column span
				log.Span(1, 4),
			},
			want: "::notice title=Too Many Lines,file=script.py,line=15,endLine=19::are you mad!?\n", // Column info should be omitted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := log.New(buf)

			logger.Notice(tt.message, tt.annotations...)

			got := buf.String()
			test.Diff(t, got, tt.want)
		})
	}
}

func TestWarning(t *testing.T) {
	// Most of the logic is tested in the big test above, we really just want to see ::warning
	buf := &bytes.Buffer{}
	logger := log.New(buf)

	logger.Warning("This is dangerous", log.Title("Be Careful!"), log.File("src/main.py"), log.Lines(2, 6))

	got := buf.String()
	want := "::warning title=Be Careful!,file=src/main.py,line=2,endLine=6::This is dangerous\n"
	test.Diff(t, got, want)
}

func TestError(t *testing.T) {
	// Same... but ::error
	buf := &bytes.Buffer{}
	logger := log.New(buf)

	logger.Error("This is broken", log.Title("Syntax Error"), log.File("src/main.py"), log.Lines(2, 6))

	got := buf.String()
	want := "::error title=Syntax Error,file=src/main.py,line=2,endLine=6::This is broken\n"
	test.Diff(t, got, want)
}

func BenchmarkLog(b *testing.B) {
	logger := log.New(io.Discard)

	b.ResetTimer()
	for range b.N {
		logger.Notice("Hello", log.Title("A Title"), log.File("src/main.rs"), log.Lines(1, 18))
	}
}
