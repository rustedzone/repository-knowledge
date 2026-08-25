package evalharness

type Spec struct {
	SchemaVersion  string      `json:"schema_version"`
	ID             string      `json:"id"`
	Revision       string      `json:"revision"`
	Description    string      `json:"description"`
	Fixture        string      `json:"fixture"`
	Prompt         string      `json:"prompt"`
	Rubric         string      `json:"rubric"`
	AllowedChanges []string    `json:"allowed_changes"`
	Checks         []CheckSpec `json:"checks"`
}

type CheckSpec struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Path     string   `json:"path,omitempty"`
	Paths    []string `json:"paths,omitempty"`
	Glob     string   `json:"glob,omitempty"`
	Patterns []string `json:"patterns,omitempty"`
	Minimum  int      `json:"minimum,omitempty"`
}

type Baseline struct {
	SchemaVersion string            `json:"schema_version"`
	CaseID        string            `json:"case_id"`
	CaseRevision  string            `json:"case_revision"`
	Files         map[string]string `json:"files"`
}

type PrepareOptions struct {
	CasesRoot string
	CaseID    string
	Output    string
	Agent     string
}

type PrepareResult struct {
	CaseID   string `json:"case_id"`
	Revision string `json:"revision"`
	Target   string `json:"target"`
	Agent    string `json:"agent"`
	Prompt   string `json:"prompt"`
	Rubric   string `json:"rubric"`
}

type GradeOptions struct {
	CasesRoot string
	CaseID    string
	Target    string
}

type CheckResult struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type GradeResult struct {
	CaseID              string        `json:"case_id"`
	Revision            string        `json:"revision"`
	Target              string        `json:"target"`
	DeterministicStatus string        `json:"deterministic_status"`
	SemanticStatus      string        `json:"semantic_status"`
	OverallStatus       string        `json:"overall_status"`
	Checks              []CheckResult `json:"checks"`
	Passed              int           `json:"passed"`
	Failed              int           `json:"failed"`
	Rubric              string        `json:"rubric"`
}
