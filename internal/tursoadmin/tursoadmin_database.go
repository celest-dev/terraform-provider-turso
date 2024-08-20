package tursoadmin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoadmin/internal/httpmodel"
)

// CreateDatabaseRequest is the configuration for creating a new Turso database.
type CreateDatabaseRequest struct {
	Name string `json:"name"`

	// The name of the group where the database should be created. The group must already exist.
	Group string `json:"group,omitempty"`

	Seed *DatabaseSeed `json:"seed,omitempty"`

	// The maximum size of the database in bytes. Values with units are also accepted, e.g. 1mb, 256mb, 1gb.
	SizeLimit string `json:"size_limit,omitempty"`

	// Mark this database as the parent schema database that updates child databases with any schema changes.
	//
	// See [Multi-DB Schemas](https://docs.turso.tech/features/multi-db-schemas).
	IsSchema bool `json:"is_schema,omitempty"`

	// The name of the parent database to use as the schema.
	//
	// See [Multi-DB Schemas](https://docs.turso.tech/features/multi-db-schemas).
	SchemaDatabase string `json:"schema,omitempty"`
}

type DatabaseSeedType string

const (
	DatabaseSeedDatabase DatabaseSeedType = "database"
	DatabaseSeedDump     DatabaseSeedType = "dump"
)

type DatabaseSeed struct {
	// The type of seed to be used to create a new database.
	Type DatabaseSeedType `json:"type"`

	// The name of the existing database when `database` is used as a seed type.
	Name string `json:"name,omitempty"`

	// The URL returned by [upload dump](https://docs.turso.tech/api-reference/databases/upload-dump)
	// can be used with the `dump` seed type.
	URL string `json:"url,omitempty"`

	// A formatted ISO 8601 recovery point to create a database from. This must be within the
	// last 24 hours, or 30 days on the scaler plan.
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// Database represents a Turso database's metadata.
type Database struct {
	// The database universal unique identifier (UUID).
	ID string `json:"DbId"`

	// The DNS hostname used for client libSQL and HTTP connections.
	Hostname string `json:"Hostname"`

	// The database name, unique across your organization.
	Name string `json:"Name"`
}

type DatabaseConfig struct {
	// The maximum size of the database in bytes. Values with units are also accepted, e.g. 1mb, 256mb, 1gb.
	SizeLimit *string `json:"size_limit,omitempty"`

	// Allow or disallow attaching databases to the current database.
	AllowAttach *bool `json:"allow_attach,omitempty"`

	// Block all database reads.
	BlockReads *bool `json:"block_reads,omitempty"`

	// Block all database writes.
	BlockWrites *bool `json:"block_writes,omitempty"`
}

// A Database with its configuration, returned from GetDatabase.
type DatabaseWithConfig struct {
	// The database universal unique identifier (UUID).
	ID string `json:"DbId"`

	// The DNS hostname used for client libSQL and HTTP connections.
	Hostname string `json:"Hostname"`

	// The database name, unique across your organization.
	Name string `json:"Name"`

	// The name of the group the database belongs to.
	Group string `json:"group"`

	// Allow or disallow attaching databases to the current database.
	AllowAttach bool `json:"allow_attach"`

	// Block all database reads.
	BlockReads bool `json:"block_reads"`

	// Block all database writes.
	BlockWrites bool `json:"block_writes"`

	// A list of regions for the group the database belongs to.
	Regions []string `json:"regions"`

	// The primary region location code the group the database belongs to.
	PrimaryRegion string `json:"primaryRegion"`

	// The string representing the object type.
	Type string `json:"type"`

	// The current libSQL version the database is running.
	Version string `json:"version"`

	// If this database controls other child databases then this will be true.
	//
	// See [Multi-DB Schemas](https://docs.turso.tech/features/multi-db-schemas).
	IsSchema bool `json:"is_schema"`

	// The name of the parent database that owns the schema for this database.
	//
	// See [Multi-DB Schemas](https://docs.turso.tech/features/multi-db-schemas).
	Schema string `json:"schema,omitempty"`
}

type DatabaseInstanceType string

const (
	DatabaseInstanceTypePrimary DatabaseInstanceType = "primary"
	DatabaseInstanceTypeReplica DatabaseInstanceType = "replica"
)

// An instance of a Database.
type DatabaseInstance struct {
	// The instance universal unique identifier (UUID).
	UUID string `json:"uuid"`

	// The name of the instance (location code).
	Name string `json:"name"`

	// The type of database instance. One of: `primary` or `replica`.
	Type DatabaseInstanceType `json:"type"`

	// The location code for the region this instance is located.
	Region string `json:"region"`

	// The DNS hostname used for client libSQL and HTTP connections (specific to this instance only).
	Hostname string `json:"hostname"`
}

// CreateDatabase creates a new database in a group for the organization or user.
//
// https://docs.turso.tech/api-reference/databases/create
func (c *Client) CreateDatabase(ctx context.Context, request CreateDatabaseRequest) (Database, error) {
	type response struct {
		Database Database `json:"database"`
	}
	model := httpmodel.NewJSON[response](http.MethodPost, "/organizations/"+c.orgName+"/databases", request)
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return Database{}, fmt.Errorf("failed to create database: %v", err)
	}
	return resp.Database, nil
}

