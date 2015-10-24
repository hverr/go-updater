package updater

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileBuffer(t *testing.T) {
	// Pre-defined file
	{
		// Get filename
		f, err := ioutil.TempFile("", "testing-")
		require.Nil(t, err)
		path := f.Name()
		f.Close()

		// Write file
		b := &FileBuffer{
			Path: path,
		}
		defer func() {
			err := b.Close()
			assert.Nil(t, err, "Close error: %v", err)
		}()

		_, err = b.Write([]byte("hello world"))
		assert.Nil(t, err, "Write error: %v", err)

		// Check contents
		f, err = os.Open(path)
		require.Nil(t, err, "Could not open file: %v", err)
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		assert.Equal(t, "hello world", string(data))

		// Clean up
		err = os.Remove(path)
		assert.Nil(t, err, "Could not clean up: %v", err)
	}

	// Temporary file
	{
		// Write file
		b := &FileBuffer{}
		defer func() {
			err := b.Close()
			assert.Nil(t, err, "Close error: %v", err)
		}()

		_, err := b.Write([]byte("hello world"))
		assert.Nil(t, err, "Write error: %v", err)
		require.NotEqual(t, "", b.Path)

		// Check contents
		f, err := os.Open(b.Path)
		require.Nil(t, err, "Could not open file: %v", err)
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		assert.Equal(t, "hello world", string(data))

		// Clean up
		err = os.Remove(b.Path)
		assert.Nil(t, err, "Could not clean up: %v", err)
	}

	// Error file
	{
		b := &FileBuffer{
			Path: "/n/o/n/e/x/i/s/t/i/n/g/file",
		}
		defer func() {
			err := b.Close()
			assert.Nil(t, err, "Close error: %v", err)
		}()

		_, err := b.Write([]byte("hello world"))
		assert.Error(t, err)
	}

	// Close unopend file
	{
		b := &FileBuffer{}
		err := b.Close()
		assert.Nil(t, err)
	}
}
