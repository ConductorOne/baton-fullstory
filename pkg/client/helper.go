package client

func (u *SCIMUser) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}

	if u.Name.GivenName != "" && u.Name.FamilyName != "" {
		return u.Name.GivenName + " " + u.Name.FamilyName
	}

	return u.UserName
}

// GetEmail returns the primary email, if no emails are found, returns the username.
func (u *SCIMUser) GetEmail() *SCIMEmail {
	for _, e := range u.Emails {
		if e.Primary {
			return e
		}
	}

	if len(u.Emails) > 0 {
		return u.Emails[0]
	}

	return &SCIMEmail{
		Value:   u.UserName,
		Primary: false,
	}
}
