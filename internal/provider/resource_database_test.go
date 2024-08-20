package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceDatabase(t *testing.T) {
	name := randomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read test
			{
				Config: testAccCreateConfig(`
				resource "turso_database" "test" {
					group = "test"
					name = "` + name + `"
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("turso_database.test", tfjsonpath.New("instances"), knownvalue.MapSizeExact(1)),
					statecheck.ExpectKnownValue("turso_database.test", tfjsonpath.New("allow_attach"), knownvalue.Bool(false)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("turso_database.test", "name", name),
					resource.TestCheckResourceAttr("turso_database.test", "group", "test"),
				),
			},

			// ImportState test
			{
				ResourceName:                         "turso_database.test",
				ImportStateIdPrefix:                  "test/", // group name
				ImportStateId:                        name,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},

			// Update and Read test
			{
				Config: testAccCreateConfig(`
				resource "turso_database" "test" {
					group = "test"
					name = "` + name + `"
					allow_attach = true
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("turso_database.test", tfjsonpath.New("instances"), knownvalue.MapSizeExact(1)),
					statecheck.ExpectKnownValue("turso_database.test", tfjsonpath.New("allow_attach"), knownvalue.Bool(true)),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("turso_database.test", "name", name),
					resource.TestCheckResourceAttr("turso_database.test", "group", "test"),
				),
			},
		},
	})
}
