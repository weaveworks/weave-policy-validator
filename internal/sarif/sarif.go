package sarif

const (
	schema  = "http://json.schemastore.org/sarif-2.1.0-rtm.4"
	version = "2.1.0"
)

// New create new report
func New() *Report {
	return &Report{
		Schema:  schema,
		Version: version,
		Runs:    []*Run{},
	}
}

// AddRun add new run
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

// AddRun add new rule
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

// AddRun add result
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

// SetResultLocation set violation location
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
