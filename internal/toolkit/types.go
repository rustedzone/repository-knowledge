package toolkit

type Change struct {
	Status string `json:"status"`
	Path   string `json:"path"`
}

type ManagedFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ToolkitManifest struct {
	SchemaVersion        string            `json:"schema_version"`
	ToolkitVersion       string            `json:"toolkit_version"`
	Source               string            `json:"source"`
	Ref                  string            `json:"ref"`
	InstalledAt          string            `json:"installed_at"`
	AgentAdapters        []string          `json:"agent_adapters"`
	PreflightModes       map[string]string `json:"preflight_modes,omitempty"`
	AntigravityPreflight string            `json:"antigravity_preflight,omitempty"`
	ManagedFiles         []ManagedFile     `json:"managed_files"`
	Ownership            map[string]string `json:"ownership"`
}

type InstallOptions struct {
	Target               string
	Source               string
	Ref                  string
	AgentAdapters        []string
	AllAgentAdapters     bool
	Update               bool
	AllowDowngrade       bool
	PreflightModes       map[string]string
	AgentPreflight       []string
	AntigravityPreflight string
}

const (
	AntigravityPreflightObserve = "observe"
	AntigravityPreflightStrict  = "strict"
)

type InstallResult struct {
	Action                         string   `json:"action"`
	Target                         string   `json:"target"`
	Version                        string   `json:"version"`
	ManagedFileCount               int      `json:"managed_file_count"`
	RepositoryOwnedFilesCreated    []string `json:"repository_owned_files_created"`
	RepositoryOwnedFilesPreserved  bool     `json:"repository_owned_files_preserved"`
	ObsoleteManagedFilesRemoved    []string `json:"obsolete_managed_files_removed"`
	ModifiedObsoleteFilesPreserved []string `json:"modified_obsolete_managed_files_preserved"`
	AgentHookRegistrations         []string `json:"agent_hook_registrations"`
}

type RebuildResult struct {
	Proposal                      string `json:"proposal"`
	Applied                       bool   `json:"applied"`
	GeneratedInventory            string `json:"generated_inventory,omitempty"`
	ArtifactKind                  string `json:"artifact_kind"`
	SemanticDocumentationComplete bool   `json:"semantic_documentation_complete"`
	NextStep                      string `json:"next_step"`
}

type RepositoryConfig struct {
	SchemaVersion string `json:"schema_version"`
	Repository    struct {
		Name   string   `json:"name"`
		Kind   string   `json:"kind"`
		Owners []string `json:"owners"`
	} `json:"repository"`
	Documentation struct {
		Index              string `json:"index"`
		GeneratedInventory string `json:"generated_inventory"`
	} `json:"documentation"`
	Capabilities []Capability `json:"capabilities"`
	CI           struct {
		Enforcement string `json:"enforcement"`
	} `json:"ci"`
}

type Capability struct {
	ID            string   `json:"id"`
	Documentation []string `json:"documentation"`
	Evidence      []string `json:"evidence"`
}

type Classifier struct {
	ID     string   `json:"id"`
	Tokens []string `json:"tokens"`
}

type ImpactRules struct {
	Version             string              `json:"version,omitempty"`
	DocumentationGlobs  []string            `json:"documentation_globs"`
	IgnoredGlobs        []string            `json:"ignored_globs"`
	Classifiers         []Classifier        `json:"classifiers"`
	RequiredDocMappings map[string][]string `json:"required_doc_mappings"`
}

type ImpactAcknowledgment struct {
	Schema            string `json:"$schema"`
	SchemaVersion     string `json:"schema_version"`
	ChangeFingerprint string `json:"change_fingerprint"`
	Impact            string `json:"impact"`
	Reason            string `json:"reason"`
}

type ImpactFinding struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ImpactReport struct {
	SchemaVersion         string              `json:"schema_version"`
	Base                  string              `json:"base,omitempty"`
	Head                  string              `json:"head,omitempty"`
	MaterialChanges       []Change            `json:"material_changes"`
	DocumentationChanges  []Change            `json:"documentation_changes"`
	IgnoredChanges        []Change            `json:"ignored_changes"`
	Classifications       map[string][]string `json:"classifications"`
	ChangeFingerprint     string              `json:"change_fingerprint"`
	DocumentationImpact   string              `json:"documentation_impact"`
	Enforcement           string              `json:"enforcement,omitempty"`
	AcknowledgmentMatches bool                `json:"acknowledgment_matches,omitempty"`
	Findings              []ImpactFinding     `json:"findings,omitempty"`
	Status                string              `json:"status,omitempty"`
}

type Check struct {
	ID     string `json:"id"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

type DoctorReport struct {
	Status string  `json:"status"`
	Checks []Check `json:"checks"`
}

type PreflightActivationResult struct {
	Status string   `json:"status"`
	Root   string   `json:"root"`
	Routes []string `json:"routes"`
}

type PreflightGateResult struct {
	Agent    string `json:"agent"`
	Root     string `json:"root"`
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
	Output   string `json:"-"`
}

type AuditFinding struct {
	Severity string `json:"severity"`
	ID       string `json:"id"`
	Message  string `json:"message"`
}

type AuditReport struct {
	Status   string         `json:"status"`
	Findings []AuditFinding `json:"findings"`
	Impact   *ImpactReport  `json:"impact,omitempty"`
}
