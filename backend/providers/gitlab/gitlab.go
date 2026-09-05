package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"repowipe/providers"
	"repowipe/types"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	authorizeURL = "https://gitlab.com/oauth/authorize"
	tokenURL     = "https://gitlab.com/oauth/token"
	apiBase      = "https://gitlab.com/api/v4"
)

func init() {
	providers.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() types.Provider { return types.ProviderGitLab }

func clientID() string     { return os.Getenv("GITLAB_CLIENT_ID") }
func clientSecret() string { return os.Getenv("GITLAB_CLIENT_SECRET") }

func (p *Provider) api() *resty.Client {
	return resty.New().SetHeader("Accept", "application/json")
}

func (p *Provider) AuthorizeURL(redirectURI, state, mode string) (string, error) {
	scope := "api read_user read_repository write_repository"
	u, _ := url.Parse(authorizeURL)
	q := u.Query()
	q.Set("client_id", clientID())
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", scope)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (p *Provider) ExchangeCode(code, redirectURI string) (*types.OAuthToken, error) {
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	resp, err := p.api().R().
		SetFormData(map[string]string{
			"client_id":     clientID(),
			"client_secret": clientSecret(),
			"code":          code,
			"grant_type":    "authorization_code",
			"redirect_uri":  redirectURI,
		}).
		SetResult(&tokenResp).
		Post(tokenURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK || tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("gitlab token exchange failed: %d %s", resp.StatusCode(), resp.String())
	}
	return &types.OAuthToken{
		AccessToken: tokenResp.AccessToken, RefreshToken: tokenResp.RefreshToken,
		ExpiresIn: tokenResp.ExpiresIn, Scope: tokenResp.Scope,
	}, nil
}

func (p *Provider) RefreshToken(refreshToken string) (*types.OAuthToken, error) {
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	resp, err := p.api().R().
		SetFormData(map[string]string{
			"client_id":     clientID(),
			"client_secret": clientSecret(),
			"refresh_token": refreshToken,
			"grant_type":    "refresh_token",
		}).
		SetResult(&tokenResp).
		Post(tokenURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK || tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("gitlab refresh failed: %d", resp.StatusCode())
	}
	return &types.OAuthToken{
		AccessToken: tokenResp.AccessToken, RefreshToken: tokenResp.RefreshToken,
		ExpiresIn: tokenResp.ExpiresIn,
	}, nil
}

func (p *Provider) Revoke(accessToken string) error {
	_, err := p.api().R().
		SetFormData(map[string]string{
			"client_id": clientID(), "client_secret": clientSecret(), "token": accessToken,
		}).
		Post("https://gitlab.com/oauth/revoke")
	return err
}

func (p *Provider) GetUser(accessToken string) (*types.UserProfile, error) {
	var raw struct {
		Username  string `json:"username"`
		WebURL    string `json:"web_url"`
		AvatarURL string `json:"avatar_url"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).SetResult(&raw).Get(apiBase + "/user")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("gitlab get user: %d", resp.StatusCode())
	}
	// Count owned projects (approximate)
	var projects []json.RawMessage
	_, _ = p.api().R().SetAuthToken(accessToken).
		SetQueryParams(map[string]string{"owned": "true", "per_page": "1"}).
		SetResult(&projects).
		Get(apiBase + "/projects")
	return &types.UserProfile{
		Login: raw.Username, HTMLURL: raw.WebURL, AvatarURL: raw.AvatarURL,
		Provider: types.ProviderGitLab,
	}, nil
}

type glProject struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	Description       string `json:"description"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	WebURL            string `json:"web_url"`
	StarCount         int    `json:"star_count"`
	ForksCount        int    `json:"forks_count"`
	LastActivityAt    string `json:"last_activity_at"`
	Visibility        string `json:"visibility"`
	Archived          bool   `json:"archived"`
	ForkedFromProject *struct{} `json:"forked_from_project"`
	Namespace         struct {
		Path string `json:"path"`
	} `json:"namespace"`
}

func mapProject(r glProject) types.Repo {
	return types.Repo{
		ID: r.ID, Name: r.Name, FullName: r.PathWithNamespace, Description: r.Description,
		UpdatedAt: r.LastActivityAt, HTMLURL: r.WebURL, CloneURL: r.HTTPURLToRepo,
		Stargazers: r.StarCount, Forks: r.ForksCount, Private: r.Visibility != "public",
		Fork: r.ForkedFromProject != nil, Archived: r.Archived,
		Provider: types.ProviderGitLab, OwnerLogin: r.Namespace.Path,
	}
}

func (p *Provider) ListRepos(accessToken string, page int, visibility, sort, direction string) ([]types.Repo, error) {
	if page < 1 {
		page = 1
	}
	orderBy := "updated_at"
	switch sort {
	case "full_name", "name":
		orderBy = "name"
	case "created":
		orderBy = "created_at"
	}
	params := map[string]string{
		"page": strconv.Itoa(page), "per_page": "10", "owned": "true",
		"order_by": orderBy, "sort": direction,
	}
	if visibility == "public" || visibility == "private" {
		params["visibility"] = visibility
	}
	var raw []glProject
	resp, err := p.api().R().SetAuthToken(accessToken).SetQueryParams(params).SetResult(&raw).
		Get(apiBase + "/projects")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("gitlab list: %d", resp.StatusCode())
	}
	out := make([]types.Repo, 0, len(raw))
	for _, r := range raw {
		out = append(out, mapProject(r))
	}
	return out, nil
}

func (p *Provider) SearchRepos(accessToken, username, query, language, visibility, kind, sort string) ([]types.Repo, error) {
	params := map[string]string{
		"owned": "true", "per_page": "30", "search": query, "order_by": "updated_at", "sort": "desc",
	}
	if visibility == "public" || visibility == "private" {
		params["visibility"] = visibility
	}
	var raw []glProject
	resp, err := p.api().R().SetAuthToken(accessToken).SetQueryParams(params).SetResult(&raw).
		Get(apiBase + "/projects")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("gitlab search: %d", resp.StatusCode())
	}
	out := make([]types.Repo, 0, len(raw))
	for _, r := range raw {
		repo := mapProject(r)
		if kind == "forks" && !repo.Fork {
			continue
		}
		if kind == "sources" && repo.Fork {
			continue
		}
		if kind == "archived" && !repo.Archived {
			continue
		}
		out = append(out, repo)
	}
	return out, nil
}

