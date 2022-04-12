package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type ReportType string
type ReportResult string
type ReportDataType string
type AnnotationType string
type AnnotationSeverity string

const (
	ReportTypeSeurity       ReportType     = "SECURITY"
	ReportResultPassed      ReportResult   = "PASSED"
	ReportResultFailed      ReportResult   = "FAILED"
	ReportDataTypeNumber    ReportDataType = "NUMBER"
	ReportDataTypeText      ReportDataType = "TEXT"
	ReportDataTypeLink      ReportDataType = "LINK"
	AnnotationTypeCodeSmell AnnotationType = "CODE_SMELL"
)

const (
	apiURL = "https://api.bitbucket.org/2.0"
)

type BranchTarget struct {
	Hash string `json:"hash"`
}
type CreateBranchOptions struct {
	Name   string       `json:"name"`
	Target BranchTarget `json:"target"`
}

type CommitFile struct {
	Path    string
	Content []byte
}

type CreateCommitOptions struct {
	Branch      string
	Message     string `json:"message"`
	CommitFiles []CommitFile
}

type Branch struct {
	Name string `json:"name"`
}

type PullRequestRef struct {
	Branch Branch `json:"branch"`
}

type CreatePullRequestOptions struct {
	Title       string         `json:"title"`
	Source      PullRequestRef `json:"source"`
	Destination PullRequestRef `json:"destination"`
}

type ReportDataItem struct {
	Title string         `json:"title"`
	Type  ReportDataType `json:"type"`
	Value interface{}    `json:"value"`
}

type CreateReportOptions struct {
	ID       string           `json:"-"`
	SHA      string           `json:"-"`
	Title    string           `json:"title"`
	Details  string           `json:"details"`
	Type     ReportType       `json:"report_type"`
	Reporter string           `json:"reporter"`
	Link     string           `json:"link"`
	LogoURL  string           `json:"logo_url"`
	Result   ReportResult     `json:"result"`
	Data     []ReportDataItem `json:"data"`
}

type ReportAnnotation struct {
	ExternalID string             `json:"external_id"`
	Title      string             `json:"title"`
	Summary    string             `json:"summary"`
	Type       AnnotationType     `json:"annotation_type"`
	Severity   AnnotationSeverity `json:"severity"`
	Path       string             `json:"path"`
	Line       int                `json:"line"`
}

type Client struct {
	owner      string
	repository string
	token      string
}

func NewClient(owner, repository, token string) *Client {
	return &Client{
		owner:      owner,
		repository: repository,
		token:      token,
	}
}

func (cl *Client) GetBranch(ctx context.Context, name string) (*http.Response, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/refs/branches/%s", apiURL, cl.owner, cl.repository, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(cl.owner, cl.token)

	client := &http.Client{}
	return client.Do(req)
}

func (cl *Client) CreateBranch(ctx context.Context, opts CreateBranchOptions) (*http.Response, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/refs/branches", apiURL, cl.owner, cl.repository)
	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body, error: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create branch, error: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(cl.owner, cl.token)

	client := &http.Client{}
	return client.Do(req)
}

func (cl *Client) CreateCommit(ctx context.Context, opts CreateCommitOptions) (*http.Response, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/src", apiURL, cl.owner, cl.repository)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("message", opts.Message)
	writer.WriteField("branch", opts.Branch)

	for _, file := range opts.CommitFiles {
		part, _ := writer.CreateFormFile(file.Path, filepath.Base(file.Path))
		io.Copy(part, bytes.NewReader(file.Content))
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit, error: %v", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(cl.owner, cl.token)

	client := &http.Client{}
	return client.Do(req)
}

func (cl *Client) CreatePullRequest(ctx context.Context, opts CreatePullRequestOptions) (*http.Response, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/pullrequests", apiURL, cl.owner, cl.repository)
	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body, error: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create branch, error: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(cl.owner, cl.token)

	client := &http.Client{}
	return client.Do(req)
}

func (cl *Client) CreateReport(ctx context.Context, opts CreateReportOptions) (*http.Response, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/commit/%s/reports/%s", apiURL, cl.owner, cl.repository, opts.SHA, opts.ID)
	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body, error: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create report, error: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(cl.owner, cl.token)

	client := &http.Client{}
	return client.Do(req)
}

func (cl *Client) AddAnnotationToReport(ctx context.Context, sha string, id string, annotations []ReportAnnotation) (*http.Response, error) {
	url := fmt.Sprintf("%s/repositories/%s/%s/commit/%s/reports/%s/annotations", apiURL, cl.owner, cl.repository, sha, id)

	body, err := json.Marshal(annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body, error: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to add report annotations, error: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(cl.owner, cl.token)

	client := &http.Client{}
	return client.Do(req)
}
