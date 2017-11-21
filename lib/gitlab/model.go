package gitlab

// MergeRequestWebhook is the data structure that GitLab provides us
// in the merge request webhook. See:
// https://docs.gitlab.com/ce/user/project/integrations/webhooks.html#merge-request-events
type MergeRequestWebhook struct {
	ObjectKind       string `json:"object_kind"`
	ObjectAttributes struct {
		ID              int64  `json:"id"`
		IID             int64  `json:"iid"`
		SourceBranch    string `json:"source_branch"`
		TargetProjectID int64  `json:"target_project_id"`
		Action          string `json:"action"`
	} `json:"object_attributes"`
}

// MergeRequestID represents the id of single merge request,
// which is a combination of the id of a project and the
// project specific merge request iid.
type MergeRequestID struct {
	// The id of the project that the merge request targets.
	ProjectID int64
	// The project-specific merge request id. NOTE
	// that this is specifically the iid field of a
	// merge request and _not_ the id field.
	IID int64
}

// NewMergeRequestID extracts the MergeRequestID from the
// given MergeRequestWebhook representing the merge request
// id that the webhook was dispatched for.
func NewMergeRequestID(webhook *MergeRequestWebhook) MergeRequestID {
	return MergeRequestID{
		ProjectID: webhook.ObjectAttributes.TargetProjectID,
		IID:       webhook.ObjectAttributes.IID,
	}
}

// A Note is a comment on GitLab snippets, issues or merge requests.
// https://docs.gitlab.com/ee/api/notes.html
type Note struct {
	// Body is the markdown text content of the note.
	Body string `json:"body"`
}
