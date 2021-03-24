package version

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testReleaseData = `[
  {
    "url": "https://api.github.com/repos/weaveworks/weave-gitops/releases/40370706",
    "assets_url": "https://api.github.com/repos/weaveworks/weave-gitops/releases/40370706/assets",
    "upload_url": "https://uploads.github.com/repos/weaveworks/weave-gitops/releases/40370706/assets{?name,label}",
    "html_url": "https://github.com/weaveworks/weave-gitops/releases/tag/v2.6.28",
    "id": 40370706,
    "author": {
      "login": "jrryjcksn",
      "id": 24903211,
      "node_id": "MDQ6VXNlcjI0OTAzMjEx",
      "avatar_url": "https://avatars.githubusercontent.com/u/24903211?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/jrryjcksn",
      "html_url": "https://github.com/jrryjcksn",
      "followers_url": "https://api.github.com/users/jrryjcksn/followers",
      "following_url": "https://api.github.com/users/jrryjcksn/following{/other_user}",
      "gists_url": "https://api.github.com/users/jrryjcksn/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/jrryjcksn/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/jrryjcksn/subscriptions",
      "organizations_url": "https://api.github.com/users/jrryjcksn/orgs",
      "repos_url": "https://api.github.com/users/jrryjcksn/repos",
      "events_url": "https://api.github.com/users/jrryjcksn/events{/privacy}",
      "received_events_url": "https://api.github.com/users/jrryjcksn/received_events",
      "type": "User",
      "site_admin": false
    },
    "node_id": "MDc6UmVsZWFzZTQwMzcwNzA2",
    "tag_name": "v2.6.28",
    "target_commitish": "main",
    "name": "test release",
    "draft": false,
    "prerelease": false,
    "created_at": "2021-03-24T17:28:29Z",
    "published_at": "2021-03-24T18:02:15Z",
    "assets": [

    ],
    "tarball_url": "https://api.github.com/repos/weaveworks/weave-gitops/tarball/v2.6.28",
    "zipball_url": "https://api.github.com/repos/weaveworks/weave-gitops/zipball/v2.6.28",
    "body": "Release for testing version command"
  }
]`

func TestLessThan(t *testing.T) {
	lt, err := LessThan("v0.0.0", "v0.0.0")
	assert.NoError(t, err)
	assert.False(t, lt)

	lt, err = LessThan("v0.0.0", "v0.0.1")
	assert.NoError(t, err)
	assert.True(t, lt)

	lt, err = LessThan("v0.0.1", "v0.0.0")
	assert.NoError(t, err)
	assert.False(t, lt)

	_, err = LessThan("v0.0.1", "iejfo")
	assert.Error(t, err)

	_, err = LessThan("ewiew", "v0.0.1")
	assert.Error(t, err)

	_, err = LessThan("ewiew", "gorbamand")
	assert.Error(t, err)
}

func TestExtractLatestRelease(t *testing.T) {
	var data []interface{}
	err := json.Unmarshal([]byte(testReleaseData), &data)
	assert.NoError(t, err)
	rel, err := ExtractLatestRelease(data)
	assert.NoError(t, err)
	assert.Equal(t, rel, "v2.6.28")

	rel, err = ExtractLatestRelease([]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, rel, "v0.0.0")
}

func TestGetReleases(t *testing.T) {
	_, err := GetReleases()
	assert.NoError(t, err)
}
