package requestcompiler

type RequestSourceKind string

const (
	RequestSourcePromptFile RequestSourceKind = "prompt_file"
	RequestSourceDirectText RequestSourceKind = "direct_text"
)

type RequestSource struct {
	Kind RequestSourceKind `json:"kind"`
	Path string            `json:"path,omitempty"`
	Text string            `json:"text,omitempty"`
}

type WorkbookInput struct {
	Path string `json:"path"`
	Role string `json:"role"`
}

type Input struct {
	RequestSource  RequestSource   `json:"request_source"`
	WorkspaceRoot  string          `json:"workspace_root"`
	WorkingDir     string          `json:"working_dir"`
	ScenarioSlug   string          `json:"scenario_slug"`
	InputWorkbooks []WorkbookInput `json:"input_workbooks"`
	OutputFile     string          `json:"output_file"`
}
