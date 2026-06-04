package episode

const (
	RuntimeSchemaVersion = "sheet-ops-runtime/v1"
	RuntimeName          = "sheet-ops"
)

var RuntimeCommit = ""

type RuntimeVersion struct {
	SchemaVersion string `json:"schema_version"`
	RuntimeName   string `json:"runtime_name"`
	RuntimeCommit string `json:"runtime_commit,omitempty"`
}

func CurrentRuntimeVersion() RuntimeVersion {
	return RuntimeVersion{
		SchemaVersion: RuntimeSchemaVersion,
		RuntimeName:   RuntimeName,
		RuntimeCommit: RuntimeCommit,
	}
}
