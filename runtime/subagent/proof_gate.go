package subagent

import "fmt"

type SubagentLoopState struct {
	ScenarioID        string                `json:"scenario_id"`
	Outcome           string                `json:"outcome"`
	ParentState       string                `json:"parent_state"`
	PublicEntryStatus string                `json:"public_entry_status,omitempty"`
	WorkUnitID        string                `json:"work_unit_id,omitempty"`
	Specialists       []SpecialistLoopState `json:"specialists"`
}

type SpecialistLoopState struct {
	Role                string `json:"role"`
	Carrier             string `json:"carrier"`
	Model               string `json:"model"`
	ReasoningEffort     string `json:"reasoning_effort"`
	State               string `json:"state"`
	SessionID           string `json:"session_id"`
	SpawnCompletedCount int    `json:"spawn_completed_count"`
	WaitCompletedCount  int    `json:"wait_completed_count"`
	CloseCompletedCount int    `json:"close_completed_count"`
	ExitCode            *int   `json:"exit_code,omitempty"`
	LastMessage         string `json:"last_message"`
}

func RequireSpecialistLoopState(state SubagentLoopState) error {
	if len(state.Specialists) == 0 {
		return fmt.Errorf("specialist lifecycle proof is missing")
	}

	for _, specialist := range state.Specialists {
		if specialist.Role == "" {
			return fmt.Errorf("specialist role is missing")
		}
		if specialist.Model == "" {
			return fmt.Errorf("%s missing model proof", specialist.Role)
		}
		if specialist.ReasoningEffort == "" {
			return fmt.Errorf("%s missing reasoning effort proof", specialist.Role)
		}
		if specialist.SessionID == "" {
			return fmt.Errorf("%s missing session proof", specialist.Role)
		}
		if specialist.SpawnCompletedCount < 1 {
			return fmt.Errorf("%s missing spawn proof", specialist.Role)
		}
		if specialist.WaitCompletedCount < 1 {
			return fmt.Errorf("%s missing wait proof", specialist.Role)
		}
		if specialist.CloseCompletedCount < 1 {
			return fmt.Errorf("%s missing close proof", specialist.Role)
		}
	}

	return nil
}
