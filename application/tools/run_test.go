package tools

import (
	"fmt"
	"math"
	"slices"
	"sync"
	"testing"
)

func TestCalculate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		input     calcArgs
		want      string
		wantError bool
	}{
		{name: "addition", input: calcArgs{Op: "add", A: 2, B: 3}, want: "5"},
		{name: "subtraction", input: calcArgs{Op: "sub", A: 2, B: 3}, want: "-1"},
		{name: "multiplication", input: calcArgs{Op: "mul", A: 2, B: 3}, want: "6"},
		{name: "division", input: calcArgs{Op: "div", A: 3, B: 2}, want: "1.5"},
		{name: "power", input: calcArgs{Op: "pow", A: 2, B: 10}, want: "1024"},
		{name: "division by zero", input: calcArgs{Op: "div", A: 1}, wantError: true},
		{name: "unknown operation", input: calcArgs{Op: "unknown"}, wantError: true},
		{name: "overflow", input: calcArgs{Op: "mul", A: math.MaxFloat64, B: 2}, wantError: true},
		{name: "undefined power", input: calcArgs{Op: "pow", A: -1, B: 0.5}, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := calculate(t.Context(), tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if got.IsError != tt.wantError || (!tt.wantError && got.Content != tt.want) {
				t.Fatalf("result = %+v", got)
			}
		})
	}
}

func TestNoteStoreConcurrentSave(t *testing.T) {
	t.Parallel()
	var store noteStore
	var wg sync.WaitGroup
	var want []string
	for i := range 64 {
		want = append(want, fmt.Sprint(i))
		wg.Go(func() {
			if _, err := store.save(t.Context(), noteArgs{Text: fmt.Sprint(i)}); err != nil {
				t.Error(err)
			}
			_ = store.snapshot()
		})
	}
	wg.Wait()
	notes := store.snapshot()
	slices.Sort(notes)
	slices.Sort(want)
	if !slices.Equal(notes, want) {
		t.Fatalf("notes = %q, want %q", notes, want)
	}
	notes[0] = "modified"
	if slices.Contains(store.snapshot(), "modified") {
		t.Fatal("snapshot aliases stored notes")
	}
}
