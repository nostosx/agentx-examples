package main

import "testing"

func TestParseReview(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text, wantDraft, wantVerdict string
		wantError                          bool
	}{
		{name: "accept", text: "A draft.\nVERDICT: OK", wantDraft: "A draft.", wantVerdict: "OK"},
		{name: "edit with CRLF", text: "First line.\nSecond line.\r\nVERDICT: EDIT\r\n", wantDraft: "First line.\nSecond line.", wantVerdict: "EDIT"},
		{name: "missing verdict", text: "A draft.", wantError: true},
		{name: "unknown verdict", text: "A draft.\nVERDICT: NOT_OK", wantError: true},
		{name: "verdict substring", text: "A draft.\nVERDICT: CREDIT", wantError: true},
		{name: "missing draft", text: "VERDICT: OK", wantError: true},
		{name: "trailing content", text: "A draft.\nVERDICT: OK\nmore", wantError: true},
		{name: "ambiguous verdict", text: "A draft.\nVERDICT: EDIT\nVERDICT: OK", wantError: true},
		{name: "inline verdict", text: "A draft. VERDICT: OK", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			draft, verdict, err := parseReview(tt.text)
			if tt.wantError {
				if err == nil || draft != "" || verdict != "" {
					t.Fatalf("accepted malformed review: %q, %q, %v", draft, verdict, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if draft != tt.wantDraft || verdict != tt.wantVerdict {
				t.Fatalf("review = (%q, %q), want (%q, %q)", draft, verdict, tt.wantDraft, tt.wantVerdict)
			}
		})
	}
}
