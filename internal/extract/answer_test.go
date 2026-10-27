package extract

import "testing"

func TestParseAnswer(t *testing.T) {
	tests := []struct {
		input  string
		want   string
		wantOK bool
	}{
		{
			input:  "ANSWER: yes",
			want:   "yes",
			wantOK: true,
		},
		{
			input:  "**ANSWER:** yes",
			want:   "yes",
			wantOK: true,
		},
		{
			input:  "**ANSWER**: **yes**",
			want:   "yes",
			wantOK: true,
		},
		{
			input:  "*ANSWER:* yes",
			want:   "yes",
			wantOK: true,
		},
		{
			input:  "ANSWER:",
			want:   "",
			wantOK: true,
		},
		{
			input:  "Some discussion\n**ANSWER:** 42",
			want:   "42",
			wantOK: true,
		},
		{
			input:  "No answer line here",
			want:   "",
			wantOK: false,
		},
		{
			input:  "QUESTION: what is it?",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		got, ok := ParseAnswer(tt.input)
		if ok != tt.wantOK || got != tt.want {
			t.Errorf("ParseAnswer(%q) = (%q, %v), want (%q, %v)", tt.input, got, ok, tt.want, tt.wantOK)
		}
	}
}
