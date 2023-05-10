package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/weaveworks/policy-agent/pkg/policy-core/validation"
	"github.com/weaveworks/weave-policy-validator/internal/git"
	"github.com/weaveworks/weave-policy-validator/internal/policy"
	"github.com/weaveworks/weave-policy-validator/internal/source"
	"github.com/weaveworks/weave-policy-validator/internal/trie"
	"github.com/weaveworks/weave-policy-validator/internal/types"
	"github.com/weaveworks/weave-policy-validator/internal/validator"
)

const (
	trigger string = "iac"
)

type KustomizationConf struct {
	Path           string
	HelmValuesFile string
}

type Config struct {
	EntityKustomizeConf   KustomizationConf
	PoliciesKustomizeConf KustomizationConf

	// output config
	NoExitError     bool
	SASTOutputFile  string
	SARIFOutputFile string
	JSONOutputFile  string

	// remediation config
	Remediate bool

	// git repo config
	GitRepositoryProvider string
	GitRepositoryHost     string
	GitRepositoryURL      string
	GitRepositoryToken    string
	GitRepositoryBranch   string
	GitRepositorySHA      string

	// azure config
	AzureProject string

	GenerateGitProviderReport bool
}

func (c *Config) ValidateGitRepositoryConf() error {
	if c.GitRepositoryProvider == "" {
		return errors.New("missing git-repo-provider value")
	}
	if c.GitRepositoryHost == "" && c.GitRepositoryProvider == git.GithubEnterprise {
		return errors.New("missing git-repo-host value")
	}
	if c.GitRepositoryURL == "" {
		return errors.New("missing git-repo-url value")
	}
	if c.GitRepositoryBranch == "" {
		return errors.New("missing git-repo-branch value")
	}
	if c.GitRepositorySHA == "" {
		return errors.New("missing git-repo-sha value")
	}
	if c.GitRepositoryToken == "" {
		return errors.New("missing git-repo-token value")
	}
	if c.GitRepositoryProvider == "azure-devops" && c.AzureProject == "" {
		return errors.New("missing azure project value")
	}
	return nil
}

