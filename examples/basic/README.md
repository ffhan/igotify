# Basic example

1. position yourself in the repository root
1. Run `go run ./examples/basic/`
1. run `rm -f test123 && echo 'Hello igotify!' > test123`
1. run `rm -f test123`

Expected output:
```
got event: InotifyEvent{ Wd: 1, Masks: IN_CREATE, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_OPEN, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_MODIFY, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_CLOSE_WRITE, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_OPEN, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_ACCESS, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_CLOSE_NOWRITE, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_OPEN, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_ACCESS, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_CLOSE_NOWRITE, Cookie: 0, Name: "test123         " }
got event: InotifyEvent{ Wd: 1, Masks: IN_ISDIR IN_OPEN, Cookie: 0, Name: "" }
got event: InotifyEvent{ Wd: 1, Masks: IN_ACCESS IN_ISDIR, Cookie: 0, Name: "" }
got event: InotifyEvent{ Wd: 1, Masks: IN_CLOSE_NOWRITE IN_ISDIR, Cookie: 0, Name: "" }
got event: InotifyEvent{ Wd: 1, Masks: IN_DELETE, Cookie: 0, Name: "test123         " }
```
