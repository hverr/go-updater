# go-updater
[![Build Status](https://travis-ci.org/hverr/go-updater.svg?branch=master)](https://travis-ci.org/hverr/go-updater)
[![Coverage Status](https://coveralls.io/repos/hverr/go-updater/badge.svg?branch=master)](https://coveralls.io/r/hverr/go-updater?branch=master)
[![GoDoc](https://godoc.org/github.com/hverr/go-updater?status.svg)](http://godoc.org/github.com/hverr/go-updater)

Package updater provides auto-updating functionality for your application.

Example for a GitHub application:

```go
f := NewDelayedFile(os.Args[0])
defer f.Close()

u:= &Updater{
	App: NewGitHub("hverr", "status-dashboard", nil),
	CurrentReleaseIdentifier: "789611aec3d4b90512577b5dad9cf1adb6b20dcc",
	WriterForAsset: func(a Asset) (AbortWriter, error) {
		return f, nil
	},
}

r, err := u.Check()
if err != nil {
	panic(err)
}

if r == nil {
	fmt.Println("No updates available.")
} else {
	fmt.Println("Updating to", r.Name(), "-", r.Identifier())
	err = u.UpdateTo(r)
	if err != nil {
		panic(err)
	}
}
```
