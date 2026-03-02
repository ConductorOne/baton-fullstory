package fullstory

// SCIMListResponse is the top-level SCIM list response.
type SCIMListResponse struct {
	TotalResults int        `json:"totalResults"`
	StartIndex   int        `json:"startIndex"`
	ItemsPerPage int        `json:"itemsPerPage"`
	Resources    []SCIMUser `json:"Resources"`
}

// SCIMUser represents a FullStory teammate via the SCIM API.
type SCIMUser struct {
	ID          string     `json:"id"`
	UserName    string     `json:"userName"`
	DisplayName string     `json:"displayName"`
	Name        SCIMName   `json:"name"`
	Emails      []SCIMAttr `json:"emails"`
	Active      bool       `json:"active"`
}

type SCIMName struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

type SCIMAttr struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}

// GetDisplayName returns the best available display name.
func (u *SCIMUser) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.Name.GivenName != "" || u.Name.FamilyName != "" {
		return u.Name.GivenName + " " + u.Name.FamilyName
	}
	return u.UserName
}

// GetEmail returns the primary email, falling back to userName.
func (u *SCIMUser) GetEmail() string {
	for _, e := range u.Emails {
		if e.Primary {
			return e.Value
		}
	}
	if len(u.Emails) > 0 {
		return u.Emails[0].Value
	}
	return u.UserName
}
