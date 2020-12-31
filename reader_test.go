package igotify

import (
	"errors"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

const filename = "test123.txt"

func TestReader_Listen_StopWithoutError(t *testing.T) {
	r, err := NewReader(32, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.AddWatcher(".", syscall.IN_CREATE); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := r.Listen(); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	r.Stop()
	wg.Wait()
}

func TestReader_Listen_StopWithoutError_WithNotify(t *testing.T) {
	r, err := NewReader(32, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.AddWatcher(".", syscall.IN_CREATE); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := r.Listen(); err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		defer os.Remove(filename)
		create, err := os.Create(filename)
		if err != nil {
			t.Fatal(err)
		}
		create.Close()
	}()
	gotEvent := false
	go func() {
		event, err := r.Get()
		if err != nil {
			t.Errorf("cannot get event: %v", err)
			return
		}
		if event.Name != filename {
			t.Errorf("expected event for %s, got %s", filename, event.Name)
			return
		}
		gotEvent = true
	}()

	time.Sleep(100 * time.Millisecond)
	r.Stop()
	wg.Wait()
	if !gotEvent {
		t.Error("expected event, got none")
	}
}

func TestReader_Listen_Get(t *testing.T) {
	r, err := NewReader(32, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.AddWatcher(".", syscall.IN_CREATE|syscall.IN_DELETE); err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := r.Listen(); err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		defer os.Remove(filename)
		create, err := os.Create(filename)
		if err != nil {
			t.Fatal(err)
		}
		create.Close()
	}()
	gotEvent := false
	go func() {
		event, err := r.Get()
		if err != nil {
			t.Errorf("cannot get event: %v", err)
			return
		}
		if event.Name != filename {
			t.Errorf("expected event for %s, got %s", filename, event.Name)
			return
		}
		if event.Mask&syscall.IN_CREATE == 0 {
			t.Errorf("expected event type %s, got %s", IN_CREATE, maskString(event.Mask))
			return
		}
		event, err = r.Get()
		if err != nil {
			t.Errorf("cannot get event: %v", err)
			return
		}
		t.Logf("got event %s", event)
		if event.Name != filename {
			t.Errorf("expected event for %s, got %s", filename, event.Name)
			return
		}
		if event.Mask&syscall.IN_DELETE == 0 {
			t.Errorf("expected event type %s, got %s", IN_DELETE, maskString(event.Mask))
			return
		}
		gotEvent = true
	}()

	time.Sleep(100 * time.Millisecond)
	r.Stop()
	if !gotEvent {
		t.Error("expected event, got none")
	}
}

func TestReader_AddWatcher(t *testing.T) {
	r, err := NewReader(128, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}
	go r.Listen()

	go func() {
		file, err := os.Create(filename)
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}()

	_, err = r.GetWithTimeout(100 * time.Millisecond)
	if err != nil {
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("expected timeout, got: %v", err)
		}
	}

	_, err = r.AddWatcher(".", syscall.IN_CREATE|syscall.IN_DELETE)
	if err != nil {
		t.Fatal(err)
	}

	if err = os.Remove(filename); err != nil {
		t.Fatalf("cannot remove file: %v", err)
	}

	event, err := r.GetWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if event.Name != filename || event.Mask&syscall.IN_DELETE == 0 {
		t.Fatalf("unexpected event %s", event)
	}
}

func TestReader_RemoveWatcher(t *testing.T) {
	r, err := NewReader(128, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}
	wd, err := r.AddWatcher(".", syscall.IN_CREATE|syscall.IN_DELETE)
	if err != nil {
		t.Fatal(err)
	}
	go r.Listen()

	go func() {
		file, err := os.Create(filename)
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}()

	event, err := r.GetWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if event.Name != filename || event.Mask&syscall.IN_CREATE == 0 {
		t.Fatal("unexpected event")
	}

	if err = r.RemoveWatcher(wd); err != nil {
		t.Errorf("cannot remove watcher: %v", err)
	}
	if err = os.Remove(filename); err != nil {
		t.Fatalf("cannot remove file: %v", err)
	}

	event, err = r.GetWithTimeout(100 * time.Millisecond)
	if err != nil {
		if errors.Is(err, ErrTimeout) {
			return
		}
		t.Error(err)
	} else {
		if event.Mask&syscall.IN_IGNORED != 0 {
			return
		}
		t.Errorf("expected an error, got an event %s", event)
	}
}

func TestReader_GetWithTimeout_Normal(t *testing.T) {
	r, err := NewReader(128, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.AddWatcher(".", syscall.IN_CREATE)
	if err != nil {
		t.Fatal(err)
	}
	go r.Listen()

	go func() {
		defer os.Remove(filename)
		file, err := os.Create(filename)
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}()

	event, err := r.GetWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if event.Name != filename || event.Mask&syscall.IN_CREATE == 0 {
		t.Errorf("unexpected event %s", event)
	}
}

func TestReader_GetWithTimeout_Timeout(t *testing.T) {
	r, err := NewReader(128, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}
	go r.Listen()

	go func() {
		defer os.Remove(filename)
		file, err := os.Create(filename)
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}()

	event, err := r.GetWithTimeout(100 * time.Millisecond)
	if err == nil {
		t.Errorf("expected an error, got an event %s", event)
	} else if !errors.Is(err, ErrTimeout) {
		t.Error(err)
	}
}

func TestReader_Listen_Repeated(t *testing.T) {
	r, err := NewReader(128, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}

	resultChan := make(chan bool, 1)

	go func() {
		resultChan <- true
		if err := r.Listen(); err != nil {
			t.Fatalf("cannot begin listen: %v", err)
		}
	}()
	defer r.Stop()

	<-resultChan

	go func() {
		if err := r.Listen(); err != nil {
			if errors.Is(err, ErrListening) {
				resultChan <- true
				return
			} else {
				t.Fatalf("expected error %v, got %v", ErrListening, err)
			}
		} else {
			t.Fatal("expected error from repeated listen, got nil")
		}
	}()

	select {
	case <-resultChan:
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatal("repeated listen got through, but should've failed")
	}
}

func TestReader_Listen_AfterStop(t *testing.T) {
	r, err := NewReader(128, DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}

	resultChan := make(chan bool, 1)

	go func() {
		resultChan <- true
		if err := r.Listen(); err != nil {
			t.Fatal(err)
		}
	}()
	<-resultChan
	r.Stop()

	go func() {
		if err = r.Listen(); err != nil {
			if errors.Is(err, ErrStopped) {
				resultChan <- true
				return
			} else {
				t.Fatalf("expected error %v, got %v", ErrStopped, err)
			}
		} else {
			t.Fatal("expected Listen() to return an error, got nil")
		}
	}()

	select {
	case <-resultChan:
		return
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout while waiting for test to finish")
	}
}
