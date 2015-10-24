package updater

import (
	"errors"
	"io"
)

// Updater is used to directly update the application.
type Updater struct {
	// Application to update.
	App App

	// Identifier of the current release.
	//
	// If the identifier of the latest release differs the current release
	// identifier, the updater will update the application.
	CurrentReleaseIdentifier string

	// Function to map assets to a writer.
	//
	// When the app is updated, this function will be called for each asset
	// in the latest release. The updater will write the asset to the returned
	// io.Writer.
	//
	// You can return nil to ignore the asset.
	WriterForAsset func(Asset) io.Writer
}

// Check will check for updates.
//
// When an update is available, it will return the release for this update. You
// can use it to inform the user about the update.
//
// When the application is already up to date, nil is returned.
func (u *Updater) Check() (Release, error) {
	// Query app information
	err := u.App.Query()
	if err != nil {
		return nil, err
	}

	// Get the latest available release
	r := u.App.LatestRelease()
	if r == nil {
		return nil, errors.New("No release information was found.")
	}

	// Check if the release is newer
	if r.Identifier() == u.CurrentReleaseIdentifier {
		return nil, nil
	}

	// Return the latest release
	return r, nil
}

// UpdateTo will update the application.
//
// If you don't specify a release, the updater will first fetch all releases and
// try to update to the most recent one.
func (u *Updater) Update(release Release) error {
	if release == nil {
		var err error
		release, err = u.Check()
		if err != nil {
			return err
		}
		if release == nil {
			return errors.New("The application is already up to date.")
		}
	}

	for _, a := range release.Assets() {
		w := u.WriterForAsset(a)
		if w != nil {
			e := a.Write(w)
			if e != nil {
				return e
			}
		}
	}

	return nil
}
