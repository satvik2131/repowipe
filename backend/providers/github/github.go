package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"repowipe/providers"
	"repowipe/types"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	authorizeURL = "https://github.com/login/oauth/authorize"
	tokenURL     = "https://github.com/login/oauth/access_token"
	apiBase      = "https://api.github.com"
	revokeBase   = "https://api.github.com/applications/"
)

func init() {
	providers.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() types.Provider { return types.ProviderGitHub }

func clientID() string     { return os.Getenv("GITHUB_CLIENT_ID") }
func clientSecret() string { return os.Getenv("GITHUB_CLIENT_SECRET") }

func (p *Provider) api() *resty.Client {
	return resty.New().
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28")
}

func (p *Provider) AuthorizeURL(redirectURI, state, mode string) (string, error) {
	scope := "repo,user,delete_repo"
	u, _ := url.Parse(authorizeURL)
	q := u.Query()
	q.Set("client_id", clientID())
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", scope)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (p *Provider) ExchangeCode(code, redirectURI string) (*types.OAuthToken, error) {
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	resp, err := p.api().R().
		SetHeader("Accept", "application/json").
		SetFormData(map[string]string{
			"client_id":     clientID(),
			"client_secret": clientSecret(),
			"code":          code,
			"redirect_uri":  redirectURI,
		}).
		SetResult(&tokenResp).
		Post(tokenURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK || tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("github token exchange failed: status=%d body=%s", resp.StatusCode(), resp.String())
	}
	return &types.OAuthToken{AccessToken: tokenResp.AccessToken, Scope: tokenResp.Scope}, nil
}

func (p *Provider) RefreshToken(string) (*types.OAuthToken, error) {
	return nil, errors.New("github oauth apps do not use refresh tokens")
}

func (p *Provider) Revoke(accessToken string) error {
	if accessToken == "" {
		return nil
	}
	resp, err := p.api().R().
		SetBasicAuth(clientID(), clientSecret()).
		SetBody(map[string]string{"access_token": accessToken}).
		Delete(revokeBase + clientID() + "/token")
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusNotFound {
		return fmt.Errorf("github revoke failed: %d", resp.StatusCode())
	}
	return nil
}

func (p *Provider) GetUser(accessToken string) (*types.UserProfile, error) {
	var raw struct {
		Login             string `json:"login"`
		HTMLURL           string `json:"html_url"`
		AvatarURL         string `json:"avatar_url"`
		PublicRepos       int    `json:"public_repos"`
		TotalPrivateRepos int    `json:"total_private_repos"`
	}
	resp, err := p.api().R().
		SetAuthToken(accessToken).
		SetResult(&raw).
		Get(apiBase + "/user")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("github get user: %d", resp.StatusCode())
	}
	return &types.UserProfile{
		Login:             raw.Login,
		HTMLURL:           raw.HTMLURL,
		AvatarURL:         raw.AvatarURL,
		PublicRepos:       raw.PublicRepos,
		TotalPrivateRepos: raw.TotalPrivateRepos,
		Provider:          types.ProviderGitHub,
	}, nil
}

type ghRepo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	UpdatedAt   string `json:"updated_at"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	Stargazers  int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Private     bool   `json:"private"`
	Fork        bool   `json:"fork"`
	Archived    bool   `json:"archived"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func mapRepo(r ghRepo) types.Repo {
	return types.Repo{
		ID: r.ID, Name: r.Name, FullName: r.FullName, Description: r.Description,
		Language: r.Language, UpdatedAt: r.UpdatedAt, HTMLURL: r.HTMLURL, CloneURL: r.CloneURL,
		Stargazers: r.Stargazers, Forks: r.Forks, Private: r.Private, Fork: r.Fork, Archived: r.Archived,
		Provider: types.ProviderGitHub, OwnerLogin: r.Owner.Login,
	}
}

