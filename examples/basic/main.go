package main

import (
	"fmt"
	"github.com/ffhan/igotify"
	"syscall"
)

func main() {
	reader, err := igotify.NewReader(igotify.DefaultBufferSize, igotify.DefaultFlags)
	if err != nil {
		panic(err)
	}
	defer reader.Stop()

	// Current Working Directory Watcher Descriptor
	cwdWD, err := reader.AddWatcher(".", syscall.IN_ALL_EVENTS)
	if err != nil {
		panic(err)
	}
	defer reader.RemoveWatcher(cwdWD)

	// reader will not listen for events unless we explicitly call Listen
	go func() {
		err := reader.Listen()
		if err != nil {
			panic(err)
		}
	}()

	for {
		event, err := reader.Get()
		if err != nil {
			panic(err)
		}
		fmt.Printf("got event: %s\n", event)
	}
}
