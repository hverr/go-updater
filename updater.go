// Package updater provides auto-updating functionality for your application.
//
// Example for a GitHub application:
//	f := NewDelayedFile(os.Args[0])
//	defer f.Close()
//
//	u:= &Updater{
//		App: NewGitHub("hverr", "status-dashboard", nil),
//		CurrentReleaseIdentifier: "789611aec3d4b90512577b5dad9cf1adb6b20dcc",
//		WriterForAsset: func(a Asset) (AbortWriter, error) {
//			return f, nil
//		},
//	}
//
//	r, err := u.Check()
//	if err != nil {
//		panic(err)
//	}
//
//	if r == nil {
//		fmt.Println("No updates available.")
//	} else {
//		fmt.Println("Updating to", r.Name(), "-", r.Identifier())
//		err = u.UpdateTo(r)
//		if err != nil {
//			panic(err)
//		}
//	}
//
package updater

import "errors"

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
	WriterForAsset func(Asset) (AbortWriter, error)
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
func (u *Updater) UpdateTo(release Release) error {
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
		w, err := u.WriterForAsset(a)
		if err != nil {
			return err
		}

		if w != nil {
			e := a.Write(w)
			if e != nil {
				return e
			}
		}
	}

	return nil
}
