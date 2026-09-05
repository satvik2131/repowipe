package bitbucket

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"repowipe/providers"
	"repowipe/types"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	authorizeURL = "https://bitbucket.org/site/oauth2/authorize"
	tokenURL     = "https://bitbucket.org/site/oauth2/access_token"
	apiBase      = "https://api.bitbucket.org/2.0"
)

func init() {
	providers.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() types.Provider { return types.ProviderBitbucket }

func clientID() string     { return os.Getenv("BITBUCKET_CLIENT_ID") }
func clientSecret() string { return os.Getenv("BITBUCKET_CLIENT_SECRET") }

func (p *Provider) api() *resty.Client {
	return resty.New().SetHeader("Accept", "application/json")
}

func (p *Provider) AuthorizeURL(redirectURI, state, mode string) (string, error) {
	u, _ := url.Parse(authorizeURL)
	q := u.Query()
	q.Set("client_id", clientID())
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	// Scopes are primarily configured on the Bitbucket OAuth consumer;
	// request them here when the consumer allows dynamic scopes.
	q.Set("scope", "repository repository:write account pullrequest pullrequest:write issue wiki")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (p *Provider) exchange(form map[string]string) (*types.OAuthToken, error) {
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scopes       string `json:"scopes"`
	}
	resp, err := p.api().R().
		SetBasicAuth(clientID(), clientSecret()).
		SetFormData(form).
		SetResult(&tokenResp).
		Post(tokenURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK || tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("bitbucket token failed: %d %s", resp.StatusCode(), resp.String())
	}
	return &types.OAuthToken{
		AccessToken: tokenResp.AccessToken, RefreshToken: tokenResp.RefreshToken,
		ExpiresIn: tokenResp.ExpiresIn, Scope: tokenResp.Scopes,
	}, nil
}

func (p *Provider) ExchangeCode(code, redirectURI string) (*types.OAuthToken, error) {
	return p.exchange(map[string]string{
		"grant_type": "authorization_code", "code": code, "redirect_uri": redirectURI,
	})
}

func (p *Provider) RefreshToken(refreshToken string) (*types.OAuthToken, error) {
	return p.exchange(map[string]string{
		"grant_type": "refresh_token", "refresh_token": refreshToken,
	})
}

func (p *Provider) Revoke(accessToken string) error {
	// Bitbucket Cloud consumer tokens: best-effort no dedicated revoke for all apps.
	_ = accessToken
	return nil
}

func (p *Provider) GetUser(accessToken string) (*types.UserProfile, error) {
	var raw struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Links       struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
		} `json:"links"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).SetResult(&raw).Get(apiBase + "/user")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("bitbucket get user: %d", resp.StatusCode())
	}
	login := raw.Username
	if login == "" {
		login = raw.DisplayName
	}
	return &types.UserProfile{
		Login: login, HTMLURL: raw.Links.HTML.Href, AvatarURL: raw.Links.Avatar.Href,
		Provider: types.ProviderBitbucket,
	}, nil
}

type bbRepo struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
	UpdatedOn   string `json:"updated_on"`
	Links       struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Clone []struct {
			Name string `json:"name"`
			Href string `json:"href"`
		} `json:"clone"`
	} `json:"links"`
	Language string `json:"language"`
	Parent   *struct{} `json:"parent"`
}

func mapRepo(r bbRepo, idx int64) types.Repo {
	clone := ""
	for _, c := range r.Links.Clone {
		if c.Name == "https" {
			clone = c.Href
			break
		}
	}
	parts := strings.SplitN(r.FullName, "/", 2)
	owner := ""
	if len(parts) > 0 {
		owner = parts[0]
	}
	return types.Repo{
		ID: idx, Name: r.Name, FullName: r.FullName, Description: r.Description,
		Language: r.Language, UpdatedAt: r.UpdatedOn, HTMLURL: r.Links.HTML.Href, CloneURL: clone,
		Private: r.IsPrivate, Fork: r.Parent != nil,
		Provider: types.ProviderBitbucket, OwnerLogin: owner,
	}
}

func (p *Provider) ListRepos(accessToken string, page int, visibility, sort, direction string) ([]types.Repo, error) {
	if page < 1 {
		page = 1
	}
	params := map[string]string{
		"page": strconvItoa(page), "pagelen": "10", "role": "owner",
	}
	var result struct {
		Values []bbRepo `json:"values"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).SetQueryParams(params).SetResult(&result).
		Get(apiBase + "/repositories")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("bitbucket list: %d %s", resp.StatusCode(), resp.String())
	}
	out := make([]types.Repo, 0, len(result.Values))
	for i, r := range result.Values {
		repo := mapRepo(r, int64(i+1+(page-1)*10))
		if visibility == "public" && repo.Private {
			continue
		}
		if visibility == "private" && !repo.Private {
			continue
		}
		out = append(out, repo)
	}
	return out, nil
}

func strconvItoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func (p *Provider) SearchRepos(accessToken, username, query, language, visibility, kind, sort string) ([]types.Repo, error) {
	repos, err := p.ListRepos(accessToken, 1, visibility, sort, "desc")
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	out := make([]types.Repo, 0)
	for _, r := range repos {
		if q != "" && !strings.Contains(strings.ToLower(r.Name), q) && !strings.Contains(strings.ToLower(r.Description), q) {
			continue
		}
		if language != "" && !strings.EqualFold(r.Language, language) {
			continue
		}
		if kind == "forks" && !r.Fork {
			continue
		}
		if kind == "sources" && r.Fork {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (p *Provider) DeleteRepo(accessToken, owner, name string) error {
	resp, err := p.api().R().SetAuthToken(accessToken).
		Delete(apiBase + "/repositories/" + owner + "/" + name)
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return fmt.Errorf("not found: %s/%s", owner, name)
	}
	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("bitbucket delete: %d", resp.StatusCode())
	}
	return nil
}

func (p *Provider) CreateRepo(accessToken, name, description string, private bool) (*types.Repo, error) {
	user, err := p.GetUser(accessToken)
	if err != nil {
		return nil, err
	}
	var raw bbRepo
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetBody(map[string]any{
			"scm": "git", "is_private": private, "description": description,
			"name": name,
		}).
		SetResult(&raw).
		Post(apiBase + "/repositories/" + user.Login + "/" + strings.ToLower(name))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("bitbucket create: %d %s", resp.StatusCode(), resp.String())
	}
	r := mapRepo(raw, time.Now().Unix())
	if r.OwnerLogin == "" {
		r.OwnerLogin = user.Login
	}
	return &r, nil
}

func (p *Provider) AuthenticatedCloneURL(accessToken, cloneURL string) string {
	// Bitbucket: https://x-token-auth:{access_token}@bitbucket.org/...
	u := strings.TrimPrefix(cloneURL, "https://")
	// strip any existing userinfo
	if at := strings.Index(u, "@"); at >= 0 {
		u = u[at+1:]
	}
	return "https://x-token-auth:" + accessToken + "@" + u
}

func (p *Provider) ListLabels(accessToken, owner, name string) ([]types.Label, error) {
	// Bitbucket Cloud has no first-class labels like GitHub; return empty.
	return []types.Label{}, nil
}

func (p *Provider) ListIssues(accessToken, owner, name string) ([]types.Issue, error) {
	var result struct {
		Values []struct {
			ID      int    `json:"id"`
			Title   string `json:"title"`
			Content struct {
				Raw string `json:"raw"`
			} `json:"content"`
			State     string `json:"state"`
			CreatedOn string `json:"created_on"`
			UpdatedOn string `json:"updated_on"`
			Reporter  struct {
				Nickname string `json:"nickname"`
				DisplayName string `json:"display_name"`
			} `json:"reporter"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
			} `json:"links"`
		} `json:"values"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).SetResult(&result).
		Get(apiBase + "/repositories/" + owner + "/" + name + "/issues?pagelen=50")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return []types.Issue{}, nil // issues may be disabled
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("bitbucket issues: %d", resp.StatusCode())
	}
	out := make([]types.Issue, 0, len(result.Values))
	for _, i := range result.Values {
		author := i.Reporter.Nickname
		if author == "" {
			author = i.Reporter.DisplayName
		}
		state := "open"
		if i.State == "resolved" || i.State == "closed" || i.State == "invalid" || i.State == "duplicate" {
			state = "closed"
		}
		out = append(out, types.Issue{
			Number: i.ID, Title: i.Title, Body: i.Content.Raw, State: state,
			HTMLURL: i.Links.HTML.Href, Author: author, CreatedAt: i.CreatedOn, UpdatedAt: i.UpdatedOn,
		})
	}
	return out, nil
}

