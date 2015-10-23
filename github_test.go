package updater

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func newTestClient(f func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *github.Client) {
	ts := httptest.NewServer(http.HandlerFunc(f))
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(ts.URL)
		},
	}
	client := github.NewClient(&http.Client{
		Transport: transport,
	})

	u, _ := url.Parse("http://localhost/api/")
	client.BaseURL = u

	return ts, client
}

func TestGitHubQuery(t *testing.T) {
	// With valid JSON
	{
		ts, cl := newTestClient(func(w http.ResponseWriter, r *http.Request) {
			strings.NewReader(validReleasesJSON).WriteTo(w)
		})
		defer ts.Close()

		app := NewGitHub("hverr", "reponame", cl)
		err := app.Query()

		assert.Nil(t, err, "Unexpected query error: %v", err)

		release := app.LatestRelease()
		assert.NotNil(t, release)
		if release != nil {
			assert.Equal(t, "v1.0.0", release.Name())
			assert.Equal(t, "Description of the release", release.Information())
		}
	}

	// With invalid JSON
	{
		ts, cl := newTestClient(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		})
		defer ts.Close()

		app := NewGitHub("hverr", "reponame", cl)
		err := app.Query()
		assert.NotNil(t, err)
	}
}

func TestGitHubLatestRelease(t *testing.T) {
	// No information available
	{
		app := NewGitHub("hverr", "reponame", nil)
		assert.Nil(t, app.LatestRelease())
	}

	// Valid releases
	{
		r := &githubRelease{}
		app := githubApp{
			releases: []Release{r},
		}

		assert.Equal(t, r, app.LatestRelease())
	}
}

func TestGitHubRelease(t *testing.T) {
	r := githubRelease{}

	assert.Equal(t, "", r.Name())
	assert.Equal(t, "", r.Information())

	tagName := "v1.0.1"
	body := "Hello World!"
	r.RepositoryRelease.TagName = &tagName
	r.RepositoryRelease.Body = &body

	assert.Equal(t, "v1.0.1", r.Name())
	assert.Equal(t, "Hello World!", r.Information())
}

var validReleasesJSON = `

[
  {
    "url": "https://api.github.com/repos/octocat/Hello-World/releases/1",
    "html_url": "https://github.com/octocat/Hello-World/releases/v1.0.0",
    "assets_url": "https://api.github.com/repos/octocat/Hello-World/releases/1/assets",
    "upload_url": "https://uploads.github.com/repos/octocat/Hello-World/releases/1/assets{?name,label}",
    "tarball_url": "https://api.github.com/repos/octocat/Hello-World/tarball/v1.0.0",
    "zipball_url": "https://api.github.com/repos/octocat/Hello-World/zipball/v1.0.0",
    "id": 1,
    "tag_name": "v1.0.0",
    "target_commitish": "master",
    "name": "v1.0.0",
    "body": "Description of the release",
    "draft": false,
    "prerelease": false,
    "created_at": "2013-02-27T19:35:32Z",
    "published_at": "2013-02-27T19:35:32Z",
    "author": {
      "login": "octocat",
      "id": 1,
      "avatar_url": "https://github.com/images/error/octocat_happy.gif",
      "gravatar_id": "",
      "url": "https://api.github.com/users/octocat",
      "html_url": "https://github.com/octocat",
      "followers_url": "https://api.github.com/users/octocat/followers",
      "following_url": "https://api.github.com/users/octocat/following{/other_user}",
      "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
      "organizations_url": "https://api.github.com/users/octocat/orgs",
      "repos_url": "https://api.github.com/users/octocat/repos",
      "events_url": "https://api.github.com/users/octocat/events{/privacy}",
      "received_events_url": "https://api.github.com/users/octocat/received_events",
      "type": "User",
      "site_admin": false
    },
    "assets": [
      {
        "url": "https://api.github.com/repos/octocat/Hello-World/releases/assets/1",
        "browser_download_url": "https://github.com/octocat/Hello-World/releases/download/v1.0.0/example.zip",
        "id": 1,
        "name": "example.zip",
        "label": "short description",
        "state": "uploaded",
        "content_type": "application/zip",
        "size": 1024,
        "download_count": 42,
        "created_at": "2013-02-27T19:35:32Z",
        "updated_at": "2013-02-27T19:35:32Z",
        "uploader": {
          "login": "octocat",
          "id": 1,
          "avatar_url": "https://github.com/images/error/octocat_happy.gif",
          "gravatar_id": "",
          "url": "https://api.github.com/users/octocat",
          "html_url": "https://github.com/octocat",
          "followers_url": "https://api.github.com/users/octocat/followers",
          "following_url": "https://api.github.com/users/octocat/following{/other_user}",
          "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
          "organizations_url": "https://api.github.com/users/octocat/orgs",
          "repos_url": "https://api.github.com/users/octocat/repos",
          "events_url": "https://api.github.com/users/octocat/events{/privacy}",
          "received_events_url": "https://api.github.com/users/octocat/received_events",
          "type": "User",
          "site_admin": false
        }
      }
    ]
  }
]
`
