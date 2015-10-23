package updater

import "github.com/google/go-github/github"

type githubApp struct {
	owner      string
	repository string
	client     *github.Client
	releases   []Release
}

type githubRelease struct {
	RepositoryRelease github.RepositoryRelease
	Client            *github.Client
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
	releases, _, err := app.client.Repositories.ListReleases(app.owner, app.repository, nil)
	if err != nil {
		return err
	}

	s := make([]Release, len(releases))
	for i, r := range releases {
		s[i] = &githubRelease{
			Client:            app.client,
			RepositoryRelease: r,
		}
	}
	app.releases = s

	return nil
}

func (app *githubApp) LatestRelease() Release {
	if app.releases == nil {
		return nil
	}

	return app.releases[0]
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