func (p *Provider) ListPullRequests(accessToken, owner, name string) ([]types.PullRequest, error) {
	var result struct {
		Values []struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			State       string `json:"state"`
			CreatedOn   string `json:"created_on"`
			Author      struct {
				Nickname    string `json:"nickname"`
				DisplayName string `json:"display_name"`
			} `json:"author"`
			Source struct {
				Branch struct {
					Name string `json:"name"`
				} `json:"branch"`
			} `json:"source"`
			Destination struct {
				Branch struct {
					Name string `json:"name"`
				} `json:"branch"`
			} `json:"destination"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
			} `json:"links"`
		} `json:"values"`
	}
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetQueryParams(map[string]string{"state": "OPEN,MERGED,DECLINED,SUPERSEDED", "pagelen": "50"}).
		SetResult(&result).
		Get(apiBase + "/repositories/" + owner + "/" + name + "/pullrequests")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("bitbucket prs: %d", resp.StatusCode())
	}
	out := make([]types.PullRequest, 0, len(result.Values))
	for _, pr := range result.Values {
		author := pr.Author.Nickname
		if author == "" {
			author = pr.Author.DisplayName
		}
		state := strings.ToLower(pr.State)
		merged := state == "merged"
		if state == "open" {
			state = "open"
		} else if merged {
			state = "merged"
		} else {
			state = "closed"
		}
		out = append(out, types.PullRequest{
			Number: pr.ID, Title: pr.Title, Body: pr.Description, State: state,
			HTMLURL: pr.Links.HTML.Href, Author: author,
			HeadRef: pr.Source.Branch.Name, BaseRef: pr.Destination.Branch.Name,
			CreatedAt: pr.CreatedOn, Merged: merged,
		})
	}
	return out, nil
}

func (p *Provider) EnsureLabels(accessToken, owner, name string, labels []types.Label) error {
	return nil
}

func (p *Provider) CreateIssue(accessToken, owner, name string, issue types.Issue) error {
	body := issue.Body
	if issue.HTMLURL != "" {
		body += fmt.Sprintf("\n\n---\nImported from %s by @%s", issue.HTMLURL, issue.Author)
	}
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetBody(map[string]any{
			"title":   issue.Title,
			"content": map[string]string{"raw": body},
		}).
		Post(apiBase + "/repositories/" + owner + "/" + name + "/issues")
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("bitbucket create issue: %d %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (p *Provider) CreatePullRequest(accessToken, owner, name string, pr types.PullRequest) error {
	body := pr.Body
	if pr.HTMLURL != "" {
		body += fmt.Sprintf("\n\n---\nImported from %s", pr.HTMLURL)
	}
	resp, err := p.api().R().SetAuthToken(accessToken).
		SetBody(map[string]any{
			"title":       pr.Title,
			"description": body,
			"source":      map[string]any{"branch": map[string]string{"name": pr.HeadRef}},
			"destination": map[string]any{"branch": map[string]string{"name": pr.BaseRef}},
		}).
		Post(apiBase + "/repositories/" + owner + "/" + name + "/pullrequests")
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("bitbucket create pr: %d %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (p *Provider) WikiCloneURL(accessToken, owner, name string) (string, bool) {
	// Bitbucket wiki is not a standard git wiki for all plans; skip by default.
	return "", false
}