func projectPath(owner, name string) string {
	return url.PathEscape(owner + "/" + name)
}

func (p *Provider) DeleteRepo(accessToken, owner, name string) error {
	resp, err := p.api().R().SetAuthToken(accessToken).
		Delete(apiBase + "/projects/" + projectPath(owner, name))
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return fmt.Errorf("not found: %s/%s", owner, name)
	}
	if resp.StatusCode() != http.StatusAccepted && resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("gitlab delete: %d", resp.StatusCode())
	}
	return nil
}

func (p *Provider) CreateRepo(accessToken, name, description string, private bool) (*types.Repo, error) {
	vis := "public"
	if private {
		vis = "private"
	}
	var raw glProject
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetBody(map[string]any{"name": name, "description": description, "visibility": vis, "initialize_with_readme": false}).
		SetResult(&raw).
		Post(apiBase + "/projects")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("gitlab create: %d %s", resp.StatusCode(), resp.String())
	}
	r := mapProject(raw)
	return &r, nil
}

func (p *Provider) AuthenticatedCloneURL(accessToken, cloneURL string) string {
	u := strings.TrimPrefix(cloneURL, "https://")
	return "https://oauth2:" + accessToken + "@" + u
}

func (p *Provider) ListLabels(accessToken, owner, name string) ([]types.Label, error) {
	var raw []struct {
		Name        string `json:"name"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).SetResult(&raw).
		Get(apiBase + "/projects/" + projectPath(owner, name) + "/labels?per_page=100")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("gitlab labels: %d", resp.StatusCode())
	}
	out := make([]types.Label, 0, len(raw))
	for _, l := range raw {
		out = append(out, types.Label{Name: l.Name, Color: strings.TrimPrefix(l.Color, "#"), Description: l.Description})
	}
	return out, nil
}

func (p *Provider) ListIssues(accessToken, owner, name string) ([]types.Issue, error) {
	var raw []struct {
		IID       int      `json:"iid"`
		Title     string   `json:"title"`
		Description string `json:"description"`
		State     string   `json:"state"`
		WebURL    string   `json:"web_url"`
		CreatedAt string   `json:"created_at"`
		UpdatedAt string   `json:"updated_at"`
		Author    struct {
			Username string `json:"username"`
		} `json:"author"`
		Labels []string `json:"labels"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetQueryParams(map[string]string{"state": "all", "per_page": "100"}).
		SetResult(&raw).
		Get(apiBase + "/projects/" + projectPath(owner, name) + "/issues")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("gitlab issues: %d", resp.StatusCode())
	}
	out := make([]types.Issue, 0, len(raw))
	for _, i := range raw {
		state := i.State
		if state == "opened" {
			state = "open"
		}
		out = append(out, types.Issue{
			Number: i.IID, Title: i.Title, Body: i.Description, State: state,
			HTMLURL: i.WebURL, Author: i.Author.Username, Labels: i.Labels,
			CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt,
		})
	}
	return out, nil
}

