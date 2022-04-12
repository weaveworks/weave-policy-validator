package sarif

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
