package client

import "testing"

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		user     SCIMUser
		expected string
	}{
		{
			name: "returns DisplayName when set",
			user: SCIMUser{
				DisplayName: "Alice Smith",
				Name:        SCIMName{GivenName: "Alice", FamilyName: "Smith"},
				UserName:    "alice",
			},
			expected: "Alice Smith",
		},
		{
			name: "returns GivenName + FamilyName when both set and DisplayName empty",
			user: SCIMUser{
				Name:     SCIMName{GivenName: "Alice", FamilyName: "Smith"},
				UserName: "alice",
			},
			expected: "Alice Smith",
		},
		{
			name: "falls back to UserName when only GivenName set",
			user: SCIMUser{
				Name:     SCIMName{GivenName: "Alice"},
				UserName: "alice",
			},
			expected: "alice",
		},
		{
			name: "falls back to UserName when only FamilyName set",
			user: SCIMUser{
				Name:     SCIMName{FamilyName: "Smith"},
				UserName: "alice",
			},
			expected: "alice",
		},
		{
			name: "falls back to UserName when all name fields empty",
			user: SCIMUser{
				UserName: "alice",
			},
			expected: "alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.GetDisplayName()
			if got != tt.expected {
				t.Errorf("GetDisplayName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetEmail(t *testing.T) {
	tests := []struct {
		name          string
		user          SCIMUser
		expectedValue string
		expectedPrim  bool
	}{
		{
			name: "returns primary email",
			user: SCIMUser{
				Emails: []*SCIMEmail{
					{Value: "work@example.com", Primary: false},
					{Value: "primary@example.com", Primary: true},
				},
				UserName: "alice",
			},
			expectedValue: "primary@example.com",
			expectedPrim:  true,
		},
		{
			name: "returns first email when no primary is marked",
			user: SCIMUser{
				Emails: []*SCIMEmail{
					{Value: "first@example.com", Primary: false},
					{Value: "second@example.com", Primary: false},
				},
				UserName: "alice",
			},
			expectedValue: "first@example.com",
			expectedPrim:  false,
		},
		{
			name: "falls back to UserName when emails slice is empty",
			user: SCIMUser{
				UserName: "alice",
			},
			expectedValue: "alice",
			expectedPrim:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.GetEmail()
			if got.Value != tt.expectedValue {
				t.Errorf("GetEmail().Value = %q, want %q", got.Value, tt.expectedValue)
			}
			if got.Primary != tt.expectedPrim {
				t.Errorf("GetEmail().Primary = %v, want %v", got.Primary, tt.expectedPrim)
			}
		})
	}
}
