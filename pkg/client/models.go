package client

type SCIMListResponse struct {
	TotalResults int         `json:"totalResults"`
	StartIndex   int         `json:"startIndex"`
	ItemsPerPage int         `json:"itemsPerPage"`
	Resources    []*SCIMUser `json:"Resources"`
}

// SCIMUser represents a FullStory teammate via the SCIM API.
type SCIMUser struct {
	ID          string       `json:"id"`
	UserName    string       `json:"userName"`
	DisplayName string       `json:"displayName"`
	Name        SCIMName     `json:"name"`
	Emails      []*SCIMEmail `json:"emails"`
	Active      bool         `json:"active"`
}

type SCIMName struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}