func (p *Provider) ListRepos(accessToken string, page int, visibility, sort, direction string) ([]types.Repo, error) {
	if page < 1 {
		page = 1
	}
	if visibility == "" {
		visibility = "all"
	}
	if sort == "" {
		sort = "updated"
	}
	if direction == "" {
		direction = "desc"
	}
	var raw []ghRepo
	resp, err := p.api().R().
		SetAuthToken(accessToken).
		SetQueryParams(map[string]string{
			"page": fmt.Sprintf("%d", page), "per_page": "10",
			"visibility": visibility, "sort": sort, "direction": direction, "affiliation": "owner",
		}).
		SetResult(&raw).
		Get(apiBase + "/user/repos")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("github list repos: %d", resp.StatusCode())
	}
	out := make([]types.Repo, 0, len(raw))
	for _, r := range raw {
		out = append(out, mapRepo(r))
	}
	return out, nil
}

func (p *Provider) SearchRepos(accessToken, username, query, language, visibility, kind, sort string) ([]types.Repo, error) {
	q := fmt.Sprintf("user:%s", username)
	if query != "" {
		q += " " + query + " in:name,description"
	}
	if language != "" {
		q += " language:" + language
	}
	switch visibility {
	case "public":
		q += " is:public"
	case "private":
		q += " is:private"
	}
	switch kind {
	case "forks":
		q += " fork:only archived:false"
	case "sources":
		q += " fork:false archived:false"
	case "archived":
		q += " archived:true"
	}
	if sort != "stars" && sort != "updated" && sort != "forks" {
		sort = "updated"
	}
	var result struct {
		Items []ghRepo `json:"items"`
	}
	resp, err := p.api().R().
		SetAuthToken(accessToken).
		SetQueryParams(map[string]string{"q": q, "sort": sort, "order": "desc", "per_page": "30"}).
		SetResult(&result).
		Get(apiBase + "/search/repositories")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("github search: %d", resp.StatusCode())
	}
	out := make([]types.Repo, 0, len(result.Items))
	for _, r := range result.Items {
		out = append(out, mapRepo(r))
	}
	return out, nil
}

func (p *Provider) DeleteRepo(accessToken, owner, name string) error {
	resp, err := p.api().R().SetAuthToken(accessToken).Delete(apiBase + "/repos/" + owner + "/" + name)
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return fmt.Errorf("not found: %s/%s", owner, name)
	}
	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("github delete: %d", resp.StatusCode())
	}
	return nil
}

func (p *Provider) CreateRepo(accessToken, name, description string, private bool) (*types.Repo, error) {
	var raw ghRepo
	resp, err := p.api().R().
		SetAuthToken(accessToken).
		SetBody(map[string]any{"name": name, "description": description, "private": private, "auto_init": false}).
		SetResult(&raw).
		Post(apiBase + "/user/repos")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("github create repo: %d %s", resp.StatusCode(), resp.String())
	}
	r := mapRepo(raw)
	return &r, nil
}

func (p *Provider) AuthenticatedCloneURL(accessToken, cloneURL string) string {
	// https://github.com/owner/repo.git → https://x-access-token:TOKEN@github.com/owner/repo.git
	u := strings.TrimPrefix(cloneURL, "https://")
	return "https://x-access-token:" + accessToken + "@" + u
}

