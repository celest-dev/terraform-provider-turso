package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceDatabaseInstances(t *testing.T) {
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
				data "turso_database_instances" "test" {
					id = turso_database.test.id
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.turso_database_instances.test", tfjsonpath.New("instances"), listNotEmpty{}),
					statecheck.ExpectKnownValue("data.turso_database_instances.test", tfjsonpath.New("instances"), listOfNonNulls{}),
				},
			},
		},
	})
}