func main() {
	conf := Config{}
	app := cli.NewApp()

	app.Name = "Weave IaC Validator"
	app.Usage = "validate validate kubernetes resources"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "path",
			Usage:       "path to scan resources from",
			Destination: &conf.EntityKustomizeConf.Path,
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "helm-values-file",
			Usage:       "path to resources helm values file",
			Destination: &conf.EntityKustomizeConf.HelmValuesFile,
		},
		&cli.StringFlag{
			Name:        "policies-path",
			Usage:       "path to policies kustomization directory",
			Required:    true,
			Destination: &conf.PoliciesKustomizeConf.Path,
		},
		&cli.StringFlag{
			Name:        "policies-helm-values-file",
			Usage:       "path to policies helm values file",
			Destination: &conf.PoliciesKustomizeConf.HelmValuesFile,
		},
		&cli.StringFlag{
			Name:        "git-repo-provider",
			Usage:       "git repository provider",
			Destination: &conf.GitRepositoryProvider,
			EnvVars:     []string{"WEAVE_REPO_PROVIDER"},
		},
		&cli.StringFlag{
			Name:        "git-repo-host",
			Usage:       "git repository host",
			Destination: &conf.GitRepositoryHost,
			EnvVars:     []string{"WEAVE_REPO_HOST"},
		},
		&cli.StringFlag{
			Name:        "git-repo-url",
			Usage:       "git repository url",
			Destination: &conf.GitRepositoryURL,
			EnvVars:     []string{"WEAVE_REPO_URL"},
		},
		&cli.StringFlag{
			Name:        "git-repo-branch",
			Usage:       "git repository branch",
			Destination: &conf.GitRepositoryBranch,
			EnvVars:     []string{"WEAVE_REPO_BRANCH"},
		},
		&cli.StringFlag{
			Name:        "git-repo-sha",
			Usage:       "git repository commit sha",
			Destination: &conf.GitRepositorySHA,
			EnvVars:     []string{"WEAVE_REPO_SHA"},
		},
		&cli.StringFlag{
			Name:        "git-repo-token",
			Usage:       "git repository token",
			Destination: &conf.GitRepositoryToken,
			EnvVars:     []string{"WEAVE_REPO_TOKEN"},
		},
		&cli.StringFlag{
			Name:        "azure-project",
			Usage:       "azure project name",
			Destination: &conf.AzureProject,
			EnvVars:     []string{"AZURE_PROJECT"},
		},
		&cli.PathFlag{
			Name:        "sast",
			Usage:       "save result as gitlab sast format",
			Destination: &conf.SASTOutputFile,
		},
		&cli.PathFlag{
			Name:        "sarif",
			Usage:       "save result as sarif format",
			Destination: &conf.SARIFOutputFile,
		},
		&cli.PathFlag{
			Name:        "json",
			Usage:       "save result as json format",
			Destination: &conf.JSONOutputFile,
		},
		&cli.BoolFlag{
			Name:        "generate-git-report",
			Usage:       "generate git report if supported",
			Value:       false,
			EnvVars:     []string{"WEAVE_GENERATE_GIT_PROVIDER_REPORT"},
			Destination: &conf.GenerateGitProviderReport,
		},
		&cli.BoolFlag{
			Name:        "remediate",
			Usage:       "auto remediate resources if possible",
			Value:       false,
			Destination: &conf.Remediate,
		},
		&cli.BoolFlag{
			Name:        "no-exit-error",
			Usage:       "exit with no error",
			Value:       false,
			Destination: &conf.NoExitError,
		},
	}

	app.Before = func(context *cli.Context) error {
		var err error
		if conf.EntityKustomizeConf.Path, err = filepath.Abs(conf.EntityKustomizeConf.Path); err != nil {
			return fmt.Errorf("invalid entities path: %w", err)
		}
		if conf.PoliciesKustomizeConf.Path, err = filepath.Abs(conf.PoliciesKustomizeConf.Path); err != nil {
			return fmt.Errorf("invalid policies path: %w", err)
		}
		if conf.Remediate || conf.GenerateGitProviderReport {
			if err := conf.ValidateGitRepositoryConf(); err != nil {
				return err
			}
		}
		return nil
	}

	app.Action = func(context *cli.Context) error {
		return App(context.Context, conf)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func App(ctx context.Context, conf Config) error {
	files, err := scan(ctx, conf.EntityKustomizeConf)
	if err != nil {
		return fmt.Errorf("failed to get resources, error: %v", err)
	}

	policySource, err := getSource(conf.PoliciesKustomizeConf)
	if err != nil {
		return fmt.Errorf("failed to init policies source, error: %v", err)
	}

	policySource := policy.NewFilesystemSource(policySource)
	// sinks := []domain.PolicyValidationSink{}
	opaValidator := validation.NewOPAValidator(policySource, false, "", "", "", false)
	validator := validator.NewValidator(opaValidator, conf.Remediate)

	var gitrepo *git.GitRepository
	if conf.Remediate || conf.GenerateGitProviderReport {
		gitrepo, err = git.NewGitRepository(
			conf.GitRepositoryProvider,
			conf.GitRepositoryHost,
			conf.GitRepositoryURL,
			conf.GitRepositoryToken,
			conf.AzureProject,
		)
		if err != nil {
			return err
		}
	}

	result, err := validator.Validate(ctx, files)
	if err != nil {
		return fmt.Errorf("failed to validate resources, error: %v", err)
	}

	result.Print()

	if conf.Remediate && !git.IsRemediationBranch(conf.GitRepositoryBranch) {
		var remediatedFiles []*types.File
		for _, file := range files {
			if file.Remediated {
				remediatedFiles = append(remediatedFiles, file)
			}
		}
		if len(remediatedFiles) > 0 {
			pullRequestURL, err := gitrepo.OpenPullRequest(ctx, conf.GitRepositoryBranch, conf.GitRepositorySHA, remediatedFiles)
			if err != nil {
				return err
			}
			if pullRequestURL != nil {
				result.PullRequestURL = pullRequestURL
			}
		}
	}

	if conf.GenerateGitProviderReport {
		err = gitrepo.CreateReport(ctx, conf.GitRepositorySHA, *result)
		if err != nil {
			return err
		}
	}

	if conf.SARIFOutputFile != "" {
		sarif, err := result.SARIF()
		if err != nil {
			return fmt.Errorf("failed to export result as sarif, error: %v", err)
		}
		err = saveOutputFile(conf.SARIFOutputFile, sarif)
		if err != nil {
			return err
		}
	}

	if conf.SASTOutputFile != "" {
		sast, err := result.SAST()
		if err != nil {
			return fmt.Errorf("failed to export result as sast, error: %v", err)
		}
		err = saveOutputFile(conf.SASTOutputFile, sast)
		if err != nil {
			return err
		}
	}

	if conf.JSONOutputFile != "" {
		js, err := result.JSON()
		if err != nil {
			return fmt.Errorf("failed to export result as json, error: %v", err)
		}
		err = saveOutputFile(conf.JSONOutputFile, js)
		if err != nil {
			return err
		}
	}

	if result.ViolationCount > 0 {
		if conf.NoExitError {
			os.Exit(0)
		}
		os.Exit(1)
	}

	return nil
}

func getSource(conf KustomizationConf) (source.Source, error) {
	source, err := source.GetSourceFromPath(conf.Path)
	if err != nil {
		return nil, err
	}

	if source.Type() == source.HelmType && conf.HelmValuesFile != "" {
		source.(*source.Helm).SetValueFile(conf.HelmValuesFile)
	}

	return source, nil
}

func scan(ctx context.Context, conf KustomizationConf) ([]*types.File, error) {
	var paths []string
	err := filepath.Walk(conf.Path, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	t := trie.NewTrie()

	var files []*types.File
	for _, path := range paths {
		if t.Search(filepath.Dir(path)) {
			t.Insert(path)
			continue
		}

		conf.Path = path
		if source, err := getSource(conf); err == nil {
			t.Insert(path)
			kfiles, err := source.ResourceFiles(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get resources, path: %s, error: %v", path, err)
			}
			files = append(files, kfiles...)
		}
	}
	return files, nil
}

func saveOutputFile(path string, content string) error {
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to save result to the output file, err: %w", err)
	}
	return nil
}
