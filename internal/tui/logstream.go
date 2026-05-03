package tui

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
)

// logBuffer is a thread-safe circular buffer for storing log lines.
type logBuffer struct {
	mu     sync.Mutex
	lines  []string
	size   int
	offset int
}

// newLogBuffer creates a new logBuffer with the given capacity.
func newLogBuffer(size int) *logBuffer {
	return &logBuffer{lines: make([]string, 0, size), size: size}
}

// append adds a line to the buffer, overwriting the oldest line if full.
func (b *logBuffer) append(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.lines) < b.size {
		b.lines = append(b.lines, line)
	} else {
		b.lines[b.offset] = line
		b.offset = (b.offset + 1) % b.size
	}
}

// get returns a copy of the buffer's contents in chronological order.
func (b *logBuffer) get() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.lines) < b.size {
		out := make([]string, len(b.lines))
		copy(out, b.lines)
		return out
	}
	out := make([]string, b.size)
	for i := 0; i < b.size; i++ {
		out[i] = b.lines[(b.offset+i)%b.size]
	}
	return out
}

// logStreamer runs "docker compose logs -f" and stores output in a ring buffer.
type logStreamer struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	buffer *logBuffer
}

// startLogStreamer creates a new logStreamer for the given compose file.
func startLogStreamer(file string) *logStreamer {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", file, "logs", "-f", "--tail", "100")
	cmd.Stderr = io.Discard
	return &logStreamer{
		cmd:    cmd,
		cancel: cancel,
		buffer: newLogBuffer(100),
	}
}

// run starts the log command and begins reading output into the buffer.
func (ls *logStreamer) run() error {
	stdout, err := ls.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := ls.cmd.Start(); err != nil {
		return err
	}
	// Reaper: must call Wait() to avoid zombie processes.
	go func() {
		_ = ls.cmd.Wait()
	}()
	scanner := bufio.NewScanner(stdout)
	const maxCapacity = 512 * 1024 // 512KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	go func() {
		for scanner.Scan() {
			ls.buffer.append(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			ls.buffer.append("[log stream error: " + err.Error() + "]")
		}
	}()
	return nil
}

// stop terminates the log streaming process.
func (ls *logStreamer) stop() {
	if ls.cancel != nil {
		ls.cancel()
	}
	if ls.cmd != nil && ls.cmd.Process != nil {
		_ = ls.cmd.Process.Kill()
	}
}
