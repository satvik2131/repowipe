package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"repowipe/config"
	"repowipe/providers"
	"repowipe/types"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const transferTTL = 24 * time.Hour

var (
	transferSem = make(chan struct{}, 2)
	workerOnce  sync.Once
	jobQueue    = make(chan string, 64)
)

// StartTransferWorker starts the in-process transfer worker once.
func StartTransferWorker() {
	workerOnce.Do(func() {
		go func() {
			for id := range jobQueue {
				transferSem <- struct{}{}
				go func(jobID string) {
					defer func() { <-transferSem }()
					runTransferJob(jobID)
				}(id)
			}
		}()
	})
}

// EnqueueTransfer creates a transfer job and queues it.
func EnqueueTransfer(sessionID string, req types.TransferRequest) (*types.TransferJob, error) {
	if !req.Source.Valid() || !req.Destination.Valid() {
		return nil, fmt.Errorf("invalid source or destination provider")
	}
	if req.Source == req.Destination {
		return nil, fmt.Errorf("source and destination must differ")
	}
	if len(req.Repos) == 0 {
		return nil, fmt.Errorf("no repositories specified")
	}

	doc, err := GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if _, ok := doc.Providers[req.Source]; !ok {
		return nil, fmt.Errorf("source provider not linked")
	}
	if _, ok := doc.Providers[req.Destination]; !ok {
		return nil, fmt.Errorf("destination provider not linked")
	}

	results := make([]types.TransferRepoResult, 0, len(req.Repos))
	for _, r := range req.Repos {
		results = append(results, types.TransferRepoResult{Repo: r, Status: "pending"})
	}

	now := time.Now().Unix()
	job := &types.TransferJob{
		ID:          uuid.New().String(),
		SessionID:   sessionID,
		Source:      req.Source,
		Destination: req.Destination,
		Status:      "queued",
		Repos:       results,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := saveTransferJob(job); err != nil {
		return nil, err
	}
	StartTransferWorker()
	select {
	case jobQueue <- job.ID:
	default:
		go func(id string) {
			jobQueue <- id
		}(job.ID)
	}
	return job, nil
}

// GetTransferJob loads a transfer job from Redis.
func GetTransferJob(id string) (*types.TransferJob, error) {
	raw, err := config.RedisClient.Get(config.Ctx, "transfer:"+id).Result()
	if err != nil {
		return nil, err
	}
	var job types.TransferJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func saveTransferJob(job *types.TransferJob) error {
	job.UpdatedAt = time.Now().Unix()
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return config.RedisClient.Set(config.Ctx, "transfer:"+job.ID, b, transferTTL).Err()
}

func runTransferJob(jobID string) {
	job, err := GetTransferJob(jobID)
	if err != nil {
		log.Printf("transfer %s: load failed: %v", jobID, err)
		return
	}
	job.Status = "running"
	_ = saveTransferJob(job)

	srcProv, err := providers.Get(job.Source)
	if err != nil {
		failJob(job, err.Error())
		return
	}
	dstProv, err := providers.Get(job.Destination)
	if err != nil {
		failJob(job, err.Error())
		return
	}

	srcToken, err := EnsureFreshToken(job.SessionID, job.Source)
	if err != nil {
		failJob(job, "source token: "+err.Error())
		return
	}
	dstToken, err := EnsureFreshToken(job.SessionID, job.Destination)
	if err != nil {
		failJob(job, "destination token: "+err.Error())
		return
	}

	workRoot := filepath.Join(os.TempDir(), "transfers", job.ID)
	_ = os.MkdirAll(workRoot, 0o700)
	defer os.RemoveAll(workRoot)

	for i := range job.Repos {
		job.Repos[i].Status = "running"
		_ = saveTransferJob(job)

		result := transferOneRepo(srcProv, dstProv, srcToken, dstToken, job.Repos[i].Repo, workRoot)
		job.Repos[i] = result
		_ = saveTransferJob(job)
	}

	job.Status = "completed"
	for _, r := range job.Repos {
		if r.Status == "failed" {
			job.Status = "completed" // still completed; partial failures are per-repo
			break
		}
	}
	_ = saveTransferJob(job)
}

func failJob(job *types.TransferJob, msg string) {
	job.Status = "failed"
	job.Error = msg
	_ = saveTransferJob(job)
}

func transferOneRepo(
	src, dst providers.Provider,
	srcToken, dstToken, repoRef, workRoot string,
) types.TransferRepoResult {
	result := types.TransferRepoResult{Repo: repoRef, Status: "failed"}
	owner, name := splitRepo(repoRef)
	if name == "" {
		result.Error = "invalid repo name"
		return result
	}

	// Resolve source repo metadata via search/list when owner missing.
	srcRepos, err := src.SearchRepos(srcToken, owner, name, "", "", "", "updated")
	var srcRepo *types.Repo
	if err == nil {
		for i := range srcRepos {
			r := srcRepos[i]
			if strings.EqualFold(r.Name, name) || strings.EqualFold(r.FullName, repoRef) {
				srcRepo = &r
				break
			}
		}
	}
	if srcRepo == nil {
		listed, lerr := src.ListRepos(srcToken, 1, "all", "updated", "desc")
		if lerr == nil {
			for i := range listed {
				r := listed[i]
				if strings.EqualFold(r.Name, name) || strings.EqualFold(r.FullName, repoRef) {
					srcRepo = &r
					break
				}
			}
		}
	}
	if srcRepo == nil {
		// Fallback: construct from owner/name
		if owner == "" {
			result.Error = "could not resolve source repository"
			return result
		}
		srcRepo = &types.Repo{
			Name: name, FullName: owner + "/" + name, OwnerLogin: owner,
			CloneURL: guessCloneURL(src.Name(), owner, name),
		}
	}
	if owner == "" {
		owner = srcRepo.OwnerLogin
	}
	if owner == "" && strings.Contains(srcRepo.FullName, "/") {
		owner = strings.SplitN(srcRepo.FullName, "/", 2)[0]
	}

	created, err := dst.CreateRepo(dstToken, srcRepo.Name, srcRepo.Description, srcRepo.Private)
	if err != nil {
		result.Error = "create dest repo: " + err.Error()
		return result
	}
	result.DestURL = created.HTMLURL

	cloneURL := srcRepo.CloneURL
	if cloneURL == "" {
		cloneURL = guessCloneURL(src.Name(), owner, srcRepo.Name)
	}
	authSrc := src.AuthenticatedCloneURL(srcToken, cloneURL)
	authDst := dst.AuthenticatedCloneURL(dstToken, created.CloneURL)
	if created.CloneURL == "" {
		authDst = dst.AuthenticatedCloneURL(dstToken, guessCloneURL(dst.Name(), created.OwnerLogin, created.Name))
	}

	mirrorDir := filepath.Join(workRoot, srcRepo.Name+".git")
	if err := gitMirror(authSrc, mirrorDir); err != nil {
		result.Error = "git clone mirror: " + err.Error()
		result.Warnings = append(result.Warnings, "git mirror failed")
		return result
	}
	if err := gitPushMirror(mirrorDir, authDst); err != nil {
		result.Error = "git push mirror: " + err.Error()
		return result
	}

	warnings := transferMetadata(src, dst, srcToken, dstToken, owner, srcRepo.Name, created.OwnerLogin, created.Name)
	result.Warnings = warnings
	if len(warnings) > 0 {
		result.Status = "partial"
	} else {
		result.Status = "succeeded"
	}
	return result
}

func transferMetadata(
	src, dst providers.Provider,
	srcToken, dstToken, srcOwner, srcName, dstOwner, dstName string,
) []string {
	var warnings []string

	labels, err := src.ListLabels(srcToken, srcOwner, srcName)
	if err != nil {
		warnings = append(warnings, "export labels: "+err.Error())
	} else if err := dst.EnsureLabels(dstToken, dstOwner, dstName, labels); err != nil {
		warnings = append(warnings, "import labels: "+err.Error())
	}

	issues, err := src.ListIssues(srcToken, srcOwner, srcName)
	if err != nil {
		warnings = append(warnings, "export issues: "+err.Error())
	} else {
		for _, issue := range issues {
			if err := dst.CreateIssue(dstToken, dstOwner, dstName, issue); err != nil {
				warnings = append(warnings, fmt.Sprintf("import issue #%d: %v", issue.Number, err))
			}
		}
	}

	prs, err := src.ListPullRequests(srcToken, srcOwner, srcName)
	if err != nil {
		warnings = append(warnings, "export pull requests: "+err.Error())
	} else {
		for _, pr := range prs {
			if pr.State != "open" {
				// Recreate open PRs only; closed/merged become tracking issues.
				note := types.Issue{
					Title:   fmt.Sprintf("[imported PR #%d] %s", pr.Number, pr.Title),
					Body:    fmt.Sprintf("%s\n\nOriginal: %s\nAuthor: @%s\nBranches: %s → %s\nState: %s", pr.Body, pr.HTMLURL, pr.Author, pr.HeadRef, pr.BaseRef, pr.State),
					HTMLURL: pr.HTMLURL,
					Author:  pr.Author,
				}
				if err := dst.CreateIssue(dstToken, dstOwner, dstName, note); err != nil {
					warnings = append(warnings, fmt.Sprintf("tracking issue for PR #%d: %v", pr.Number, err))
				}
				continue
			}
			if err := dst.CreatePullRequest(dstToken, dstOwner, dstName, pr); err != nil {
				warnings = append(warnings, fmt.Sprintf("import PR #%d as issue: %v", pr.Number, err))
				note := types.Issue{
					Title:   fmt.Sprintf("[imported PR #%d] %s", pr.Number, pr.Title),
					Body:    fmt.Sprintf("%s\n\nOriginal: %s\nCould not recreate PR (branches missing or API error).\nAuthor: @%s\nBranches: %s → %s", pr.Body, pr.HTMLURL, pr.Author, pr.HeadRef, pr.BaseRef),
					HTMLURL: pr.HTMLURL,
					Author:  pr.Author,
				}
				if ierr := dst.CreateIssue(dstToken, dstOwner, dstName, note); ierr != nil {
					warnings = append(warnings, fmt.Sprintf("tracking issue for PR #%d: %v", pr.Number, ierr))
				}
			}
		}
	}

	if wikiURL, ok := src.WikiCloneURL(srcToken, srcOwner, srcName); ok {
		dstWiki, dstOK := dst.WikiCloneURL(dstToken, dstOwner, dstName)
		if !dstOK {
			warnings = append(warnings, "destination does not support git wiki mirror")
		} else {
			wikiDir := filepath.Join(os.TempDir(), "transfers", "wiki-"+srcName)
			_ = os.RemoveAll(wikiDir)
			authWikiSrc := src.AuthenticatedCloneURL(srcToken, wikiURL)
			authWikiDst := dst.AuthenticatedCloneURL(dstToken, dstWiki)
			if err := gitMirror(authWikiSrc, wikiDir); err != nil {
				// Most repos never had a wiki page created, so the guessed
				// .wiki.git URL simply doesn't exist — that's normal, not a failure.
				if !strings.Contains(strings.ToLower(err.Error()), "not found") {
					warnings = append(warnings, "could not copy wiki (git clone failed)")
				}
			} else if err := gitPushMirror(wikiDir, authWikiDst); err != nil {
				warnings = append(warnings, "could not push wiki to destination")
			}
			_ = os.RemoveAll(wikiDir)
		}
	}

	return warnings
}

func gitMirror(url, dir string) error {
	cmd := exec.Command("git", "clone", "--mirror", url, dir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, truncate(string(out), 500))
	}
	return nil
}

func gitPushMirror(dir, url string) error {
	cmd := exec.Command("git", "push", "--mirror", url)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, truncate(string(out), 500))
	}
	return nil
}

func splitRepo(ref string) (owner, name string) {
	ref = strings.TrimSpace(ref)
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		return parts[0], parts[1]
	}
	return "", ref
}

func guessCloneURL(provider types.Provider, owner, name string) string {
	switch provider {
	case types.ProviderGitHub:
		return fmt.Sprintf("https://github.com/%s/%s.git", owner, name)
	case types.ProviderGitLab:
		return fmt.Sprintf("https://gitlab.com/%s/%s.git", owner, name)
	case types.ProviderBitbucket:
		return fmt.Sprintf("https://bitbucket.org/%s/%s.git", owner, name)
	default:
		return ""
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
