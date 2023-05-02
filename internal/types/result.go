package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/weaveworks/weave-iac-validator/internal/markdown"
	"github.com/weaveworks/weave-iac-validator/internal/sarif"
	sast "gitlab.com/gitlab-org/security-products/analyzers/report/v3"
)

const (
	scannerID      = "weaveworks"
	scannerName    = "Weaveworks"
	scannerURL     = "https://weave.works"
	scannerVendor  = "Weaveworks"
	scannerVersion = "0.0.1"
)

type Policy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
	HowToSolve  string `json:"how_to_solve"`
}

type Entity struct {
	Name      string
	Namespace string
	Kind      string
}

type Location struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type Details struct {
	ViolatingKey     *string
	RecommendedValue interface{}
}

type Violation struct {
	ID       string   `json:"id"`
	Message  string   `json:"message"`
	Policy   Policy   `json:"policy"`
	Entity   Entity   `json:"entity"`
	Details  Details  `json:"-"`
	Location Location `json:"location"`
}

type Result struct {
	Scanned        int         `json:"scanned"`
	ViolationCount int         `json:"violations"`
	Remediated     int         `json:"remediated"`
	Violations     []Violation `json:"items"`
	PullRequestURL *string     `json:"pull_request"`
}

type resultSummary struct {
	Policy     Policy
	Violations int
}

var SARIFSeverityMap = map[string]string{
	"low":    "note",
	"medium": "warning",
	"high":   "error",
}

var SASTSeverityMap = map[string]sast.SeverityLevel{
	"low":    sast.SeverityLevelLow,
	"medium": sast.SeverityLevelMedium,
	"high":   sast.SeverityLevelCritical,
}

// JSON return result in json format
func (r *Result) JSON() (string, error) {
	return tojson(r)
}

// SARIF return result in sarif format
func (r *Result) SARIF() (string, error) {
	report := sarif.New()
	run := report.AddRun(scannerName)
	rules := map[string]*sarif.Rule{}
	for i := range r.Violations {
		violation := r.Violations[i]
		if !matchSecurityCategory(violation.Policy.Category) {
			continue
		}
		if _, ok := rules[violation.Policy.ID]; !ok {
			rule := run.AddRule(
				violation.Policy.ID,
				violation.Policy.Name,
				violation.Policy.Description,
				violation.Policy.HowToSolve,
			)
			rules[violation.Policy.ID] = rule
		}
		ruleResult := run.AddResult(violation.Policy.ID, violation.Message, SARIFSeverityMap[violation.Policy.Severity])
		ruleResult.SetResultLocation(
			violation.Location.Path,
			violation.Location.StartLine,
			violation.Location.EndLine,
		)
	}
	return tojson(report)
}

// SAST return result in sast format
func (r *Result) SAST() (string, error) {
	vulnerabilities := []sast.Vulnerability{}
	for i := range r.Violations {
		violation := r.Violations[i]
		if !matchSecurityCategory(violation.Policy.Category) {
			continue
		}
		vulnerabilities = append(vulnerabilities, sast.Vulnerability{
			Name:        violation.Policy.Name,
			Description: violation.Policy.Description,
			Message:     violation.Message,
			Severity:    sast.SeverityLevel(SASTSeverityMap[violation.Policy.Severity]),
			Category:    sast.CategorySast,
			Solution:    violation.Policy.HowToSolve,
			Scanner: sast.Scanner{
				ID:   scannerID,
				Name: scannerName,
			},
			Identifiers: []sast.Identifier{
				{
					Value: violation.Policy.ID,
					Name:  violation.Policy.Name,
					Type:  sast.IdentifierType(violation.Policy.Category),
				},
			},
			Location: sast.Location{
				File:      violation.Location.Path,
				LineStart: violation.Location.StartLine,
				LineEnd:   violation.Location.EndLine,
				KubernetesResource: &sast.KubernetesResource{
					Name:      violation.Entity.Name,
					Namespace: violation.Entity.Namespace,
					Kind:      violation.Entity.Kind,
				},
			},
		})
	}

	report := sast.NewReport()
	report.Vulnerabilities = vulnerabilities

	now := sast.ScanTime(time.Now())
	report.Scan = sast.Scan{
		Type: sast.CategorySast,
		Scanner: sast.ScannerDetails{
			ID:   scannerID,
			Name: scannerName,
			URL:  scannerURL,
			Vendor: sast.Vendor{
				Name: scannerVendor,
			},
			Version: scannerVersion,
		},
		Analyzer: sast.AnalyzerDetails{
			ID:   scannerID,
			Name: scannerName,
			URL:  scannerURL,
			Vendor: sast.Vendor{
				Name: scannerVendor,
			},
			Version: scannerVersion,
		},
		StartTime: &now,
		EndTime:   &now,
	}

	if r.ViolationCount == 0 {
		report.Scan.Status = sast.StatusSuccess
	} else {
		report.Scan.Status = sast.StatusFailure
	}

	return tojson(report)
}

// TEXT return result in text format
func (r *Result) TEXT() string {
	var output string
	for i := range r.Violations {
		violation := r.Violations[i]
		output += fmt.Sprintln("====================================================================")
		output += fmt.Sprintln("Policy", ":", violation.Policy.Name)
		output += fmt.Sprintln("Category", ":", violation.Policy.Category)
		output += fmt.Sprintln("Severity", ":", violation.Policy.Severity)

		var location string
		if violation.Location.StartLine == violation.Location.EndLine {
			location = fmt.Sprintf("#%d", violation.Location.StartLine)
		} else {
			location = fmt.Sprintf("#%d-%d", violation.Location.StartLine, violation.Location.EndLine)
		}

		output += fmt.Sprintln("File", ":", violation.Location.Path, location)
		output += fmt.Sprintln("Message", ":", violation.Message)
	}
	output += fmt.Sprintln("====================================================================")
	output += fmt.Sprintln("Summary", ":")
	output += fmt.Sprintln("scanned:", r.Scanned, "violations:", r.ViolationCount, "remediated:", r.Remediated)

	return output
}

// MarkdowSummary returns result summary in markdown
func (r *Result) MarkdowSummary() string {
	summaryMap := make(map[string]resultSummary)
	for _, violation := range r.Violations {
		if summary, ok := summaryMap[violation.Policy.ID]; ok {
			summary.Violations++
			summaryMap[violation.Policy.ID] = summary
		} else {
			summaryMap[violation.Policy.ID] = resultSummary{
				Policy:     violation.Policy,
				Violations: 1,
			}
		}
	}

	columns := []string{
		"Policy",
		"Category",
		"Severity",
		"Violations",
	}

	rows := [][]string{}
	for _, item := range summaryMap {
		rows = append(rows, []string{
			item.Policy.Name,
			item.Policy.Category,
			item.Policy.Severity,
			fmt.Sprint(item.Violations),
		})
	}

	md := markdown.New()
	md.Head3("Scanned %d resources, found %d violations", r.Scanned, r.ViolationCount)
	md.Table(columns, rows)

	if r.PullRequestURL != nil {
		md.Paragraph("This PR %s remediates %d violation(s)", *r.PullRequestURL, r.Remediated)
	}

	return md.String()
}

func (r *Result) Print() {
	fmt.Println(r.TEXT())
}

func tojson(in interface{}) (string, error) {
	output, err := json.MarshalIndent(in, "", "\t")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func matchSecurityCategory(category string) bool {
	return strings.Contains(category, "security")
}
