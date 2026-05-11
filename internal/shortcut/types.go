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
}

// Objective mirrors Shortcut's milestone/objective resource.
type Objective struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	AppURL string `json:"app_url"`
}
