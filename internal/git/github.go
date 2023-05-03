package git

import (
	"context"
	"fmt"

	"github.com/google/go-github/v41/github"
	"github.com/weaveworks/weave-policy-validator/internal/types"
	"golang.org/x/oauth2"
)

var (
	githubNewFileMode                          = "100644"
	githubBlobTypeFile                         = "blob"
	githubCheckRunName                         = "Weave Result Report"
	githubCheckRunStatusCompleted              = "completed"
	githubCheckRunConclusionSuccess            = "success"
	githubCheckRunConclusionFailure            = "failure"
	githubCheckRunAnnotationLevel              = "failure"
	githubCheckRunMaxAnnotationsPerRequest int = 50
	githubReportTitle                          = "Weave Result Report"
)

type GithubProvider struct {
	client *github.Client
	owner  string
	repo   string
}

func newGithubProvider(owner, provider, host, repo, token string) (*GithubProvider, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	var client *github.Client

	if provider == GithubEnterprise {
		gheURL := fmt.Sprintf("https://%s", host)
		var err error
		client, err = github.NewEnterpriseClient(gheURL, gheURL, tc)

		if err != nil {
			return nil, err
		}
	} else {
		client = github.NewClient(tc)
	}

	return &GithubProvider{
		client: client,
		owner:  owner,
		repo:   repo,
	}, nil
}

// CreateBranch creates new branch from given commit SHA
func (gh *GithubProvider) CreateBranch(ctx context.Context, branch string, sha string) error {
	branchRef := getRefName(branch)
	_, _, err := gh.client.Git.GetRef(ctx, gh.owner, gh.repo, branchRef)
	if err == nil {
		return nil
	}

	referance := &github.Reference{
		Ref: &branchRef,
		Object: &github.GitObject{
			SHA: &sha,
		},
	}

	_, _, err = gh.client.Git.CreateRef(ctx, gh.owner, gh.repo, referance)
	if err != nil {
		return err
	}

	return nil
}

// CreateCommit creates new commit
func (gh *GithubProvider) CreateCommit(ctx context.Context, branch, message string, files []*types.File) error {
	entries := make([]*github.TreeEntry, 0)
	for _, file := range files {
		content, err := file.Content()
		if err != nil {
			return fmt.Errorf("failed to get file content, file: %s, error: %v", file.Path, err)
		}
		entries = append(entries, &github.TreeEntry{
			Path:    &file.Path,
			Mode:    &githubNewFileMode,
			Type:    &githubBlobTypeFile,
			Content: &content,
		})
	}

	branchOBJ, _, err := gh.client.Repositories.GetBranch(ctx, gh.owner, gh.repo, branch, false)
	if err != nil {
		return fmt.Errorf("failed to get branch: %s, error: %v", branch, err)
	}

	lastTree := branchOBJ.GetCommit().Commit.GetTree()
	tree, _, err := gh.client.Git.CreateTree(ctx, gh.owner, gh.repo, *lastTree.SHA, entries)
	if err != nil {
		return fmt.Errorf("failed to create tree, error: %v", err)
	}

	commit := &github.Commit{
		Message: &message,
		Tree:    tree,
		Parents: []*github.Commit{
			{
				SHA: branchOBJ.GetCommit().SHA,
			},
		},
	}

	newCommit, _, err := gh.client.Git.CreateCommit(ctx, gh.owner, gh.repo, commit)
	if err != nil {
		return fmt.Errorf("failed to create commit, error: %v", err)
	}

	branchRef := getRefName(branch)
	ref := &github.Reference{
		Ref: &branchRef,
		Object: &github.GitObject{
			SHA: newCommit.SHA,
		},
	}

	if _, _, err := gh.client.Git.UpdateRef(ctx, gh.owner, gh.repo, ref, true); err != nil {
		return fmt.Errorf("failed to update branch with the new commit, error: %v", err)
	}

	return nil
}

// CreatePullRequest creates new pull request
func (gh *GithubProvider) CreatePullRequest(ctx context.Context, source, target, title, description string) (*string, error) {
	listOpts := &github.PullRequestListOptions{
		Base: target,
		Head: source,
	}

	pulls, _, err := gh.client.PullRequests.List(ctx, gh.owner, gh.repo, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests, error: %v", err)
	}

	if len(pulls) > 0 {
		return pulls[0].HTMLURL, nil
	}

	createOpts := &github.NewPullRequest{
		Title: &title,
		Body:  &description,
		Base:  &target,
		Head:  &source,
	}

	pull, _, err := gh.client.PullRequests.Create(ctx, gh.owner, gh.repo, createOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request, error: %v", err)
	}

	return pull.HTMLURL, nil
}

// CreateReport creates github checkrun
func (gh *GithubProvider) CreateReport(ctx context.Context, sha string, result types.Result) error {
	var conclusion string

	if result.ViolationCount == 0 {
		conclusion = githubCheckRunConclusionSuccess
	} else {
		conclusion = githubCheckRunConclusionFailure
	}

	summary := result.MarkdowSummary()
	options := github.CreateCheckRunOptions{
		Name:       githubCheckRunName,
		Conclusion: &conclusion,
		Status:     &githubCheckRunStatusCompleted,
		Output: &github.CheckRunOutput{
			Title:   &githubReportTitle,
			Summary: &summary,
		},
		HeadSHA: sha,
	}

	annotations := []*github.CheckRunAnnotation{}
	for i := range result.Violations {
		violation := result.Violations[i]
		annotations = append(annotations, &github.CheckRunAnnotation{
			Path:            &violation.Location.Path,
			StartLine:       &violation.Location.StartLine,
			EndLine:         &violation.Location.EndLine,
			Title:           &violation.Policy.Name,
			Message:         &violation.Message,
			AnnotationLevel: &githubCheckRunAnnotationLevel,
		})
	}

	annotationCount := len(annotations)
	annotationLimit := min(annotationCount, githubCheckRunMaxAnnotationsPerRequest)

	if annotationCount > 0 {
		options.Output.Annotations = annotations[:annotationLimit]
	}

	checkrun, _, err := gh.client.Checks.CreateCheckRun(ctx, gh.owner, gh.repo, options)
	if err != nil {
		return fmt.Errorf("failed to create checkrun, error: %v", err)
	}

	if len(annotations) > githubCheckRunMaxAnnotationsPerRequest {
		for i := githubCheckRunMaxAnnotationsPerRequest; i < len(annotations); i += githubCheckRunMaxAnnotationsPerRequest {
			annotationLimit := min(annotationCount, i+githubCheckRunMaxAnnotationsPerRequest)
			options := github.UpdateCheckRunOptions{
				Name: *checkrun.Name,
				Output: &github.CheckRunOutput{
					Title:       checkrun.Output.Title,
					Summary:     checkrun.Output.Summary,
					Annotations: annotations[i : i+annotationLimit],
				},
			}
			_, _, err := gh.client.Checks.UpdateCheckRun(ctx, gh.owner, gh.repo, *checkrun.ID, options)
			if err != nil {
				return fmt.Errorf("failed to update checkrun, error: %v", err)
			}
		}
	}
	return nil
}
