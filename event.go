package igotify

import (
	"fmt"
	"strings"
	"syscall"
)

const (
	IN_ACCESS        = "IN_ACCESS"
	IN_ATTRIB        = "IN_ATTRIB"
	IN_CLOSE_NOWRITE = "IN_CLOSE_NOWRITE"
	IN_CLOSE_WRITE   = "IN_CLOSE_WRITE"
	IN_CREATE        = "IN_CREATE"
	IN_DELETE        = "IN_DELETE"
	IN_DELETE_SELF   = "IN_DELETE_SELF"
	IN_IGNORED       = "IN_IGNORED"
	IN_ISDIR         = "IN_ISDIR"
	IN_MODIFY        = "IN_MODIFY"
	IN_MOVE_SELF     = "IN_MOVE_SELF"
	IN_MOVED_FROM    = "IN_MOVED_FROM"
	IN_MOVED_TO      = "IN_MOVED_TO"
	IN_OPEN          = "IN_OPEN"
	IN_Q_OVERFLOW    = "IN_Q_OVERFLOW"
	IN_UNMOUNT       = "IN_UNMOUNT"
)

type maskPair struct {
	name string
	mask uint32
}

var pairs = [...]maskPair{
	{IN_ACCESS, syscall.IN_ACCESS},
	{IN_ATTRIB, syscall.IN_ATTRIB},
	{IN_CLOSE_NOWRITE, syscall.IN_CLOSE_NOWRITE},
	{IN_CLOSE_WRITE, syscall.IN_CLOSE_WRITE},
	{IN_CREATE, syscall.IN_CREATE},
	{IN_DELETE, syscall.IN_DELETE},
	{IN_DELETE_SELF, syscall.IN_DELETE_SELF},
	{IN_IGNORED, syscall.IN_IGNORED},
	{IN_ISDIR, syscall.IN_ISDIR},
	{IN_MODIFY, syscall.IN_MODIFY},
	{IN_MOVE_SELF, syscall.IN_MOVE_SELF},
	{IN_MOVED_FROM, syscall.IN_MOVED_FROM},
	{IN_MOVED_TO, syscall.IN_MOVED_TO},
	{IN_OPEN, syscall.IN_OPEN},
	{IN_Q_OVERFLOW, syscall.IN_Q_OVERFLOW},
	{IN_UNMOUNT, syscall.IN_UNMOUNT},
}

type InotifyEvent struct {
	Wd     int32
	Mask   uint32
	Cookie uint32
	Len    uint32
	Name   string
}

func toEventString(sb *strings.Builder, mask, val uint32, representation string) {
	if mask&val != 0 {
		if sb.Len() > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteString(representation)
	}
}

func maskString(mask uint32) string {
	var sb strings.Builder
	for _, p := range pairs {
		toEventString(&sb, mask, p.mask, p.name)
	}
	return sb.String()
}

func (i InotifyEvent) String() string {
	return fmt.Sprintf("InotifyEvent{ Wd: %d, Masks: %s, Cookie: %d, Name: \"%s\" }",
		i.Wd, maskString(i.Mask), i.Cookie, i.Name)
}

func makeInotifyEvent(ev *syscall.InotifyEvent, name string) InotifyEvent {
	return InotifyEvent{
		Wd:     ev.Wd,
		Mask:   ev.Mask,
		Cookie: ev.Cookie,
		Len:    ev.Len,
		Name:   name,
	}
}
