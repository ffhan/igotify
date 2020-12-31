# igotify

A thin wrapper around *inotify* Linux syscalls and capabilities. Allows for easier interfacing with *inotify* through
usual Go patterns instead of directly using the *syscall* library.

See `examples/basic` for basic usage and tips on how to use the wrapper.

## How to use

`go get github.com/ffhan/igotify@v0.3.0`

## Prerequisites

* *igotify* uses `unistd.h` GNU library to fetch a `NAME_MAX` configuration value for the system. If not present on the
  machine, *igotify* will assume NAME_MAX to be 255. You can test if that assumption is correct for the system with
  `getconf NAME_MAX <filepath>` for any path.
