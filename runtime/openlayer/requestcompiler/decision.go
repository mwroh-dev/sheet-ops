package requestcompiler

type Status string

const (
	StatusCompiled             Status = "compiled"
	StatusNeedsHumanCheckpoint Status = "needs_human_checkpoint"
	StatusBlocked              Status = "blocked"
)

type Checkpoint struct {
	Required bool
	Kind     string
	Question string
	Options  []string
}

type Decision struct {
	Status            Status
	SelectedOperation string
	Confidence        float64
	Notes             []string
	Checkpoint        Checkpoint
	SheetCandidates   []string
	StructuralSignals []string
}
