package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/MagalixTechnologies/weave-iac-validator/internal/types"
)

const (
	Github       string = "github"
	Gitlab       string = "gitlab"
	Bitbucket    string = "bitbucket"
	branchPrefix string = "weave-fix-"
)

type Provider interface {
	CreateBranch(ctx context.Context, name string, sha string) error
	CreateCommit(ctx context.Context, branch, message string, files []*types.File) error
	CreatePullRequest(ctx context.Context, source, target, title, description string) (*string, error)
	CreateReport(ctx context.Context, sha string, result types.Result) error
}

type GitRepository struct {
	provider Provider
	url      string
	token    string
}

// NewGitRepository get new repository struct
func NewGitRepository(provider, url, token string) (*GitRepository, error) {
	owner, repo, err := parseRepoSlug(url)
	if err != nil {
		return nil, err
	}

	var p Provider
	switch provider {
	case Github:
		p, err = newGithubProvider(owner, repo, token)
	case Gitlab:
		p, err = newGitlabProvider(owner, repo, token)
	case Bitbucket:
		p, err = newBitbucketProvider(owner, repo, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to init provider: %s, error: %v", provider, err)
	}

	return &GitRepository{
		provider: p,
		url:      url,
		token:    token,
	}, nil
}

// OpenPullRequest opens pull request and return its url
func (r *GitRepository) OpenPullRequest(ctx context.Context, base, sha string, files []*types.File) (*string, error) {
	source := branchPrefix + base
	err := r.provider.CreateBranch(ctx, source, sha)
	if err != nil {
		return nil, err
	}

	commitMessage := fmt.Sprintf("fix iac violations of commit %s", sha[:7])
	err = r.provider.CreateCommit(ctx, source, commitMessage, files)
	if err != nil {
		return nil, err
	}

	var remediatedResources int
	for i := range files {
		for j := range files[i].Resources {
			if files[i].Resources[j].Remediated {
				remediatedResources++
			}
		}
	}

	title := fmt.Sprintf("Weave - Remediate violating resources of branch (%s)", base)
	description := fmt.Sprintf("This PR remediates %d violating resource(s) in %d file(s)", remediatedResources, len(files))
	pull, err := r.provider.CreatePullRequest(ctx, source, base, title, description)
	if err != nil {
		return nil, err
	}

	return pull, nil
}

// CreateReport executes the provider's CreateReport
func (r *GitRepository) CreateReport(ctx context.Context, sha string, result types.Result) error {
	return r.provider.CreateReport(ctx, sha, result)
}

// IsRemediationBranch checks if the given branch name is a remediation branch
func IsRemediationBranch(name string) bool {
	return strings.HasPrefix(name, branchPrefix)
}
