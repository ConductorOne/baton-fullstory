package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/conductorone/baton-fullstory/pkg/fullstory"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	client       *fullstory.Client
	resourceType *v2.ResourceType
}

func userResource(user *fullstory.SCIMUser) (*v2.Resource, error) {
	displayName := user.GetDisplayName()
	email := user.GetEmail()

	profile := map[string]interface{}{
		"scim_id":      user.ID,
		"user_name":    user.UserName,
		"display_name": displayName,
		"first_name":   user.Name.GivenName,
		"last_name":    user.Name.FamilyName,
	}

	var status v2.UserTrait_Status_Status
	if user.Active {
		status = v2.UserTrait_Status_STATUS_ENABLED
	} else {
		status = v2.UserTrait_Status_STATUS_DISABLED
	}

	res, err := rs.NewUserResource(
		displayName,
		userResourceType,
		user.ID,
		[]rs.UserTraitOption{
			rs.WithEmail(email, true),
			rs.WithUserProfile(profile),
			rs.WithStatus(status),
		},
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (u *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func (u *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, token, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: userResourceType.Id})
	if err != nil {
		return nil, "", nil, fmt.Errorf("fullstory-connector: error parsing page token: %w", err)
	}

	// Convert string page token to SCIM startIndex
	startIndex := 1
	if token != "" {
		if idx, err := strconv.Atoi(token); err == nil {
			startIndex = idx
		}
	}

	pgVars := fullstory.NewPaginationVars(startIndex)
	users, nextStart, err := u.client.ListUsers(ctx, pgVars)
	if err != nil {
		return nil, "", nil, fmt.Errorf("fullstory-connector: error listing users: %w", err)
	}

	var rv []*v2.Resource
	for _, user := range users {
		ur, err := userResource(&user)
		if err != nil {
			return nil, "", nil, fmt.Errorf("fullstory-connector: error creating user resource: %w", err)
		}
		rv = append(rv, ur)
	}

	// Convert next startIndex back to string page token
	nextPage := ""
	if nextStart > 0 {
		nextPage = strconv.Itoa(nextStart)
	}

	nextToken, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, fmt.Errorf("fullstory-connector: error creating next page token: %w", err)
	}

	return rv, nextToken, nil, nil
}

func (u *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (u *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *fullstory.Client) *userBuilder {
	return &userBuilder{
		client:       client,
		resourceType: userResourceType,
	}
}
