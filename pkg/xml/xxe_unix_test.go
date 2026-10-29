//go:build unix

package xml

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestValidate_XXEDoesNotOpenFiles points external entities at a FIFO. Opening
// a FIFO for reading blocks until a writer appears, so a writer goroutine
// completing its own open means something opened the path for reading.
func TestValidate_XXEDoesNotOpenFiles(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "trap")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skip("mkfifo:", err)
	}

	// Control: a plain read of the FIFO must trip the trap, or the test below
	// would pass vacuously.
	if !trapTripped(fifo, func() { _, _ = os.ReadFile(fifo) }) {
		t.Fatal("control failed: reading the FIFO did not complete the writer's open")
	}

	opened := make(chan struct{})
	go func() {
		// Blocks until a reader opens the FIFO.
		if w, err := os.OpenFile(fifo, os.O_WRONLY, 0); err == nil {
			close(opened)
			_ = w.Close()
		}
	}()
	// Release the writer goroutine when the test ends.
	defer func() {
		if r, err := os.OpenFile(fifo, os.O_RDONLY|syscall.O_NONBLOCK, 0); err == nil {
			_ = r.Close()
		}
	}()

	docs := []string{
		`<!DOCTYPE d [<!ENTITY x SYSTEM "file://` + fifo + `">]><d>&x;</d>`,
		`<!DOCTYPE d [<!ENTITY x SYSTEM "file://` + fifo + `">]><d>ok</d>`,
		`<!DOCTYPE d [<!ENTITY % p SYSTEM "` + fifo + `"> %p;]><d/>`,
		`<!DOCTYPE d SYSTEM "` + fifo + `"><d/>`,
	}
	for _, doc := range docs {
		_ = Validate(doc)
	}

	select {
	case <-opened:
		t.Fatal("Validate opened the file named by an external entity or DTD")
	case <-time.After(200 * time.Millisecond):
	}
}

// trapTripped runs read while a writer waits on the FIFO and reports whether the
// writer's open completed.
func trapTripped(fifo string, read func()) bool {
	done := make(chan struct{})
	go func() {
		if w, err := os.OpenFile(fifo, os.O_WRONLY, 0); err == nil {
			_ = w.Close()
		}
		close(done)
	}()
	readDone := make(chan struct{})
	go func() { read(); close(readDone) }()
	select {
	case <-done:
		<-readDone // let the reader finish so it cannot satisfy a later open
		return true
	case <-time.After(2 * time.Second):
		return false
	}
}
