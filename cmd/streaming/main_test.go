package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestStreamingProcessStopsWithIdleInput(t *testing.T) {
	const helperEnv = "AGENTX_STREAMING_TEST_PROCESS"
	if os.Getenv(helperEnv) == "1" {
		err := run()
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if runtime.GOOS == "windows" {
		t.Skip("SIGTERM is not supported by Windows process signaling")
	}
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestStreamingProcessStopsWithIdleInput$")
	cmd.Env = append(os.Environ(), helperEnv+"=1", "AGENTX_PROVIDER=openai", "OPENAI_API_KEY=fixture")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	prompt := make([]byte, 2)
	if _, err := io.ReadFull(output, prompt); err != nil || string(prompt) != "> " {
		t.Fatalf("prompt=%q error=%v", prompt, err)
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	// Keep the writer open so EOF cannot hide a blocked stdin read.
	if err := cmd.Wait(); err != nil {
		t.Fatalf("exit=%v stderr=%s", err, &stderr)
	}
}
