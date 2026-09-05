package providers

import "repowipe/types"

// Provider is the git-host adapter used for OAuth, repos, and transfer metadata.
type Provider interface {
	Name() types.Provider

	// OAuth
	AuthorizeURL(redirectURI, state, mode string) (string, error)
	ExchangeCode(code, redirectURI string) (*types.OAuthToken, error)
	RefreshToken(refreshToken string) (*types.OAuthToken, error)
	Revoke(accessToken string) error
	GetUser(accessToken string) (*types.UserProfile, error)

	// Repos
	ListRepos(accessToken string, page int, visibility, sort, direction string) ([]types.Repo, error)
	SearchRepos(accessToken, username, query, language, visibility, kind, sort string) ([]types.Repo, error)
	DeleteRepo(accessToken, owner, name string) error
	CreateRepo(accessToken, name, description string, private bool) (*types.Repo, error)

	// Clone URL with embedded credentials for git mirror.
	AuthenticatedCloneURL(accessToken, cloneURL string) string

	// Metadata export (source)
	ListLabels(accessToken, owner, name string) ([]types.Label, error)
	ListIssues(accessToken, owner, name string) ([]types.Issue, error)
	ListPullRequests(accessToken, owner, name string) ([]types.PullRequest, error)

	// Metadata import (destination)
	EnsureLabels(accessToken, owner, name string, labels []types.Label) error
	CreateIssue(accessToken, owner, name string, issue types.Issue) error
	CreatePullRequest(accessToken, owner, name string, pr types.PullRequest) error

	// Wiki: returns true if a git-based wiki clone URL is available.
	WikiCloneURL(accessToken, owner, name string) (string, bool)
}
