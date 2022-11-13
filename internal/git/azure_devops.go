package git

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/MagalixTechnologies/weave-iac-validator/internal/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

var (
	oldObjectId = "0000000000000000000000000000000000000000"
	author      = "weave-iac-validator"
)

type AzureDevopsProvider struct {
	client          git.Client
	project         string
	repo            string
	organizationUrl string
}

func newAzureGitopsProvider(organizationUrl, project, repo, token string) (*AzureDevopsProvider, error) {

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(organizationUrl, token)
	ctx := context.Background()

	client, err := git.NewClient(ctx, connection)
	if err != nil {
		log.Fatal(err)
	}
	return &AzureDevopsProvider{
		client:          client,
		organizationUrl: organizationUrl,
		project:         project,
		repo:            repo,
	}, nil
}

// GetBranchRef forms the full branch name
func (az *AzureDevopsProvider) GetBranchRef(branch string) string {
	return fmt.Sprintf("refs/heads/%s", branch)
}

// GetBranch Gets the branch with name
func (az *AzureDevopsProvider) GetBranch(ctx context.Context, branch string) (*git.GitBranchStats, error) {
	return az.client.GetBranch(ctx, git.GetBranchArgs{
		RepositoryId: &az.repo,
		Name:         &branch,
		Project:      &az.project,
		BaseVersionDescriptor: &git.GitVersionDescriptor{
			Version: &branch,
		},
	})
}

// CreateBranch creates new branch from given commit SHA
func (az *AzureDevopsProvider) CreateBranch(ctx context.Context, branch string, sha string) error {
	// make sure branch is not existed
	_, err := az.GetBranch(ctx, branch)
	if err == nil {
		return nil
	}
	locked := true
	branchName := az.GetBranchRef(branch)
	args := git.UpdateRefsArgs{
		RefUpdates: &[]git.GitRefUpdate{
			{
				IsLocked:    &locked,
				Name:        &branchName,
				OldObjectId: &oldObjectId,
				NewObjectId: &sha,
			},
		},
		Project:      &az.project,
		RepositoryId: &az.repo,
	}
	_, err = az.client.UpdateRefs(ctx, args)

	if err != nil {
		return fmt.Errorf("failed to create branch, name %s, commit %s due to %w", branch, sha, err)
	}
	return nil
}

// CreateCommit creates new commit
func (az *AzureDevopsProvider) CreateCommit(ctx context.Context, branch, message string, files []*types.File) error {
	changes := make([]*git.GitChange, 0)

	for _, file := range files {
		content, err := file.Content()
		if err != nil {
			return fmt.Errorf("failed to get file content, file: %s, error: %v", file.Path, err)
		}

		changes = append(changes, &git.GitChange{
			ChangeType: &git.VersionControlChangeTypeValues.Edit,
			Item: &git.GitLastChangeItem{
				Path: &file.Path,
			},
			NewContent: &git.ItemContent{
				Content:     &content,
				ContentType: &git.ItemContentTypeValues.RawText,
			},
		})
	}

	changesInterface := make([]interface{}, len(changes))
	for i, v := range changes {
		changesInterface[i] = v
	}

	branchName := az.GetBranchRef(branch)

	branchObject, err := az.GetBranch(ctx, branch)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("couldn't get branch %s", branch), err)
	}

	args := git.CreatePushArgs{
		Push: &git.GitPush{
			Commits: &[]git.GitCommitRef{
				{
					Author: &git.GitUserDate{
						Name: &author,
					},
					Changes: &changesInterface,
					Comment: &message,
				},
			},
			RefUpdates: &[]git.GitRefUpdate{
				{
					Name:        &branchName,
					OldObjectId: branchObject.Commit.CommitId,
				},
			},
		},
		RepositoryId: &az.repo,
		Project:      &az.project,
	}

	_, err = az.client.CreatePush(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to create commit, error: %v", err)
	}

	return nil
}

// CreatePullRequest creates new pull request
func (az *AzureDevopsProvider) CreatePullRequest(ctx context.Context, source, target, title, description string) (*string, error) {
	source = az.GetBranchRef(source)
	target = az.GetBranchRef(target)
	listArgs := git.GetPullRequestsArgs{
		RepositoryId: &az.repo,
		Project:      &az.project,
		SearchCriteria: &git.GitPullRequestSearchCriteria{
			SourceRefName: &source,
			TargetRefName: &target,
		},
	}
	pulls, err := az.client.GetPullRequests(ctx, listArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests, error: %v", err)
	}
	if len(*pulls) > 0 {
		return (*pulls)[0].Repository.RemoteUrl, nil
	}
	args := git.CreatePullRequestArgs{
		GitPullRequestToCreate: &git.GitPullRequest{
			Title:         &title,
			Description:   &description,
			SourceRefName: &source,
			TargetRefName: &target,
		},
		RepositoryId: &az.repo,
		Project:      &az.project,
	}
	pullRequest, err := az.client.CreatePullRequest(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request, error: %v", err)
	}
	return pullRequest.Url, nil
}

// CreateReport not implemented
func (az *AzureDevopsProvider) CreateReport(ctx context.Context, sha string, result types.Result) error {
	return errors.New("not implemented")
}
