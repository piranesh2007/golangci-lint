package printers

import (
	"encoding/json"
	"io"

	"github.com/golangci/golangci-lint/pkg/result"
)

const (
	sarifVersion   = "2.1.0"
	sarifSchemaURI = "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.4.json"
)

type SarifResult struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver struct {
		Name string `json:"name"`
	} `json:"driver"`
}

type sarifResult struct {
	RuleID  string `json:"ruleId"`
	Level   string `json:"level"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI   string `json:"uri"`
	Index int    `json:"index"`
}
type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}

type Sarif struct {
	w io.Writer
}

func NewSarif(w io.Writer) *Sarif {
	return &Sarif{
		w: w,
	}
}

func (p Sarif) Print(issues []result.Issue) error {
	res := SarifResult{}
	res.Version = sarifVersion
	res.Schema = sarifSchemaURI
	res.Runs = []sarifRun{}

	toolMap := map[string][]result.Issue{}

	for i := range issues {
		issue := issues[i]
		linter := issue.FromLinter
		toolMap[linter] = append(toolMap[linter], issue)
	}

	for curtool, issues := range toolMap {
		tool := sarifTool{}
		tool.Driver.Name = curtool
		sr := sarifRun{}
		sr.Tool = tool

		for i := range issues {
			issue := issues[i]
			severity := issue.Severity
			// set default to warning
			if severity == "" {
				severity = "warning"
			}

			physLoc := sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: issue.FilePath()},
				Region:           sarifRegion{StartLine: issue.Line(), StartColumn: issue.Column()},
			}
			loc := sarifLocation{PhysicalLocation: physLoc}

			curResult := sarifResult{
				RuleID: issue.Text,
				Level:  severity,
				Message: struct {
					Text string "json:\"text\""
				}{Text: issue.Text},
				Locations: []sarifLocation{loc},
			}

			sr.Results = append(sr.Results, curResult)
		}
		res.Runs = append(res.Runs, sr)
	}

	return json.NewEncoder(p.w).Encode(res)
}
