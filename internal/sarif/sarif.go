package sarif

const (
	schema  = "http://json.schemastore.org/sarif-2.1.0-rtm.4"
	version = "2.1.0"
)

type Report struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []*Run `json:"runs"`
}

type Run struct {
	Tool    Tool      `json:"tool"`
	Results []*Result `json:"results"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name    string `json:"name"`
	Rules   []Rule `json:"rules"`
	ruleMap map[string]bool
}

type Text struct {
	Text string `json:"text"`
}

type Rule struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ShortDescription Text   `json:"shortDescription"`
	FullDescription  Text   `json:"fullDescription"`
	Help             Text   `json:"help"`
}

type Result struct {
	RuleID              string              `json:"ruleId"`
	Message             Text                `json:"message"`
	Locations           []Location          `json:"locations"`
	Level               string              `json:"level"`
	PartialFingerprints PartialFingerprints `json:"partialFingerprints"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

type ArtifactLocation struct {
	URI string `json:"uri"`
}

type Region struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine"`
	StartColumn int `json:"startColumn"`
	EndColumn   int `json:"endColumn"`
}

type PartialFingerprints struct {
	PrimaryLocationLineHash string `json:"primaryLocationLineHash"`
}

// New creates a new report
func New() *Report {
	return &Report{
		Schema:  schema,
		Version: version,
		Runs:    []*Run{},
	}
}

// AddRun adds a new run to the report
func (rprt *Report) AddRun(name string) *Run {
	run := &Run{
		Tool: Tool{
			Driver: Driver{
				Name:    name,
				Rules:   []Rule{},
				ruleMap: map[string]bool{},
			},
		},
		Results: []*Result{},
	}
	rprt.Runs = append(rprt.Runs, run)
	return run
}

// AddRule adds a new rule to the report
func (rn *Run) AddRule(id, name, description, help string) *Rule {
	rule := Rule{
		ID:   id,
		Name: name,
		ShortDescription: Text{
			Text: name,
		},
		FullDescription: Text{
			Text: description,
		},
		Help: Text{
			Text: help,
		},
	}
	rn.Tool.Driver.Rules = append(rn.Tool.Driver.Rules, rule)
	return &rule
}

// AddResult adds a new result to the report
func (rn *Run) AddResult(ruleID, message, level string) *Result {
	result := &Result{
		RuleID: ruleID,
		Message: Text{
			Text: message,
		},
		Level: level,
	}
	rn.Results = append(rn.Results, result)
	return result
}

// SetResultLocation sets violation location
func (rs *Result) SetResultLocation(file string, startLine, endLine int) *Result {
	location := Location{
		PhysicalLocation: PhysicalLocation{
			ArtifactLocation: ArtifactLocation{
				URI: file,
			},
			Region: Region{
				StartLine:   startLine,
				EndLine:     endLine,
				StartColumn: 1,
				EndColumn:   1,
			},
		},
	}
	rs.Locations = []Location{location}
	return rs
}
