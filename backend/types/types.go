package types

type TempCode struct{
	Code string `json:"code"`
	Status string `json:"status"`
}

type RequestBodyAuth struct{
	ClientID string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
	CodeVerifier string `json:code_verifier`
}

type AccessTokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    Scope       string `json:"scope"`
}

type User struct{
	Name string `json:"login"`
	GithubUrl string `json:"html_url"`
	AvatarUrl string `json:"avatar_url"`
	PublicRepos int `json:"public_repos"`
	TotalPrivateRepos int `json:"total_private_repos"`
}

type GitHubRepo struct {
	ID          int64  `json:"id"`               // unique repo ID
	Name        string `json:"name"`        // title of the repo
	Description string `json:"description"` // repo description
	Language    string `json:"language"`    // main language used
	UpdatedAt   string `json:"updated_at"`  // last updated timestamp
	HTMLURL     string `json:"html_url"`    // URL to the repo
	Stargazers  int    `json:"stargazers_count"` // stars
	Forks       int    `json:"forks_count"`     // forks
	Private 	bool   `json:"private"`
}

type GitHubSearchResponse struct {
	Items GitHubRepoList `json:"items"`
}

// GitHubRepoList is just a slice of repos
type GitHubRepoList []GitHubRepo
