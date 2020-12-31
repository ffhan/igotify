package igotify

import (
	"errors"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	DefaultBufferSize = 128
	DefaultFlags      = 0
)

var (
	ErrStopped   = errors.New("reader is stopped")
	ErrTimeout   = errors.New("reader timeout")
	ErrListening = errors.New("reader is already listening")
)

// reader is a wrapper for reading and controlling inotify events.
type reader struct {
	fd         int               // inotify file descriptor
	wds        map[uint32]bool   // watch descriptors
	wdMutex    sync.Mutex        // watch descriptor mutex
	bufferSize int               // read buffer size (in sizeof(inotify_event))
	resultChan chan InotifyEvent //
	listening  bool              // indicates whether the Listen method has been called
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
// returns an error if the reader has been stopped (ErrStopped).
func (r *reader) Get() (InotifyEvent, error) {
	event, ok := <-r.resultChan
	if !ok {
		return event, ErrStopped
	}
	return event, nil
}

// retrieves a single InotifyEvent, blocks until a read can be completed or until a timeout.
//
// returns an error if the reader has been stopped (ErrStopped) or a timeout (ErrTimeout) happened.
func (r *reader) GetWithTimeout(timeout time.Duration) (InotifyEvent, error) {
	select {
	case event, ok := <-r.resultChan:
		if !ok {
			return event, ErrStopped
		}
		return event, nil
	case <-time.After(timeout):
		return InotifyEvent{}, ErrTimeout
	}
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
	if r.listening {
		return ErrListening
	}
	r.listening = true

	defer close(r.resultChan)
	const iSize = syscall.SizeofInotifyEvent

	if r.stopped {
		return ErrStopped
	}
	/*
		sizeof(inotify_event) * r.bufferSize ensures we can store max bufferSize
		 events with empty names
		there are two extremes we have to take care of:
		- bufferSize number of events with no filename attached (len 0)
		- 1 event with maximum allowed filename length

		Anything in between is not a problem because read() shouldn't give us less than the whole inotify_event structs in the buffer.

		So, when we receive no filenames we can receive bufferSize events from the same read() syscall.
		NAME_MAX + 1 serves two purposes: NAME_MAX ensures we can store at least one event with max filename path
		and +1 ensures we can store the \0 byte (end of the string).

		In that case, we can store at least bufferSize-1 additional events with empty filenames.
		Realistically, we mostly expect short filenames or empty filenames, and all the cases in between should be handled by read() syscall.

		Practically this means we can store minimum of 1 event with max filename up top bufferSize events with no filename provided.
	*/

	buffer := make([]byte, iSize*r.bufferSize+GetNameMax()+1)

	for !r.stopped {
		var event *syscall.InotifyEvent
		read, err := syscall.Read(r.fd, buffer) // might hang here indefinitely
		if err != nil {
			if err == syscall.EBADF && r.stopped { // if fd not open return without error - we're expecting this.
				return nil
			}
			return err
		}

		offset := 0
		for offset < read {
			event = (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[offset]))

			start := offset + iSize
			name := string(buffer[start : start+int(event.Len)])

			r.resultChan <- makeInotifyEvent(event, strings.TrimRight(name, "\000"))
			offset += iSize + int(event.Len)
		}
	}
	return nil
}

// stop listening to inotify events and remove all watch descriptors.
//
// a stopped reader cannot be started (by calling Listen) again.
func (r *reader) Stop() {
	r.stopped = true
	r.listening = false
	r.wdMutex.Lock()
	for wd := range r.wds {
		_, _ = syscall.InotifyRmWatch(r.fd, wd)
	}
	r.wdMutex.Unlock()
	_ = syscall.Close(r.fd)
}
