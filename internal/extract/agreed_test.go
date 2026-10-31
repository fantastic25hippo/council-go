package extract

import "testing"

func TestIsAgreed(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty string", input: "", want: false},
		{name: "whitespace only", input: "   \t\n  \r\n", want: false},
		{name: "plain agreed", input: "AGREED", want: true},
		{name: "lowercase agreed", input: "agreed", want: true},
		{name: "mixed case agreed", input: "AgReEd", want: true},
		{name: "agreed with period", input: "AGREED.", want: true},
		{name: "agreed with spaces around", input: "  AGREED  ", want: true},
		{name: "agreed bold markdown", input: "**AGREED**", want: true},
		{name: "agreed italic markdown", input: "*AGREED*", want: true},
		{name: "agreed underscore italic", input: "_AGREED_", want: true},
		{name: "agreed backticks", input: "`AGREED`", want: true},
		{name: "bold with period inside", input: "**AGREED.**", want: true},
		{name: "bold with period outside", input: "**AGREED**.", want: true},
		{name: "markdown mixed with period", input: "_**AGREED**._", want: true},
		{name: "plan with agreed last line", input: "plan details\nLGTM\nAGREED", want: true},
		{name: "trailing blank lines after agreed", input: "AGREED\n\n\n", want: true},
		{name: "crlf line endings", input: "step 1\r\nstep 2\r\n**AGREED**\r\n", want: true},
		{name: "not agreed last", input: "AGREED\nmore discussion here", want: false},
		{name: "not agreed text", input: "not agreed", want: false},
		{name: "explicit NOT AGREED", input: "NOT AGREED", want: false},
		{name: "agreed within sentence", input: "We are agreed on this", want: false},
		{name: "agreed with punctuation question mark", input: "AGREED?", want: false},
		{name: "agreed with exclamation mark", input: "AGREED!", want: false},
		{name: "agreed with prefix word", input: "DISAGREED", want: false},
		{name: "agreed with multiple periods", input: "AGREED..", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAgreed(tt.input)
			if got != tt.want {
				t.Errorf("IsAgreed(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseLaunch(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "no launch line", input: "some notes\nAGREED", want: ""},
		{name: "simple launch", input: "LAUNCH: local", want: "local"},
		{name: "launch with worker key", input: "LAUNCH: worker=jules", want: "jules"},
		{name: "launch worker uppercase syntax", input: "LAUNCH WORKER: agy", want: "agy"},
		{name: "case insensitivity in prefix", input: "launch: local", want: "local"},
		{name: "launch worker case insensitivity", input: "launch worker: Jules", want: "jules"},
		{name: "worker with dashes and underscores", input: "LAUNCH: custom-worker_01", want: "custom-worker_01"},
		{name: "bold markdown wrapped", input: "**LAUNCH: local**", want: "local"},
		{name: "whitespace surrounding line and values", input: "  \tLAUNCH:   worker = local  \n", want: "local"},
		{name: "multi-line takes first match", input: "LAUNCH: first\nLAUNCH: second", want: "first"},
		{name: "multi-line with context", input: "PLAN\nLAUNCH: jules\nAGREED", want: "jules"},
		{name: "non-matching prefix", input: "DO NOT LAUNCH: local", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLaunch(tt.input)
			if got != tt.want {
				t.Errorf("ParseLaunch(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