func (p *Provider) ListPullRequests(accessToken, owner, name string) ([]types.PullRequest, error) {
	var raw []struct {
		IID       int    `json:"iid"`
		Title     string `json:"title"`
		Description string `json:"description"`
		State     string `json:"state"`
		WebURL    string `json:"web_url"`
		CreatedAt string `json:"created_at"`
		MergedAt  *string `json:"merged_at"`
		Author    struct {
			Username string `json:"username"`
		} `json:"author"`
		SourceBranch string   `json:"source_branch"`
		TargetBranch string   `json:"target_branch"`
		Labels       []string `json:"labels"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetQueryParams(map[string]string{"state": "all", "per_page": "100"}).
		SetResult(&raw).
		Get(apiBase + "/projects/" + projectPath(owner, name) + "/merge_requests")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("gitlab mrs: %d", resp.StatusCode())
	}
	out := make([]types.PullRequest, 0, len(raw))
	for _, mr := range raw {
		merged := mr.MergedAt != nil
		state := mr.State
		if merged {
			state = "merged"
		} else if state == "opened" {
			state = "open"
		}
		out = append(out, types.PullRequest{
			Number: mr.IID, Title: mr.Title, Body: mr.Description, State: state,
			HTMLURL: mr.WebURL, Author: mr.Author.Username, HeadRef: mr.SourceBranch, BaseRef: mr.TargetBranch,
			Labels: mr.Labels, CreatedAt: mr.CreatedAt, Merged: merged,
		})
	}
	return out, nil
}

func (p *Provider) EnsureLabels(accessToken, owner, name string, labels []types.Label) error {
	for _, l := range labels {
		color := l.Color
		if color != "" && !strings.HasPrefix(color, "#") {
			color = "#" + color
		}
		if color == "" {
			color = "#ededed"
		}
		resp, err := p.api().R().SetAuthToken(accessToken).
			SetBody(map[string]string{"name": l.Name, "color": color, "description": l.Description}).
			Post(apiBase + "/projects/" + projectPath(owner, name) + "/labels")
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusBadRequest {
			// 400 often means exists
			if resp.StatusCode() != http.StatusConflict {
				return fmt.Errorf("gitlab label %s: %d", l.Name, resp.StatusCode())
			}
		}
	}
	return nil
}

func (p *Provider) CreateIssue(accessToken, owner, name string, issue types.Issue) error {
	body := issue.Body
	if issue.Author != "" || issue.HTMLURL != "" {
		body += fmt.Sprintf("\n\n---\n_Imported issue #%d by @%s — %s_", issue.Number, issue.Author, issue.HTMLURL)
	}
	payload := map[string]any{"title": issue.Title, "description": body}
	if len(issue.Labels) > 0 {
		payload["labels"] = strings.Join(issue.Labels, ",")
	}
	resp, err := p.api().R().SetAuthToken(accessToken).SetBody(payload).
		Post(apiBase + "/projects/" + projectPath(owner, name) + "/issues")
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("gitlab create issue: %d %s", resp.StatusCode(), resp.String())
	}
	if issue.State == "closed" {
		var created struct {
			IID int `json:"iid"`
		}
		_ = json.Unmarshal(resp.Body(), &created)
		if created.IID > 0 {
			_, _ = p.api().R().SetAuthToken(accessToken).
				SetBody(map[string]string{"state_event": "close"}).
				Put(apiBase + "/projects/" + projectPath(owner, name) + "/issues/" + strconv.Itoa(created.IID))
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
			"title": pr.Title, "description": body,
			"source_branch": pr.HeadRef, "target_branch": pr.BaseRef,
		}).
		Post(apiBase + "/projects/" + projectPath(owner, name) + "/merge_requests")
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("gitlab create mr: %d %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (p *Provider) WikiCloneURL(accessToken, owner, name string) (string, bool) {
	return fmt.Sprintf("https://gitlab.com/%s/%s.wiki.git", owner, name), true
}
