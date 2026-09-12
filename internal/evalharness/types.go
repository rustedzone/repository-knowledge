package evalharness

import "github.com/rustedzone/repository-knowledge/internal/toolkit"

const (
	FamilyConformance = "conformance"
	FamilyBenchmark   = "outcome_benchmark"

	ConditionConformance = "conformance"
	ConditionControl     = "control"
	ConditionTreatment   = "treatment"

	PreflightContextFull    = toolkit.PreflightContextFull
	PreflightContextCompact = toolkit.PreflightContextCompact
)

type HookPayloadMetrics struct {
	Profile          string `json:"profile"`
	Bytes            int    `json:"bytes"`
	Characters       int    `json:"characters"`
	GenerationMillis int64  `json:"generation_millis"`
	ArtifactCount    int    `json:"artifact_count"`
	RouteCount       int    `json:"route_count"`
}

type Spec struct {
	SchemaVersion           string      `json:"schema_version"`
	Family                  string      `json:"family"`
	ID                      string      `json:"id"`
	Revision                string      `json:"revision"`
	SourceCommit            string      `json:"source_commit,omitempty"`
	Description             string      `json:"description"`
	Fixture                 string      `json:"fixture"`
	Prompt                  string      `json:"prompt"`
	Rubric                  string      `json:"rubric"`
	AllowedChanges          []string    `json:"allowed_changes"`
	ExpectedBehavioralTrace []string    `json:"expected_behavioral_trace,omitempty"`
	Checks                  []CheckSpec `json:"checks"`
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
	SchemaVersion               string              `json:"schema_version"`
	Family                      string              `json:"family"`
	Condition                   string              `json:"condition"`
	CaseID                      string              `json:"case_id"`
	CaseRevision                string              `json:"case_revision"`
	SourceCommit                string              `json:"source_commit,omitempty"`
	Agent                       string              `json:"agent"`
	AgentVersion                string              `json:"agent_version"`
	ModelVersion                string              `json:"model_version"`
	ReasoningConfiguration      string              `json:"reasoning_configuration"`
	RepositoryKnowledgeRevision string              `json:"repository_knowledge_revision"`
	TrialNumber                 int                 `json:"trial_number"`
	PreflightContext            string              `json:"preflight_context,omitempty"`
	HookPayload                 *HookPayloadMetrics `json:"hook_payload,omitempty"`
	Files                       map[string]string   `json:"files"`
}

type PrepareOptions struct {
	CasesRoot                   string
	CaseID                      string
	Output                      string
	Condition                   string
	Agent                       string
	AgentVersion                string
	ModelVersion                string
	ReasoningConfiguration      string
	RepositoryKnowledgeRevision string
	TrialNumber                 int
	PreflightContext            string
}

type PrepareResult struct {
	Family                      string              `json:"family"`
	Condition                   string              `json:"condition"`
	CaseID                      string              `json:"case_id"`
	Revision                    string              `json:"revision"`
	SourceCommit                string              `json:"source_commit,omitempty"`
	Target                      string              `json:"target"`
	Baseline                    string              `json:"baseline"`
	Agent                       string              `json:"agent"`
	AgentVersion                string              `json:"agent_version"`
	ModelVersion                string              `json:"model_version"`
	ReasoningConfiguration      string              `json:"reasoning_configuration"`
	RepositoryKnowledgeRevision string              `json:"repository_knowledge_revision"`
	TrialNumber                 int                 `json:"trial_number"`
	PreflightContext            string              `json:"preflight_context,omitempty"`
	HookPayload                 *HookPayloadMetrics `json:"hook_payload,omitempty"`
	Prompt                      string              `json:"prompt"`
	Rubric                      string              `json:"rubric,omitempty"`
}

type GradeOptions struct {
	CasesRoot        string
	CaseID           string
	Target           string
	DurationMillis   int64
	TokenUsage       *int64
	Usage            *TokenUsageDetails
	SemanticStatus   string
	SemanticScore    *SemanticScore
	SemanticReviewer string
}

const TokenSourceUnavailable = "unavailable"

type TokenUsageDetails struct {
	Source       string `json:"source"`
	InputTokens  *int64 `json:"input_tokens,omitempty"`
	OutputTokens *int64 `json:"output_tokens,omitempty"`
	CachedTokens *int64 `json:"cached_tokens,omitempty"`
	TotalTokens  *int64 `json:"total_tokens,omitempty"`
}

type SemanticScore struct {
	Earned    int `json:"earned"`
	Available int `json:"available"`
}

type CheckResult struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type GradeResult struct {
	SchemaVersion               string              `json:"schema_version"`
	Family                      string              `json:"family"`
	Condition                   string              `json:"condition"`
	CaseID                      string              `json:"case_id"`
	Revision                    string              `json:"revision"`
	SourceCommit                string              `json:"source_commit,omitempty"`
	Target                      string              `json:"target,omitempty"`
	Agent                       string              `json:"agent"`
	AgentVersion                string              `json:"agent_version"`
	ModelVersion                string              `json:"model_version"`
	ReasoningConfiguration      string              `json:"reasoning_configuration"`
	RepositoryKnowledgeRevision string              `json:"repository_knowledge_revision"`
	TrialNumber                 int                 `json:"trial_number"`
	PreflightContext            string              `json:"preflight_context,omitempty"`
	HookPayload                 *HookPayloadMetrics `json:"hook_payload,omitempty"`
	RunDate                     string              `json:"run_date,omitempty"`
	DurationMillis              int64               `json:"duration_millis"`
	TokenUsage                  *int64              `json:"token_usage,omitempty"`
	Usage                       *TokenUsageDetails  `json:"usage,omitempty"`
	PreservedArtifact           string              `json:"preserved_artifact,omitempty"`
	DeterministicStatus         string              `json:"deterministic_status"`
	SemanticStatus              string              `json:"semantic_status"`
	SemanticScore               *SemanticScore      `json:"semantic_score,omitempty"`
	SemanticReviewer            string              `json:"semantic_reviewer,omitempty"`
	OverallStatus               string              `json:"overall_status"`
	Checks                      []CheckResult       `json:"checks"`
	Passed                      int                 `json:"passed"`
	Failed                      int                 `json:"failed"`
	Rubric                      string              `json:"rubric"`
}

type RecordOptions struct {
	ResultsRoot string
	RunDate     string
	Artifact    string
}

type RecordResult struct {
	ResultPath   string      `json:"result_path"`
	ArtifactPath string      `json:"artifact_path"`
	Result       GradeResult `json:"result"`
}
