package streaming

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/model"
)

func TestChatRetainsConversation(t *testing.T) {
	t.Parallel()
	calls := 0
	m := model.Func(func(ctx context.Context, req *model.Request) (*model.Response, error) {
		if _, ok := ctx.Deadline(); !ok {
			t.Error("turn has no deadline")
		}
		calls++
		var got []string
		for _, message := range req.Messages {
			got = append(got, string(message.Role)+":"+message.Text())
		}
		want := []string{"user:first"}
		if calls == 2 {
			want = append(want, "assistant:answer", "user:second")
		}
		if !slices.Equal(got, want) {
			t.Errorf("turn %d messages = %q, want %q", calls, got, want)
		}
		return &model.Response{Message: model.AssistantText("answer")}, nil
	})
	a := agent.MustNew("chat", agent.WithModel(m), agent.WithStreaming())
	var output bytes.Buffer
	if err := chat(t.Context(), agent.NewRunner(a, nil), strings.NewReader("\nfirst\nsecond\n/exit\nignored\n"), &output); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || strings.Count(output.String(), "answer") != 2 {
		t.Fatalf("calls=%d output=%q", calls, output.String())
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestChatInputFailure(t *testing.T) {
	t.Parallel()
	failure := errors.New("input failed")
	for _, tt := range []struct {
		name  string
		input io.Reader
		want  error
	}{
		{name: "read failure", input: failingReader{err: failure}, want: failure},
		{name: "oversized line", input: strings.NewReader(strings.Repeat("x", 70<<10))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := model.Func(func(context.Context, *model.Request) (*model.Response, error) {
				t.Error("invalid input reached the model")
				return nil, errors.New("unexpected model call")
			})
			a := agent.MustNew("chat", agent.WithModel(m), agent.WithStreaming())
			err := chat(t.Context(), agent.NewRunner(a, nil), tt.input, io.Discard)
			if err == nil || (tt.want != nil && !errors.Is(err, tt.want)) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestChatModelFailure(t *testing.T) {
	t.Parallel()
	failure := errors.New("model failed")
	m := model.Func(func(context.Context, *model.Request) (*model.Response, error) { return nil, failure })
	a := agent.MustNew("chat", agent.WithModel(m), agent.WithStreaming())
	if err := chat(t.Context(), agent.NewRunner(a, nil), strings.NewReader("hello\n"), io.Discard); !errors.Is(err, failure) {
		t.Fatalf("error = %v", err)
	}
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

func TestChatOutputFailureStopsConversation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, failOn          string
		occurrence, wantCalls int
	}{
		{name: "prompt", failOn: "> ", occurrence: 1, wantCalls: 0},
		{name: "delta", failOn: "answer", occurrence: 1, wantCalls: 1},
		{name: "summary", failOn: "output tokens]", occurrence: 1, wantCalls: 1},
		{name: "next prompt", failOn: "> ", occurrence: 2, wantCalls: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			m := model.Func(func(context.Context, *model.Request) (*model.Response, error) {
				calls++
				return &model.Response{Message: model.AssistantText("answer")}, nil
			})
			a := agent.MustNew("chat", agent.WithModel(m), agent.WithStreaming())
			failure := errors.New("output failed")
			var written strings.Builder
			output := writerFunc(func(p []byte) (int, error) {
				written.Write(p)
				if strings.Count(written.String(), tt.failOn) >= tt.occurrence {
					return 0, failure
				}
				return len(p), nil
			})
			err := chat(t.Context(), agent.NewRunner(a, nil), strings.NewReader("first\nsecond\n"), output)
			if !errors.Is(err, failure) {
				t.Fatalf("error = %v, want %v", err, failure)
			}
			if calls != tt.wantCalls {
				t.Fatalf("calls = %d, want %d; output = %q", calls, tt.wantCalls, written.String())
			}
		})
	}
}

func TestChatTurnDeadline(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		m := model.Func(func(ctx context.Context, _ *model.Request) (*model.Response, error) {
			<-ctx.Done()
			return nil, context.Cause(ctx)
		})
		a := agent.MustNew("chat", agent.WithModel(m), agent.WithStreaming())
		start := time.Now()
		err := chat(t.Context(), agent.NewRunner(a, nil), strings.NewReader("hello\n"), io.Discard)
		if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) != 5*time.Minute {
			t.Fatalf("elapsed=%v error=%v", time.Since(start), err)
		}
	})
}

func TestChatCancellation(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(t.Context())
		failure := errors.New("chat canceled")
		m := model.Func(func(ctx context.Context, _ *model.Request) (*model.Response, error) {
			cancel(failure)
			<-ctx.Done()
			return nil, context.Cause(ctx)
		})
		a := agent.MustNew("chat", agent.WithModel(m), agent.WithStreaming())
		if err := chat(ctx, agent.NewRunner(a, nil), strings.NewReader("hello\n"), io.Discard); !errors.Is(err, failure) {
			t.Fatalf("error=%v", err)
		}
	})
}
