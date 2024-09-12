package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/tursodatabase/libsql-client-go/libsql"
)

func TestAccResourceGroupToken_NoExpiration(t *testing.T) {
	name := randomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateConfig(`
				resource "turso_database" "test" {
					group = "test"
					name = "` + name + `"
				}
				data "turso_group_token" "test" {
					id = "test"
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectSensitiveValue("data.turso_group_token.test", tfjsonpath.New("jwt")),
					statecheck.ExpectKnownValue("data.turso_group_token.test", tfjsonpath.New("jwt"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.turso_group_token.test", tfjsonpath.New("expiration"), knownvalue.Null()), // no expiration
				},
				Check: func(s *terraform.State) error {
					dbName, ok := s.RootModule().Resources["turso_database.test"].Primary.Attributes["name"]
					if !ok {
						return fmt.Errorf("missing database")
					}
					token, ok := s.RootModule().Resources["data.turso_group_token.test"].Primary.Attributes["jwt"]
					if !ok {
						return fmt.Errorf("missing token")
					}

					dbUri := fmt.Sprintf("libsql://%s-celest-dev.turso.io", dbName)
					connector, err := libsql.NewConnector(dbUri, libsql.WithAuthToken(token))
					if err != nil {
						return fmt.Errorf("error creating database connector: %v", err)
					}
					db := sql.OpenDB(connector)
					defer db.Close()

					// Test connection
					if err := db.Ping(); err != nil {
						return fmt.Errorf("error pinging database: %v", err)
					}
					if _, err := db.Query("select 1"); err != nil {
						return fmt.Errorf("error querying database: %v", err)
					}

					return nil
				},
			},
		},
	})
}

func TestAccResourceGroupToken_WithExpiration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateConfig(`
				data "turso_group_token" "test" {
					id = "test"
					expiration = "1h"
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectSensitiveValue("data.turso_group_token.test", tfjsonpath.New("jwt")),
					statecheck.ExpectKnownValue("data.turso_group_token.test", tfjsonpath.New("jwt"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.turso_group_token.test", tfjsonpath.New("expiration"), knownvalue.NotNull()), // no expiration
				},
			},
		},
	})
}

func TestAccResourceGroupToken_E2E(t *testing.T) {
	name := randomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateConfig(`
				resource "turso_database" "test" {
					group = "test"
					name = "` + name + `"
				}
				data "turso_group_token" "test" {
					id = "test"
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectSensitiveValue("data.turso_group_token.test", tfjsonpath.New("jwt")),
					statecheck.ExpectKnownValue("data.turso_group_token.test", tfjsonpath.New("jwt"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.turso_group_token.test", tfjsonpath.New("expiration"), knownvalue.Null()), // no expiration
				},
				Check: func(s *terraform.State) error {
					dbName, ok := s.RootModule().Resources["turso_database.test"].Primary.Attributes["name"]
					if !ok {
						return fmt.Errorf("missing database")
					}
					token, ok := s.RootModule().Resources["data.turso_group_token.test"].Primary.Attributes["jwt"]
					if !ok {
						return fmt.Errorf("missing token")
					}

					dbUri := fmt.Sprintf("libsql://%s-celest-dev.turso.io", dbName)
					connector, err := libsql.NewConnector(dbUri, libsql.WithAuthToken(token))
					if err != nil {
						return fmt.Errorf("error creating database connector: %v", err)
					}
					db := sql.OpenDB(connector)
					defer db.Close()

					// Test connection
					if err := db.Ping(); err != nil {
						return fmt.Errorf("error pinging database: %v", err)
					}
					if _, err := db.Query("select 1"); err != nil {
						return fmt.Errorf("error querying database: %v", err)
					}

					return nil
				},
			},
		},
	})
}
