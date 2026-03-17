package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/conductorone/baton-fullstory/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	client       *client.Client
	resourceType *v2.ResourceType
}

func (b *userBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return b.resourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (b *userBuilder) List(ctx context.Context, _ *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	bag, token, err := parsePageToken(opts.PageToken.Token, &v2.ResourceId{ResourceType: userResourceType.Id})
	if err != nil {
		return nil, nil, fmt.Errorf("baton-fullstory: error parsing page token: %w", err)
	}

	startIndex := 1
	if token != "" {
		idx, err := strconv.Atoi(token)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-fullstory: error parsing page token: %w", err)
		}

		startIndex = idx
	}

	pgVars := client.NewPaginationVars(startIndex)
	users, nextStart, err := b.client.ListUsers(ctx, pgVars)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-fullstory: error listing users: %w", err)
	}

	var rv []*v2.Resource
	for _, user := range users {
		userResource, err := b.createResource(user)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-fullstory: error creating user resource: %w", err)
		}

		rv = append(rv, userResource)
	}

	// Convert next startIndex back to string page token
	nextPage := ""
	if nextStart > 0 {
		nextPage = strconv.Itoa(nextStart)
	}

	nextToken, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-fullstory: error creating next page token: %w", err)
	}

	return rv, &rs.SyncOpResults{
		NextPageToken: nextToken,
	}, nil
}

func (b *userBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

func (b *userBuilder) Grants(_ context.Context, _ *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

func (b *userBuilder) createResource(user *client.SCIMUser) (*v2.Resource, error) {
	displayName := user.GetDisplayName()
	email := user.GetEmail()

	profile := map[string]interface{}{
		"scim_id":      user.ID,
		"user_name":    user.UserName,
		"display_name": displayName,
	}

	if user.Name.GivenName != "" {
		profile["first_name"] = user.Name.GivenName
	}

	if user.Name.FamilyName != "" {
		profile["last_name"] = user.Name.FamilyName
	}

	var status v2.UserTrait_Status_Status
	if user.Active {
		status = v2.UserTrait_Status_STATUS_ENABLED
	} else {
		status = v2.UserTrait_Status_STATUS_DISABLED
	}

	res, err := rs.NewUserResource(
		displayName,
		b.resourceType,
		user.ID,
		[]rs.UserTraitOption{
			rs.WithEmail(email.Value, email.Primary),
			rs.WithUserLogin(user.UserName),
			rs.WithUserProfile(profile),
			rs.WithStatus(status),
		},
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func newUserBuilder(client *client.Client) *userBuilder {
	return &userBuilder{
		client:       client,
		resourceType: userResourceType,
	}
}
