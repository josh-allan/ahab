package tui

import "testing"

func Test_logBuffer(t *testing.T) {
	b := newLogBuffer(3)
	b.append("a")
	b.append("b")
	b.append("c")
	got := b.get()
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("len=%d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("get()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	b.append("d")
	got = b.get()
	want = []string{"b", "c", "d"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("after overflow get()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