func (p *Provider) ListLabels(accessToken, owner, name string) ([]types.Label, error) {
	var raw []struct {
		Name        string `json:"name"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).SetResult(&raw).
		Get(fmt.Sprintf("%s/repos/%s/%s/labels?per_page=100", apiBase, owner, name))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("github labels: %d", resp.StatusCode())
	}
	out := make([]types.Label, 0, len(raw))
	for _, l := range raw {
		out = append(out, types.Label{Name: l.Name, Color: l.Color, Description: l.Description})
	}
	return out, nil
}

func (p *Provider) ListIssues(accessToken, owner, name string) ([]types.Issue, error) {
	var raw []struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		State     string `json:"state"`
		HTMLURL   string `json:"html_url"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		PullRequest *struct{} `json:"pull_request"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetQueryParams(map[string]string{"state": "all", "per_page": "100"}).
		SetResult(&raw).
		Get(fmt.Sprintf("%s/repos/%s/%s/issues", apiBase, owner, name))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("github issues: %d", resp.StatusCode())
	}
	out := make([]types.Issue, 0)
	for _, i := range raw {
		if i.PullRequest != nil {
			continue // PRs appear in issues API; skip
		}
		labels := make([]string, 0, len(i.Labels))
		for _, l := range i.Labels {
			labels = append(labels, l.Name)
		}
		out = append(out, types.Issue{
			Number: i.Number, Title: i.Title, Body: i.Body, State: i.State,
			HTMLURL: i.HTMLURL, Author: i.User.Login, Labels: labels,
			CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt,
		})
	}
	return out, nil
}

func (p *Provider) ListPullRequests(accessToken, owner, name string) ([]types.PullRequest, error) {
	var raw []struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
		CreatedAt string `json:"created_at"`
		MergedAt  *string `json:"merged_at"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
		Head struct {
			Ref string `json:"ref"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetQueryParams(map[string]string{"state": "all", "per_page": "100"}).
		SetResult(&raw).
		Get(fmt.Sprintf("%s/repos/%s/%s/pulls", apiBase, owner, name))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("github pulls: %d", resp.StatusCode())
	}
	out := make([]types.PullRequest, 0, len(raw))
	for _, pr := range raw {
		labels := make([]string, 0, len(pr.Labels))
		for _, l := range pr.Labels {
			labels = append(labels, l.Name)
		}
		merged := pr.MergedAt != nil
		state := pr.State
		if merged {
			state = "merged"
		}
		out = append(out, types.PullRequest{
			Number: pr.Number, Title: pr.Title, Body: pr.Body, State: state,
			HTMLURL: pr.HTMLURL, Author: pr.User.Login, HeadRef: pr.Head.Ref, BaseRef: pr.Base.Ref,
			Labels: labels, CreatedAt: pr.CreatedAt, Merged: merged,
		})
	}
	return out, nil
}

func (p *Provider) EnsureLabels(accessToken, owner, name string, labels []types.Label) error {
	for _, l := range labels {
		color := strings.TrimPrefix(l.Color, "#")
		if color == "" {
			color = "ededed"
		}
		resp, err := p.api().R().SetAuthToken(accessToken).
			SetBody(map[string]string{"name": l.Name, "color": color, "description": l.Description}).
			Post(fmt.Sprintf("%s/repos/%s/%s/labels", apiBase, owner, name))
		if err != nil {
			return err
		}
		// 422 = already exists — ok
		if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusUnprocessableEntity {
			return fmt.Errorf("github create label %s: %d", l.Name, resp.StatusCode())
		}
	}
	return nil
}

func (p *Provider) CreateIssue(accessToken, owner, name string, issue types.Issue) error {
	body := issue.Body
	if issue.Author != "" || issue.HTMLURL != "" {
		body += fmt.Sprintf("\n\n---\n_Imported from GitHub issue #%d by @%s — %s_", issue.Number, issue.Author, issue.HTMLURL)
	}
	payload := map[string]any{"title": issue.Title, "body": body}
	if len(issue.Labels) > 0 {
		payload["labels"] = issue.Labels
	}
	resp, err := p.api().R().SetAuthToken(accessToken).SetBody(payload).
		Post(fmt.Sprintf("%s/repos/%s/%s/issues", apiBase, owner, name))
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("github create issue: %d %s", resp.StatusCode(), resp.String())
	}
	if issue.State == "closed" {
		var created struct {
			Number int `json:"number"`
		}
		_ = json.Unmarshal(resp.Body(), &created)
		if created.Number > 0 {
			_, _ = p.api().R().SetAuthToken(accessToken).
				SetBody(map[string]string{"state": "closed"}).
				Patch(fmt.Sprintf("%s/repos/%s/%s/issues/%d", apiBase, owner, name, created.Number))
		}
	}
	return nil
}

func (p *Provider) CreatePullRequest(accessToken, owner, name string, pr types.PullRequest) error {
	body := pr.Body
	if pr.HTMLURL != "" {
		body += fmt.Sprintf("\n\n---\n_Imported from %s_", pr.HTMLURL)
	}
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetBody(map[string]string{
			"title": pr.Title, "body": body, "head": pr.HeadRef, "base": pr.BaseRef,
		}).
		Post(fmt.Sprintf("%s/repos/%s/%s/pulls", apiBase, owner, name))
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("github create pr: %d %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (p *Provider) WikiCloneURL(accessToken, owner, name string) (string, bool) {
	return fmt.Sprintf("https://github.com/%s/%s.wiki.git", owner, name), true
}
