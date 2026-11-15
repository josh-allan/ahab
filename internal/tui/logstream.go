package tui

import (
	"bufio"
	"context"
	"os/exec"
	"slices"
	"sync"

	ahab "github.com/josh-allan/ahab/pkg"
)

// logBuffer keeps the last size lines, safe for concurrent use.
type logBuffer struct {
	mu    sync.Mutex
	lines []string
	size  int
}

func newLogBuffer(size int) *logBuffer { return &logBuffer{size: size} }

func (b *logBuffer) append(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = append(b.lines, line)
	if len(b.lines) > b.size {
		b.lines = b.lines[1:]
	}
}

func (b *logBuffer) get() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.lines)
}

// logStreamer follows `docker compose logs -f` for one stack.
type logStreamer struct {
	cancel context.CancelFunc
	buffer *logBuffer
}

func startLogStreamer(s ahab.Stack) (*logStreamer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "docker", ahab.ComposeArgs(s, "logs", "-f", "--tail", "100")...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	ls := &logStreamer{cancel: cancel, buffer: newLogBuffer(100)}
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)
		for scanner.Scan() {
			ls.buffer.append(scanner.Text())
		}
		// Wait only after stdout is drained. An exit caused by stop() is expected;
		// any other failure is shown in the pane.
		if err := cmd.Wait(); err != nil && ctx.Err() == nil {
			ls.buffer.append("[logs exited: " + err.Error() + "]")
		}
	}()
	return ls, nil
}

// stop kills the log process; CommandContext sends SIGKILL on cancel.
func (ls *logStreamer) stop() { ls.cancel() }
