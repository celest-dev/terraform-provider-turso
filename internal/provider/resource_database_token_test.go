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

func TestAccResourceDatabaseToken_E2E(t *testing.T) {
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
				resource "turso_database_token" "test" {
					database = turso_database.test.name
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("turso_database.test", tfjsonpath.New("instances"), knownvalue.MapSizeExact(1)),
					statecheck.ExpectKnownValue("turso_database.test", tfjsonpath.New("hostname"), knownvalue.NotNull()),

					statecheck.ExpectSensitiveValue("turso_database_token.test", tfjsonpath.New("token")),
					statecheck.ExpectKnownValue("turso_database_token.test", tfjsonpath.New("token"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("turso_database_token.test", tfjsonpath.New("expires_at"), knownvalue.Null()), // no expiration
				},
				Check: func(s *terraform.State) error {
					dbHostname, ok := s.RootModule().Resources["turso_database.test"].Primary.Attributes["hostname"]
					if !ok {
						return fmt.Errorf("missing hostname")
					}
					token, ok := s.RootModule().Resources["turso_database_token.test"].Primary.Attributes["token"]
					if !ok {
						return fmt.Errorf("missing token")
					}

					connector, err := libsql.NewConnector("libsql://"+dbHostname, libsql.WithAuthToken(token))
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
