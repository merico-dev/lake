// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package generated

type PullRequests struct {
	// SdcRepository corresponds to the JSON schema field "_sdc_repository".
	SdcRepository *string `json:"_sdc_repository,omitempty"`

	// Base corresponds to the JSON schema field "base".
	Base *PullRequestsBase `json:"base,omitempty"`

	// Body corresponds to the JSON schema field "body".
	Body *string `json:"body,omitempty"`

	// ClosedAt corresponds to the JSON schema field "closed_at".
	ClosedAt *string `json:"closed_at,omitempty"`

	// CreatedAt corresponds to the JSON schema field "created_at".
	CreatedAt *string `json:"created_at,omitempty"`

	// Id corresponds to the JSON schema field "id".
	Id *string `json:"id,omitempty"`

	// Labels corresponds to the JSON schema field "labels".
	Labels []PullRequestsLabelsElem `json:"labels,omitempty"`

	// MergedAt corresponds to the JSON schema field "merged_at".
	MergedAt *string `json:"merged_at,omitempty"`

	// Number corresponds to the JSON schema field "number".
	Number *int `json:"number,omitempty"`

	// State corresponds to the JSON schema field "state".
	State *string `json:"state,omitempty"`

	// Title corresponds to the JSON schema field "title".
	Title *string `json:"title,omitempty"`

	// UpdatedAt corresponds to the JSON schema field "updated_at".
	UpdatedAt *string `json:"updated_at,omitempty"`

	// Url corresponds to the JSON schema field "url".
	Url *string `json:"url,omitempty"`

	// User corresponds to the JSON schema field "user".
	User *PullRequestsUser `json:"user,omitempty"`
}

type PullRequestsBase struct {
	// Label corresponds to the JSON schema field "label".
	Label *string `json:"label,omitempty"`

	// Ref corresponds to the JSON schema field "ref".
	Ref *string `json:"ref,omitempty"`

	// Repo corresponds to the JSON schema field "repo".
	Repo *PullRequestsBaseRepo `json:"repo,omitempty"`

	// Sha corresponds to the JSON schema field "sha".
	Sha *string `json:"sha,omitempty"`
}

type PullRequestsBaseRepo struct {
	// Id corresponds to the JSON schema field "id".
	Id *int `json:"id,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name *string `json:"name,omitempty"`

	// Url corresponds to the JSON schema field "url".
	Url *string `json:"url,omitempty"`
}

type PullRequestsLabelsElem struct {
	// Color corresponds to the JSON schema field "color".
	Color *string `json:"color,omitempty"`

	// Default corresponds to the JSON schema field "default".
	Default *bool `json:"default,omitempty"`

	// Description corresponds to the JSON schema field "description".
	Description *string `json:"description,omitempty"`

	// Id corresponds to the JSON schema field "id".
	Id *int `json:"id,omitempty"`

	// Name corresponds to the JSON schema field "name".
	Name *string `json:"name,omitempty"`

	// NodeId corresponds to the JSON schema field "node_id".
	NodeId *string `json:"node_id,omitempty"`

	// Url corresponds to the JSON schema field "url".
	Url *string `json:"url,omitempty"`
}

type PullRequestsUser struct {
	// Id corresponds to the JSON schema field "id".
	Id *int `json:"id,omitempty"`

	// Login corresponds to the JSON schema field "login".
	Login *string `json:"login,omitempty"`
}
