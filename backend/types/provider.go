package types

// Provider identifies a git host.
type Provider string

const (
	ProviderGitHub    Provider = "github"
	ProviderGitLab    Provider = "gitlab"
	ProviderBitbucket Provider = "bitbucket"
)

func (p Provider) Valid() bool {
	switch p {
	case ProviderGitHub, ProviderGitLab, ProviderBitbucket:
		return true
	default:
		return false
	}
}

// ProviderCredentials holds OAuth tokens for one linked account.
type ProviderCredentials struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	ExpiresAt    int64       `json:"expires_at,omitempty"` // unix seconds; 0 = unknown/no expiry
	User         UserProfile `json:"user"`
}

// SessionDocument is stored in Redis under session:{id}.
type SessionDocument struct {
	Primary   Provider                       `json:"primary"`
	Providers map[Provider]ProviderCredentials `json:"providers"`
}

// UserProfile is a normalized identity across hosts.
type UserProfile struct {
	Login             string   `json:"login"`
	HTMLURL           string   `json:"html_url"`
	AvatarURL         string   `json:"avatar_url"`
	PublicRepos       int      `json:"public_repos"`
	TotalPrivateRepos int      `json:"total_private_repos,omitempty"`
	Provider          Provider `json:"provider"`
}

// Repo is a normalized repository across hosts.
type Repo struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	UpdatedAt   string   `json:"updated_at"`
	HTMLURL     string   `json:"html_url"`
	CloneURL    string   `json:"clone_url"`
	Stargazers  int      `json:"stargazers_count"`
	Forks       int      `json:"forks_count"`
	Private     bool     `json:"private"`
	Fork        bool     `json:"fork"`
	Archived    bool     `json:"archived"`
	Provider    Provider `json:"provider"`
	OwnerLogin  string   `json:"owner_login"`
}

// Label is a normalized issue/PR label.
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// Issue is a normalized issue.
type Issue struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	State     string   `json:"state"` // open | closed
	HTMLURL   string   `json:"html_url"`
	Author    string   `json:"author"`
	Labels    []string `json:"labels"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// PullRequest is a normalized PR/MR.
type PullRequest struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	State     string   `json:"state"` // open | closed | merged
	HTMLURL   string   `json:"html_url"`
	Author    string   `json:"author"`
	HeadRef   string   `json:"head_ref"`
	BaseRef   string   `json:"base_ref"`
	Labels    []string `json:"labels"`
	CreatedAt string   `json:"created_at"`
	Merged    bool     `json:"merged"`
}

// OAuthToken holds a token exchange result.
type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // seconds; 0 if unknown
	Scope        string
}

// AuthCallbackRequest is the frontend → backend OAuth callback payload.
type AuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
	Mode  string `json:"mode"` // login | link
}

// RepoDeleteRequest deletes repos on a provider.
type RepoDeleteRequest struct {
	Repos    []string `json:"repos"`
	Username string   `json:"username"`
}

// TransferRequest starts an any-to-any transfer job.
type TransferRequest struct {
	Source      Provider `json:"source"`
	Destination Provider `json:"destination"`
	Repos       []string `json:"repos"` // full_name or name
}

// TransferRepoResult is per-repo outcome.
type TransferRepoResult struct {
	Repo     string   `json:"repo"`
	Status   string   `json:"status"` // pending | running | succeeded | partial | failed
	DestURL  string   `json:"dest_url,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// TransferJob is stored in Redis under transfer:{id}.
type TransferJob struct {
	ID          string               `json:"id"`
	SessionID   string               `json:"session_id"`
	Source      Provider             `json:"source"`
	Destination Provider             `json:"destination"`
	Status      string               `json:"status"` // queued | running | completed | failed
	Repos       []TransferRepoResult `json:"repos"`
	CreatedAt   int64                `json:"created_at"`
	UpdatedAt   int64                `json:"updated_at"`
	Error       string               `json:"error,omitempty"`
}

// ConnectionsResponse is returned by GET /auth/connections.
type ConnectionsResponse struct {
	Primary     Provider                `json:"primary"`
	Connections map[Provider]UserProfile `json:"connections"`
}
