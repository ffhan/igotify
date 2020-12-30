package igotify

import (
	"errors"
	"sync"
	"syscall"
	"unsafe"
)

const (
	DefaultBufferSize = 128
	DefaultFlags      = 0
)

var (
	ErrStopped = errors.New("reader is stopped")
)

// reader is a wrapper for reading and controlling inotify events.
type reader struct {
	fd         int               // inotify file descriptor
	wds        map[uint32]bool   // watch descriptors
	wdMutex    sync.Mutex        // watch descriptor mutex
	bufferSize int               // read buffer size (in sizeof(inotify_event))
	resultChan chan InotifyEvent //
	stopped    bool              // indicates whether the fd and wds are closed
}

// create a new reader with a specified buffer size.
//
// bufferSize does not define the max number of inotify events in a buffer because their length is variable due to event names.
//
// bufferSize defines the number of bytes of a read buffer: sizeof(inotify_event) * bufferSize
//
// default flags value is 0.
func NewReader(bufferSize int, flags int) (*reader, error) {
	fd, err := syscall.InotifyInit1(flags)
	if err != nil {
		return nil, err
	}
	return &reader{
		fd:         fd,
		bufferSize: bufferSize,
		resultChan: make(chan InotifyEvent, bufferSize),
		wds:        make(map[uint32]bool),
		stopped:    false,
	}, nil
}

// retrieves a single InotifyEvent, blocks until a read can be completed.
//
// returns an error if a reader has been stopped.
func (r *reader) Get() (InotifyEvent, error) {
	event, ok := <-r.resultChan
	if !ok {
		return event, ErrStopped
	}
	return event, nil
}

// add a watcher - use masks defined in syscall.
//
// returns a watch descriptor - store it only if you want to manually remove watch descriptors yourself.
func (r *reader) AddWatcher(pathname string, mask uint32) (int, error) {
	watch, err := syscall.InotifyAddWatch(r.fd, pathname, mask)
	if err != nil {
		return 0, err
	}
	r.wdMutex.Lock()
	r.wds[uint32(watch)] = true
	r.wdMutex.Unlock()
	return watch, nil
}

func (r *reader) RemoveWatcher(wd int) error {
	r.wdMutex.Lock()
	defer r.wdMutex.Unlock()

	watchDescriptor := uint32(wd)
	delete(r.wds, watchDescriptor)
	_, err := syscall.InotifyRmWatch(r.fd, watchDescriptor)
	return err
}

// start listening for inotify events. Use Stop to stop listening.
//
// Listen should be called in a separate goroutine because it is a blocking function.
func (r *reader) Listen() error {
	const iSize = syscall.SizeofInotifyEvent

	if r.stopped {
		return ErrStopped
	}
	buffer := make([]byte, iSize*r.bufferSize)

	for !r.stopped {
		var event *syscall.InotifyEvent
		read, err := syscall.Read(r.fd, buffer) // might hang here indefinitely
		if err != nil {
			return err
		}

		offset := 0
		for offset < read {
			event = (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[offset]))

			start := offset + iSize
			name := string(buffer[start : start+int(event.Len)])

			r.resultChan <- makeInotifyEvent(event, name)
			offset += iSize + int(event.Len)
		}
	}
	return nil
}

// stop listening to inotify events and remove all watch descriptors.
//
// a stopped reader cannot be started (by calling Listen) again.
func (r *reader) Stop() {
	defer close(r.resultChan)
	r.stopped = true
	r.wdMutex.Lock()
	for wd := range r.wds {
		_, _ = syscall.InotifyRmWatch(r.fd, wd)
	}
	r.wdMutex.Unlock()
	_ = syscall.Close(r.fd)
}
