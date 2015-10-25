package updater

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

const atomicFilePrefix = "atomic-"

// AbortWriter is a writer that can be aborted.
type AbortWriter interface {
	io.Writer

	// Abort writing. This is called when an error occurs.
	Abort()
}

// FileBuffer is a byte buffer stored on the filesystem.
//
// If no Path is specified, a temporary file is used and Path is set.
type FileBuffer struct {
	Path string

	opener    sync.Once
	openError error
	handle    *os.File
	aborted   bool
}

// Write data to the temporary file.
func (a *FileBuffer) Write(b []byte) (int, error) {
	if a.aborted {
		return 0, errors.New("Write operations aborted.")
	}

	// Open the file
	a.opener.Do(func() {
		if a.Path == "" {
			a.handle, a.openError = ioutil.TempFile("", "atomic-")
			a.Path = a.handle.Name()
		} else {
			a.handle, a.openError = os.Create(a.Path)
		}
	})
	if a.openError != nil {
		return 0, a.openError
	}

	// Write data to the temporary file
	return a.handle.Write(b)
}

// Abort writing. Subsequent calls to Write will return an error
func (a *FileBuffer) Abort() {
	a.aborted = true
}

// Close the file and rename to output file.
//
// The temporary file will be removed no matter what.
func (a *FileBuffer) Close() error {
	if a.handle != nil {
		return a.handle.Close()
	}
	return nil
}

// DelayedFile is a file that is first written to a temporary location.
//
// All writes to a delayed file go to a temporary file. When the file is closed,
// the destination file is replaced using os.Rename.
//
// This file type can be used to assure that all data is correctly received from
// an unreliable source, before the final destination file is written to.
type DelayedFile struct {
	path string

	buffer  FileBuffer
	aborted bool
}

// NewDelayedFile creates a new delayed file.
func NewDelayedFile(path string) *DelayedFile {
	return &DelayedFile{
		path: path,
	}
}

// Write data to the temporary file.
func (f *DelayedFile) Write(b []byte) (int, error) {
	return f.buffer.Write(b)
}

// Abort will stop the file from copying its contents to the final destination
// when the file is closed.
func (f *DelayedFile) Abort() {
	f.aborted = true
}

// Close will close the temporary file, copy its contents and delete it.
//
// If Abort was called before closing the file, the contents will not be copied
// to the final destination.
func (f *DelayedFile) Close() error {
	// Delete the temporary file
	defer os.Remove(f.buffer.Path)

	// Close the temporary file
	f.buffer.Close()

	// Don't copy if aborted
	if f.aborted {
		return nil
	}

	// Rename
	var mode *os.FileMode
	if info, _ := os.Stat(f.path); info != nil {
		m := info.Mode()
		mode = &m
	}

	err := os.Rename(f.buffer.Path, f.path)
	if err != nil {
		return err
	}

	if mode != nil {
		return os.Chmod(f.path, *mode)
	}

	return nil
}

// AbortBuffer is a buffer that can be aborted.
type AbortBuffer struct {
	Buffer *bytes.Buffer

	aborted bool
}

// NewAbortBuffer creates a new abort buffer
func NewAbortBuffer(b []byte) *AbortBuffer {
	return &AbortBuffer{
		Buffer: bytes.NewBuffer(b),
	}
}

// Write writes to the underlying buffer.
//
// If the buffer was aborted, an error is returned.
func (a *AbortBuffer) Write(b []byte) (int, error) {
	if a.aborted {
		return 0, errors.New("Write operations are aborted.")
	}

	return a.Buffer.Write(b)
}

// Abort blocks all subsequent write operations.
func (a *AbortBuffer) Abort() {
	a.aborted = true
}
