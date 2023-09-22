package report

import (
	"encoding/json"
	"fmt"
)

type Status string

const (
	Pending   Status = "pending"
	Completed        = "completed"
	Errored          = "errored"
)

// ReportJson is the entire report collected from a file or stdin
type ReportJson struct {
	Environments []ReportEnvironment `json:"environments"`
}

func FromJson(data []byte) (*ReportJson, error) {
	report := ReportJson{}
	err := json.Unmarshal(data, &report)
	return &report, err
}

func (r *ReportJson) ValidateReport() []error {
	errors := []error{}

	for envi, env := range r.Environments {
		if env.Name == "" {
			errors = append(errors, fmt.Errorf("environment %d is empty", envi))
		}
	}

	return errors
}

// ReportConfig is additional config & metadata for the report
type ReportConfig struct {
	ReportDate string `json:"date"`
	BaseUrl    string `json:"base_url"`
}

// ReportEnvironment describes a specific environment being tested upon
type ReportEnvironment struct {
	Name       string      `json:"name"`
	Status     Status      `json:"status"`
	Namespaces []Namespace `json:"namespaces"`
}

func (env *ReportEnvironment) Errors() int {
	totalErrors := 0

	for _, ns := range env.Namespaces {
		for _, s := range ns.Sections {
			totalErrors = totalErrors + len(s.Failures)
		}
	}

	return totalErrors
}

func (env *ReportEnvironment) IsHealthy() bool {
	return env.Errors() == 0
}

type Namespace struct {
	Name     string    `json:"name"`
	Sections []Section `json:"sections"`
}

type Section struct {
	Icon     string   `json:"icon"`
	Name     string   `json:"name"`
	Failures []string `json:"failures"`
}
