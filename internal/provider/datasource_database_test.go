package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceDatabase(t *testing.T) {
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
				data "turso_database" "test" {
					id = turso_database.test.id
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.turso_database.test", tfjsonpath.New("id"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("data.turso_database.test", tfjsonpath.New("database"), knownvalue.NotNull()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.turso_database.test", "id", name),
				),
			},
		},
	})
}
