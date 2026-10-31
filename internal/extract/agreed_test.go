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
		{name: "plain agreed uppercase", input: "AGREED", want: true},
		{name: "agreed lowercase", input: "agreed", want: true},
		{name: "agreed mixed case", input: "AgReEd", want: true},
		{name: "agreed with trailing dot", input: "AGREED.", want: true},
		{name: "bold agreed", input: "**AGREED**", want: true},
		{name: "bold agreed with trailing dot inside", input: "**AGREED.**", want: true},
		{name: "italic agreed with dot", input: "*AGREED.*", want: true},
		{name: "backticks agreed", input: "`AGREED`", want: true},
		{name: "underscore agreed", input: "_AGREED_", want: true},
		{name: "mixed cutset and whitespace", input: "  * _` AGREED `_ *  ", want: true},
		{name: "multi-line agreed on last line", input: "Plan details\nLooks good\nAGREED", want: true},
		{name: "multi-line with trailing whitespace and empty lines", input: "Plan\nAGREED\n   \n\n", want: true},
		{name: "multi-line with crlf endings", input: "Line 1\r\nLine 2\r\nAGREED\r\n", want: true},
		{name: "not last line", input: "AGREED\nThen some more comments", want: false},
		{name: "not agreed phrase", input: "NOT AGREED", want: false},
		{name: "disagreed", input: "DISAGREED", want: false},
		{name: "partial match within sentence", input: "I have AGREED to this plan", want: false},
		{name: "question mark", input: "AGREED?", want: false},
		{name: "multiple dots", input: "AGREED..", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAgreed(tt.input); got != tt.want {
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
		{name: "no launch line", input: "Some discussion\nAGREED", want: ""},
		{name: "standard launch local", input: "LAUNCH: local", want: "local"},
		{name: "launch with worker equals", input: "LAUNCH: worker=jules", want: "jules"},
		{name: "launch with spaces around equals", input: "LAUNCH: worker = local", want: "local"},
		{name: "launch worker format", input: "LAUNCH WORKER: agy", want: "agy"},
		{name: "launch worker with worker equals", input: "LAUNCH WORKER: worker=jules", want: "jules"},
		{name: "case insensitive launch prefix", input: "launch: local", want: "local"},
		{name: "worker name with hyphen and underscore", input: "LAUNCH: custom-worker_1", want: "custom-worker_1"},
		{name: "worker name uppercase normalized to lowercase", input: "LAUNCH: JULES", want: "jules"},
		{name: "markdown bold launch", input: "**LAUNCH: local**", want: "local"},
		{name: "markdown italic launch", input: "*LAUNCH: worker=agy*", want: "agy"},
		{name: "launch in multi-line text", input: "Plan discussion\nLAUNCH: local\nAGREED", want: "local"},
		{name: "multiple launch lines returns first", input: "LAUNCH: local\nLAUNCH: jules", want: "local"},
		{name: "relaunch should not match", input: "RELAUNCH: local", want: ""},
		{name: "launch missing worker identifier", input: "LAUNCH:", want: ""},
		{name: "launch worker missing identifier", input: "LAUNCH WORKER:", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLaunch(tt.input); got != tt.want {
				t.Errorf("ParseLaunch(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
