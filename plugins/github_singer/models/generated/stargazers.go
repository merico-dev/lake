// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package generated

type Stargazers struct {
	// SdcRepository corresponds to the JSON schema field "_sdc_repository".
	SdcRepository *string `json:"_sdc_repository,omitempty"`

	// StarredAt corresponds to the JSON schema field "starred_at".
	StarredAt *string `json:"starred_at,omitempty"`

	// User corresponds to the JSON schema field "user".
	User *StargazersUser `json:"user,omitempty"`

	// UserId corresponds to the JSON schema field "user_id".
	UserId *int `json:"user_id,omitempty"`
}

type StargazersUser struct {
	// Id corresponds to the JSON schema field "id".
	Id *int `json:"id,omitempty"`
}
