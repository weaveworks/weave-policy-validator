package git

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MagalixTechnologies/weave-iac-validator/internal/types"
	"github.com/MagalixTechnologies/weave-iac-validator/pkg/bitbucket"
)

const (
	reporterName                      string = "Weaveworks"
	reporterLink                      string = "https://weave.works"
	reporterLogoURL                   string = "https://www.weave.works/assets/images/blt37b3eb5030b2601f/weaveworks-logo-name.png"
	bitbucketMaxAnnotationsPerRequest int    = 100
)

type BitbucketProvider struct {
	client *bitbucket.Client
	owner  string
	repo   string
	token  string
}

func newBitbucketProvider(owner, repo, token string) (*BitbucketProvider, error) {
	return &BitbucketProvider{
		client: bitbucket.NewClient(owner, repo, token),
		owner:  owner,
		repo:   repo,
		token:  token,
	}, nil
}

// CreateBranch creates new branch from given commit SHA
func (bb *BitbucketProvider) CreateBranch(ctx context.Context, branch string, sha string) error {
	resp, err := bb.client.GetBranch(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to get branch: %s, error: %v", branch, err)
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	opts := bitbucket.CreateBranchOptions{
		Name: branch,
		Target: bitbucket.BranchTarget{
			Hash: sha,
		},
	}

	resp, err = bb.client.CreateBranch(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create branch, error: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create branch, error: %s", resp.Status)
	}

	return nil
}

// CreateCommit creates new commit
func (bb *BitbucketProvider) CreateCommit(ctx context.Context, branch, message string, files []*types.File) error {
	opts := bitbucket.CreateCommitOptions{
		Message: message,
		Branch:  branch,
	}

	for _, file := range files {
		content, err := file.Content()
		if err != nil {
			return nil
		}
		opts.CommitFiles = append(opts.CommitFiles, bitbucket.CommitFile{
			Path:    file.Path,
			Content: []byte(content),
		})
	}

	resp, err := bb.client.CreateCommit(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create commit, error: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create commit, error: %s", resp.Status)
	}

	return nil
}

// CreatePullRequest creates pull request, not implemented yet
func (bb *BitbucketProvider) CreatePullRequest(ctx context.Context, source, target, title, description string) (*string, error) {
	opts := bitbucket.CreatePullRequestOptions{
		Title: title,
		Source: bitbucket.PullRequestRef{
			Branch: bitbucket.Branch{Name: source},
		},
		Destination: bitbucket.PullRequestRef{
			Branch: bitbucket.Branch{Name: target},
		},
	}

	resp, err := bb.client.CreatePullRequest(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request, error: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create pull request, error: %s", resp.Status)
	}

	return nil, nil
}

// CreateReport creates report
func (bb *BitbucketProvider) CreateReport(ctx context.Context, sha string, result types.Result) error {
	opts := bitbucket.CreateReportOptions{
		ID:       fmt.Sprintf("weave-%s", sha[:7]),
		Title:    "Weaveworks IaC scan",
		Details:  fmt.Sprintf("%d scanned resources, %d has violations", result.Scanned, result.ViolationCount),
		SHA:      sha,
		Type:     bitbucket.ReportTypeSeurity,
		Reporter: reporterName,
		Link:     reporterLink,
		LogoURL:  reporterLogoURL,
		Data: []bitbucket.ReportDataItem{
			{
				Title: "Scanned",
				Type:  bitbucket.ReportDataTypeNumber,
				Value: result.Scanned,
			},
			{
				Title: "Violations",
				Type:  bitbucket.ReportDataTypeNumber,
				Value: result.ViolationCount,
			},
			{
				Title: "Remediated",
				Type:  bitbucket.ReportDataTypeNumber,
				Value: result.Remediated,
			},
		},
	}

	var annotations []bitbucket.ReportAnnotation
	for i := range result.Violations {
		violation := result.Violations[i]
		annotations = append(annotations, bitbucket.ReportAnnotation{
			ExternalID: fmt.Sprintf("%s-%d", opts.ID, i),
			Title:      violation.Policy.Name,
			Summary:    violation.Message,
			Type:       bitbucket.AnnotationTypeCodeSmell,
			Severity:   bitbucket.AnnotationSeverity(strings.ToUpper(violation.Policy.Severity)),
			Path:       violation.Location.Path,
			Line:       violation.Location.StartLine,
		})
	}

	if result.ViolationCount == 0 {
		opts.Result = bitbucket.ReportResultPassed
	} else {
		opts.Result = bitbucket.ReportResultFailed
	}

	resp, err := bb.client.CreateReport(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create report, error: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create report, error: %s", resp.Status)
	}

	annotationCount := len(annotations)
	if annotationCount > 0 {
		for i := 0; i < annotationCount; i += bitbucketMaxAnnotationsPerRequest {
			annotationLimit := min(annotationCount, i+bitbucketMaxAnnotationsPerRequest)
			resp, err = bb.client.AddAnnotationToReport(ctx, opts.SHA, opts.ID, annotations[i:i+annotationLimit])
			if err != nil {
				return fmt.Errorf("failed to create report annotation, error: %v", err)
			}
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				return fmt.Errorf("failed to create report annotation, error: %s", resp.Status)
			}
		}
	}
	return nil
}
