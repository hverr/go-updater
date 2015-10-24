package updater

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/github"
)

type githubApp struct {
	owner      string
	repository string
	client     *github.Client
	releases   []Release
}

type githubRelease struct {
	RepositoryRelease github.RepositoryRelease
	Reference         *github.Reference

	assets []Asset
}

type githubAsset struct {
	Asset github.ReleaseAsset
}

// NewGitHub creates an Application that is hosted on GitHub.
//
// Set client to nil to use the default one.
func NewGitHub(owner, repository string, client *github.Client) App {
	if client == nil {
		client = github.NewClient(nil)
	}

	return &githubApp{
		owner:      owner,
		repository: repository,

		client: client,
	}
}

func (app *githubApp) Query() error {
	// Get all available releases
	releases, _, err := app.client.Repositories.ListReleases(app.owner, app.repository, nil)
	if err != nil {
		return err
	}

	s := make([]Release, len(releases))
	for i, r := range releases {
		s[i] = newGithubRelease(r)
	}
	app.releases = s

	// Get the commit sha for the latest release
	if len(s) != 0 {
		e := s[0].(*githubRelease).queryReference(app)
		if e != nil {
			return e
		}
	}

	return nil
}

func (app *githubApp) LatestRelease() Release {
	if app.releases == nil {
		return nil
	}

	return app.releases[0]
}

func newGithubRelease(r github.RepositoryRelease) *githubRelease {
	s := make([]Asset, len(r.Assets))
	for i, a := range r.Assets {
		s[i] = &githubAsset{a}
	}

	return &githubRelease{
		RepositoryRelease: r,
		assets:            s,
	}
}

func (r *githubRelease) Name() string {
	if s := r.RepositoryRelease.TagName; s != nil {
		return *s
	}
	return ""
}

func (r *githubRelease) Information() string {
	if s := r.RepositoryRelease.Body; s != nil {
		return *s
	}
	return ""
}

func (r *githubRelease) Identifier() string {
	if r.Reference == nil || r.Reference.Object == nil || r.Reference.Object.SHA == nil {
		return ""
	}
	return *r.Reference.Object.SHA
}

func (r *githubRelease) Assets() []Asset {
	return r.assets
}

func (r *githubRelease) queryReference(app *githubApp) error {
	if r.RepositoryRelease.TagName == nil {
		return errors.New("No tag name available.")
	}

	tag := "tags/" + *r.RepositoryRelease.TagName
	ref, _, err := app.client.Git.GetRef(app.owner, app.repository, tag)
	if err != nil {
		return err
	}

	r.Reference = ref
	return nil
}

func (r *githubAsset) Name() string {
	if s := r.Asset.Name; s != nil {
		return *s
	}
	return ""
}

func (r *githubAsset) Write(w io.Writer) error {
	if r.Asset.BrowserDownloadURL == nil {
		return errors.New("No download URL available.")
	}

	resp, err := http.Get(*r.Asset.BrowserDownloadURL)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"Could not download %v: %v",
			*r.Asset.BrowserDownloadURL, resp.Status,
		)
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	return err
}
