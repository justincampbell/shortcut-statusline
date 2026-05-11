package shortcut

// Story is the subset of Shortcut story fields we care about today, plus the
// raw JSON for forward compatibility.
type Story struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	EpicID          *int     `json:"epic_id"`
	AppURL          string   `json:"app_url"`
	WorkflowStateID int      `json:"workflow_state_id"`
	Type            string   `json:"story_type"`
	OwnerIDs        []string `json:"owner_ids"`
	RequestedByID   string   `json:"requested_by_id"`
	GroupID         *string  `json:"group_id"`
}

// Epic is the subset of Shortcut epic fields we care about today.
type Epic struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	MilestoneID *int     `json:"milestone_id"`
	AppURL      string   `json:"app_url"`
	EpicStateID int      `json:"epic_state_id"`
	OwnerIDs    []string `json:"owner_ids"`
	GroupID     *string  `json:"group_id"`
}

// WorkflowState is one state within a Workflow. Stories reference these by ID
// via Story.WorkflowStateID. Type is one of "backlog", "unstarted",
// "started", "done" and drives the semantic color.
type WorkflowState struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Workflow is a Shortcut workflow (collection of states).
type Workflow struct {
	ID     int             `json:"id"`
	Name   string          `json:"name"`
	States []WorkflowState `json:"states"`
}

// EpicState is one epic-workflow state. Epics reference these by ID via
// Epic.EpicStateID. Type matches WorkflowState.Type.
type EpicState struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// EpicWorkflow is the workspace-wide epic workflow (Shortcut has only one).
type EpicWorkflow struct {
	EpicStates []EpicState `json:"epic_states"`
}

// Objective mirrors Shortcut's milestone/objective resource. State is a
// direct string from the API (e.g. "to do", "in progress", "done"), so no
// extra workflow lookup is required.
type Objective struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	AppURL string `json:"app_url"`
	State  string `json:"state"`
}

// Member is a workspace user. The handful of fields we keep is enough to
// resolve owner / requestor IDs into a short statusline-friendly label.
type Member struct {
	ID       string        `json:"id"`
	Profile  MemberProfile `json:"profile"`
	Disabled bool          `json:"disabled"`
}

// MemberProfile is the nested profile object on Member.
type MemberProfile struct {
	MentionName  string `json:"mention_name"`
	Name         string `json:"name"`
	EmailAddress string `json:"email_address"`
}

// Group (Shortcut "Team") is a collection of members. Stories and epics
// reference it via group_id.
type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MentionName string `json:"mention_name"`
	Archived    bool   `json:"archived"`
}
