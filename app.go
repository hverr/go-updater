package updater

import "io"

// App is a generic Go application capapble of querying update
// information and updating itself.
type App interface {
	// Query sould query application information from a remote source.
	Query() error

	// LatestRelease should return the most recent release of the application
	// that is available.
	LatestRelease() Release
}

// Release represents an application release.
type Release interface {
	// Name should return the version name of this release.
	Name() string

	// Information is some human-readable information for this release.
	Information() string

	// Identifier should be the identifier of this release. This identifier is
	// used to compare releases.
	Identifier() string

	// Assets sould return all assets attached to this release.
	Assets() []Asset
}

// Asset represents a downloadable asset.
type Asset interface {
	// Name should return the file name of the asset.
	Name() string

	// Write should write the contents of the asset.
	Write(w io.Writer) error
}
