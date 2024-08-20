package tursoadmin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoadmin/internal/httpmodel"
)

// Group represents a Turso group's metadata.
type Group struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	UUID            string   `json:"uuid"`
	Locations       []string `json:"locations"`
	PrimaryLocation string   `json:"primary"`
	Archived        bool     `json:"archived"`
}

type Extensions string

const (
	ExtensionsAll Extensions = "all"
)

type CreateGroupRequest struct {
	Name       string     `json:"name"`
	Location   string     `json:"location"`
	Extensions Extensions `json:"extensions,omitempty"`
}

// GetGroup retrieves a group by name.
//
// https://docs.turso.tech/api-reference/groups/retrieve
func (c *Client) GetGroup(ctx context.Context, groupName string) (Group, error) {
	type response struct {
		Group Group `json:"group"`
	}
	model := httpmodel.NewJSON[response](http.MethodGet, "/organizations/"+c.orgName+"/groups/"+groupName, struct{}{})
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return Group{}, fmt.Errorf("failed to get group: %v", err)
	}
	return resp.Group, nil
}

// CreateGroup creates a new group in the organization.
//
// https://docs.turso.tech/api-reference/groups/create
func (c *Client) CreateGroup(ctx context.Context, request CreateGroupRequest) (Group, error) {
	type response struct {
		Group Group `json:"group"`
	}
	model := httpmodel.NewJSON[response](http.MethodPost, "/organizations/"+c.orgName+"/groups", request)
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return Group{}, fmt.Errorf("failed to create group: %v", err)
	}
	return resp.Group, nil
}

// AddGroupLocation adds a location to a group.
//
// https://docs.turso.tech/api-reference/groups/add-location
func (c *Client) AddGroupLocation(ctx context.Context, groupName, location string) (Group, error) {
	type response struct {
		Group Group `json:"group"`
	}
	model := httpmodel.NewJSON[response](http.MethodPost, "/organizations/"+c.orgName+"/groups/"+groupName+"/locations/"+location, nil)
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return Group{}, fmt.Errorf("failed to add location to group: %v", err)
	}
	return resp.Group, nil
}

// RemoveGroupLocation removes a location from a group.
//
// https://docs.turso.tech/api-reference/groups/remove-location
func (c *Client) RemoveGroupLocation(ctx context.Context, groupName, location string) (Group, error) {
	type response struct {
		Group Group `json:"group"`
	}
	model := httpmodel.NewJSON[response](http.MethodDelete, "/organizations/"+c.orgName+"/groups/"+groupName+"/locations/"+location, nil)
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return Group{}, fmt.Errorf("failed to remove location from group: %v", err)
	}
	return resp.Group, nil
}

// CreateGroupToken generates an authorization token for the specified group.
//
// https://docs.turso.tech/api-reference/groups/create-token
func (c *Client) CreateGroupToken(ctx context.Context, groupName string, expiration time.Duration) (string, error) {
	type response struct {
		JWT string `json:"jwt"`
	}
	var expParam string
	if expiration == time.Duration(0) {
		expParam = "never"
	} else {
		expParam = fmt.Sprintf("%ds", int(expiration.Seconds()))
	}
	route := fmt.Sprintf("/organizations/%s/groups/%s/auth/tokens?expiration=%s&authorization=full-access", c.orgName, groupName, expParam)
	model := httpmodel.NewJSON[response](http.MethodPost, route, struct{}{})
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return "", fmt.Errorf("failed to create group token: %v", err)
	}
	return resp.JWT, nil
}

// InvalidateGroupTokens invalidates all authorization tokens for the specified group.
//
// https://docs.turso.tech/api-reference/groups/invalidate-tokens
func (c *Client) InvalidateGroupTokens(ctx context.Context, groupName string) error {
	route := fmt.Sprintf("/organizations/%s/groups/%s/auth/rotate", c.orgName, groupName)
	model := httpmodel.NewJSON[struct{}](http.MethodPost, route, nil)
	_, err := model.Send(ctx, c.http)
	if err != nil {
		return fmt.Errorf("failed to invalidate group token: %v", err)
	}
	return nil
}

// DeleteGroup deletes a group by name.
//
// https://docs.turso.tech/api-reference/groups/delete
func (c *Client) DeleteGroup(ctx context.Context, groupName string) error {
	model := httpmodel.NewJSON[struct{}](http.MethodDelete, "/organizations/"+c.orgName+"/groups/"+groupName, nil)
	_, err := model.Send(ctx, c.http)
	if err != nil {
		return fmt.Errorf("failed to delete group: %v", err)
	}
	return nil
}

// ListGroups retrieves all groups in the organization.
//
// https://docs.turso.tech/api-reference/groups/list
func (c *Client) ListGroups(ctx context.Context) ([]Group, error) {
	type response struct {
		Groups []Group `json:"groups"`
	}
	model := httpmodel.NewJSON[response]("GET", "/organizations/"+c.orgName+"/groups", struct{}{})
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %v", err)
	}
	return resp.Groups, nil
}
