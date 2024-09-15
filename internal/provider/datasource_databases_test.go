package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceDatabases(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateConfig(`
				data "turso_databases" "test" {}
				data "turso_database" "test" {
					id = data.turso_databases.test.databases[0].name
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.turso_databases.test", tfjsonpath.New("databases"), listNotEmpty{}),
					statecheck.ExpectKnownValue("data.turso_databases.test", tfjsonpath.New("databases"), listOfNonNulls{}),
					statecheck.ExpectKnownValue("data.turso_database.test", tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.turso_database.test", tfjsonpath.New("database"), knownvalue.NotNull()),
				},
			},
		},
	})
}
