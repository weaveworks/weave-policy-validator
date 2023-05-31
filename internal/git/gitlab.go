package git

import (
	"context"
	"errors"
	"fmt"

	"github.com/weaveworks/weave-policy-validator/internal/types"
	"github.com/xanzy/go-gitlab"
)

type GitlabProvider struct {
	client *gitlab.Client
	id     string
}

func newGitlabProvider(owenr, repo, token string) (*GitlabProvider, error) {
	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, fmt.Errorf("failed to init gitlab client, error: %v", err)
	}

	return &GitlabProvider{
		client: client,
		id:     fmt.Sprintf("%s/%s", owenr, repo),
	}, nil
}

// CreateBranch creates new branch from given commit SHA
func (gl *GitlabProvider) CreateBranch(ctx context.Context, name string, sha string) error {
	_, _, err := gl.client.Branches.GetBranch(gl.id, name)
	if err == nil {
		return nil
	}

	opts := &gitlab.CreateBranchOptions{
		Branch: &name,
		Ref:    &sha,
	}

	_, _, err = gl.client.Branches.CreateBranch(gl.id, opts)
	if err != nil {
		return err
	}

	return nil
}

// CreateCommit creates new commit
func (gl *GitlabProvider) CreateCommit(ctx context.Context, branch, message string, files []*types.File) error {
	opts := &gitlab.CreateCommitOptions{
		Branch:        &branch,
		CommitMessage: &message,
		Actions:       []*gitlab.CommitActionOptions{},
	}

	action := gitlab.FileUpdate
	for _, file := range files {
		content, err := file.Content()
		if err != nil {
			return err
		}

		opts.Actions = append(opts.Actions, &gitlab.CommitActionOptions{
			Action:   &action,
			FilePath: &file.Path,
			Content:  &content,
		})
	}

	_, _, err := gl.client.Commits.CreateCommit(gl.id, opts)
	return err
}

// CreatePullRequest creates new pull request
func (gl *GitlabProvider) CreatePullRequest(ctx context.Context, source, target, title, description string) (*string, error) {
	state := "opened"
	listOpts := &gitlab.ListProjectMergeRequestsOptions{
		SourceBranch: &source,
		TargetBranch: &target,
		State:        &state,
	}

	pulls, _, err := gl.client.MergeRequests.ListProjectMergeRequests(gl.id, listOpts)
	if err != nil {
		return nil, err
	}

	if len(pulls) > 0 {
		return &pulls[0].WebURL, nil
	}

	createOpts := &gitlab.CreateMergeRequestOptions{
		Title:        &title,
		Description:  &description,
		SourceBranch: &source,
		TargetBranch: &target,
	}

	pull, _, err := gl.client.MergeRequests.CreateMergeRequest(gl.id, createOpts)
	if err != nil {
		return nil, err
	}

	return &pull.WebURL, err
}

// CreateReport not implemented
func (gl *GitlabProvider) CreateReport(ctx context.Context, sha string, result types.Result) error {
	return errors.New("not implemented")
}
