package shortcut

// Story is the subset of Shortcut story fields we care about today, plus the
// raw JSON for forward compatibility.
type Story struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	EpicID          *int   `json:"epic_id"`
	AppURL          string `json:"app_url"`
	WorkflowStateID int    `json:"workflow_state_id"`
}

// Epic is the subset of Shortcut epic fields we care about today.
type Epic struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MilestoneID *int   `json:"milestone_id"`
	AppURL      string `json:"app_url"`
	EpicStateID int    `json:"epic_state_id"`
}

// WorkflowState is one state within a Workflow. Stories reference these by ID
// via Story.WorkflowStateID.
type WorkflowState struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Workflow is a Shortcut workflow (collection of states).
type Workflow struct {
	ID     int             `json:"id"`
	Name   string          `json:"name"`
	States []WorkflowState `json:"states"`
}

// EpicState is one epic-workflow state. Epics reference these by ID via
// Epic.EpicStateID.
type EpicState struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// EpicWorkflow is the workspace-wide epic workflow (Shortcut has only one).
type EpicWorkflow struct {
	EpicStates []EpicState `json:"epic_states"`
}

// Objective mirrors Shortcut's milestone/objective resource.
type Objective struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	AppURL string `json:"app_url"`
}
