package connector

import (
	"context"
	"fmt"

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

func userResource(user *fullstory.User) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"user_id": user.ID,
	}

	var status v2.UserTrait_Status_Status
	if user.IsBeingDeleted {
		status = v2.UserTrait_Status_STATUS_DISABLED
	} else {
		status = v2.UserTrait_Status_STATUS_ENABLED
	}

	res, err := rs.NewUserResource(
		user.Name,
		userResourceType,
		user.ID,
		[]rs.UserTraitOption{
			rs.WithEmail(user.Email, true),
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

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (u *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, token, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: userResourceType.Id})
	if err != nil {
		return nil, "", nil, fmt.Errorf("fullstory-connector: error parsing page token: %w", err)
	}

	pgVars := fullstory.NewPaginationVars(token)
	users, nextPage, err := u.client.ListUsers(ctx, pgVars)
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
