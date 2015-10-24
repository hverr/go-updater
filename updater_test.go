package updater

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdaterCheck(t *testing.T) {
	// Query error
	{
		testErr := errors.New("Test query error")
		app := &testApp{
			FQuery: func() error {
				return testErr
			},
		}
		u := &Updater{App: app}

		r, err := u.Check()
		assert.Nil(t, r)
		assert.Equal(t, err, testErr)
	}

	// No release information
	{
		app := &testApp{
			FLatestRelease: func() Release {
				return nil
			},
		}
		u := &Updater{App: app}

		r, err := u.Check()
		assert.Nil(t, r)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "No release info")
	}

	// Release identifiers match or differ
	{
		testRel := &testRelease{identifier: "new-release"}
		app := &testApp{
			FLatestRelease: func() Release { return testRel },
		}
		u := &Updater{
			App: app,
			CurrentReleaseIdentifier: "old-release",
		}

		r, err := u.Check()
		assert.Nil(t, err)
		assert.Equal(t, testRel, r)

		u.CurrentReleaseIdentifier = "new-release"
		r, err = u.Check()
		assert.Nil(t, err)
		assert.Nil(t, r)
	}
}

func TestUpdaterUpdateWithoutRelease(t *testing.T) {
	app := &testApp{}
	u := Updater{App: app}

	// With check error
	{
		testErr := errors.New("Test check error.")
		app.FQuery = func() error { return testErr }
		err := u.Update(nil)
		assert.Equal(t, err, testErr)
	}

	// Without release
	{
		app.FQuery = nil
		app.FLatestRelease = func() Release {
			return &testRelease{identifier: "new-release"}
		}
		u.CurrentReleaseIdentifier = "new-release"
		err := u.Update(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already up to date")
	}

	// Without error
	{
		app.FQuery = nil
		app.FLatestRelease = func() Release {
			return &testRelease{identifier: "ne-release"}
		}
		u.CurrentReleaseIdentifier = "old-release"
		err := u.Update(nil)
		assert.Nil(t, err)
	}
}

func TestUpdaterUpdateWithRelease(t *testing.T) {
	writeErr := errors.New("Writer test error")

	// Valid asset
	a1 := &testAsset{
		name: "asset1",
		write: func(w io.Writer) error {
			w.Write([]byte("Hello World!"))
			return nil
		},
	}

	// Asset without writer
	a2 := &testAsset{
		name: "asset2",
	}

	// Asset with error writer
	a3 := &testAsset{
		name:  "asset3",
		write: func(io.Writer) error { return writeErr },
	}

	validWriter := bytes.NewBuffer(nil)
	errorWriter := bytes.NewBuffer(nil)
	u := Updater{
		WriterForAsset: func(a Asset) io.Writer {
			if a == a1 {
				return validWriter
			} else if a == a2 {
				return nil
			} else if a == a3 {
				return errorWriter
			} else {
				require.True(t, false, "Unknown asset name: %v", a.Name())
				return nil
			}
		},
	}

	// Valid writer
	{
		err := u.Update(&testRelease{assets: []Asset{a1}})
		assert.Nil(t, err)
		assert.Equal(t, "Hello World!", validWriter.String())
	}

	// Asset without writer
	{
		err := u.Update(&testRelease{assets: []Asset{a2}})
		assert.Nil(t, err)
	}

	// Error writer
	{
		err := u.Update(&testRelease{assets: []Asset{a3}})
		assert.Equal(t, writeErr, err)
		assert.Equal(t, 0, errorWriter.Len())
	}
}

type testApp struct {
	FQuery         func() error
	FLatestRelease func() Release
}

func (a *testApp) Query() error {
	if a.FQuery != nil {
		return a.FQuery()
	}
	return nil
}

func (a *testApp) LatestRelease() Release {
	if a.FLatestRelease != nil {
		return a.FLatestRelease()
	}
	return nil
}

type testRelease struct {
	name, information, identifier string
	assets                        []Asset
}

func (r *testRelease) Name() string        { return r.name }
func (r *testRelease) Information() string { return r.information }
func (r *testRelease) Identifier() string  { return r.identifier }
func (r *testRelease) Assets() []Asset     { return r.assets }

type testAsset struct {
	name  string
	write func(io.Writer) error
}

func (a *testAsset) Name() string {
	return a.name
}

func (a *testAsset) Write(w io.Writer) error {
	if a.write != nil {
		return a.write(w)
	}
	return nil
}
