package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/MagalixTechnologies/policy-core/validation"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/git"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/kustomization"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/policy"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/types"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/validator"
	"github.com/urfave/cli/v2"
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
	GitRepositoryURL      string
	GitRepositoryToken    string
	GitRepositoryBranch   string
	GitRepositorySHA      string

	GenerateGitProviderReport bool
}

func (c *Config) ValidateGitRepositoryConf() error {
	if c.GitRepositoryProvider == "" {
		return errors.New("missing git-repo-provider value")
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
			Usage:       "path to resources kustomization directory",
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
	entityKustomizer, err := getKustomizer(conf.EntityKustomizeConf)
	if err != nil {
		return fmt.Errorf("failed to init entities kustomizer, error: %v", err)
	}

	policyKustomizer, err := getKustomizer(conf.PoliciesKustomizeConf)
	if err != nil {
		return fmt.Errorf("failed to init policies kustomizer, error: %v", err)
	}

	policySource := policy.NewFilesystemSource(policyKustomizer)
	opaValidator := validation.NewOPAValidator(policySource, false, trigger)
	validator := validator.NewValidator(opaValidator, conf.Remediate)

	var gitrepo *git.GitRepository
	if conf.Remediate || conf.GenerateGitProviderReport {
		gitrepo, err = git.NewGitRepository(
			conf.GitRepositoryProvider,
			conf.GitRepositoryURL,
			conf.GitRepositoryToken,
		)
		if err != nil {
			return err
		}
	}

	files, err := entityKustomizer.ResourceFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get resources, error: %v", err)
	}

	result, err := validator.Validate(ctx, files)
	if err != nil {
		return fmt.Errorf("failed to validate resources, error: %v", err)
	}

	fmt.Println(result.TEXT())

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

func getKustomizer(conf KustomizationConf) (kustomization.Kustomizer, error) {
	kustomizer, err := kustomization.GetKustomizerFromPath(conf.Path)
	if err != nil {
		return nil, err
	}

	if kustomizer.Type() == kustomization.HelmType && conf.HelmValuesFile != "" {
		kustomizer.(*kustomization.Helm).SetValueFile(conf.HelmValuesFile)
	}

	return kustomizer, nil
}

func saveOutputFile(path string, content string) error {
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to save result to the output file, err: %w", err)
	}
	return nil
}