// ListDatabaseInstances returns a list of instances of a database. Instances are the individual primary or replica databases in each region defined by the group.
//
// https://docs.turso.tech/api-reference/databases/list-instances
func (c *Client) ListDatabaseInstances(ctx context.Context, dbName string) ([]DatabaseInstance, error) {
	type response struct {
		Instances []DatabaseInstance `json:"instances"`
	}
	model := httpmodel.NewJSON[response](http.MethodGet, "/organizations/"+c.orgName+"/databases/"+dbName+"/instances", nil)
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return nil, fmt.Errorf("failed to read database instances: %v", err)
	}
	return resp.Instances, nil
}

// GetDatabase retrieves a database by name.
//
// https://docs.turso.tech/api-reference/databases/retrieve
func (c *Client) GetDatabase(ctx context.Context, dbName string) (DatabaseWithConfig, error) {
	type response struct {
		Database DatabaseWithConfig `json:"database"`
	}
	model := httpmodel.NewJSON[response](http.MethodGet, "/organizations/"+c.orgName+"/databases/"+dbName, struct{}{})
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return DatabaseWithConfig{}, fmt.Errorf("failed to get database: %v", err)
	}
	return resp.Database, nil
}

// GetDatabaseConfiguration gets a database's configuration.
//
// https://docs.turso.tech/api-reference/databases/configuration
func (c *Client) GetDatabaseConfiguration(ctx context.Context, dbName string) (DatabaseConfig, error) {
	model := httpmodel.NewJSON[DatabaseConfig](http.MethodGet, "/organizations/"+c.orgName+"/databases/"+dbName+"/configuration", nil)
	config, err := model.Send(ctx, c.http)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to get database config: %v", err)
	}
	return config, nil
}

// UpdateDatabaseConfiguration updates a database's configuration.
//
// https://docs.turso.tech/api-reference/databases/update-configuration
func (c *Client) UpdateDatabaseConfiguration(ctx context.Context, dbName string, config DatabaseConfig) (DatabaseConfig, error) {
	model := httpmodel.NewJSON[DatabaseConfig](http.MethodPatch, "/organizations/"+c.orgName+"/databases/"+dbName+"/configuration", config)
	updatedConfig, err := model.Send(ctx, c.http)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to update database config: %v", err)
	}
	return updatedConfig, nil
}

// DeleteDatabase deletes a database by name.
//
// https://docs.turso.tech/api-reference/databases/delete
func (c *Client) DeleteDatabase(ctx context.Context, dbName string) error {
	model := httpmodel.NewJSON[struct{}](http.MethodDelete, "/organizations/"+c.orgName+"/databases/"+dbName, nil)
	_, err := model.Send(ctx, c.http)
	if err != nil {
		return fmt.Errorf("failed to delete database: %v", err)
	}
	return nil
}

// CreateDatabaseToken generates an authorization token for the specified database.
//
// https://docs.turso.tech/api-reference/databases/create-token
func (c *Client) CreateDatabaseToken(ctx context.Context, dbName string, expiration time.Duration) (string, error) {
	type response struct {
		JWT string `json:"jwt"` // jwt is the JSON Web Token.
	}
	var expParam string
	if expiration == time.Duration(0) {
		expParam = "never"
	} else {
		expParam = fmt.Sprintf("%ds", int(expiration.Seconds()))
	}
	route := fmt.Sprintf("/organizations/%s/databases/%s/auth/tokens?expiration=%s&authorization=full-access", c.orgName, dbName, expParam)
	model := httpmodel.NewJSON[response](http.MethodPost, route, struct{}{})
	resp, err := model.Send(ctx, c.http)
	if err != nil {
		return "", fmt.Errorf("failed to create database token: %v", err)
	}
	return resp.JWT, nil
}

// InvalidateDatabaseTokens invalidates all authorization tokens for the specified database.
//
// https://docs.turso.tech/api-reference/databases/invalidate-tokens
func (c *Client) InvalidateDatabaseTokens(ctx context.Context, dbName string) error {
	route := fmt.Sprintf("/organizations/%s/databases/%s/auth/rotate", c.orgName, dbName)
	fmt.Printf("Invalidating all database tokens for: %s", route)
	model := httpmodel.NewJSON[struct{}](http.MethodPost, route, nil)
	_, err := model.Send(ctx, c.http)
	if err != nil {
		return fmt.Errorf("failed to invalidate database token: %v", err)
	}
	return nil
}
