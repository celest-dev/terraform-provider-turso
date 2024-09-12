package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceGroup(t *testing.T) {
	name := randomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read test
			{
				Config: testAccCreateConfig(`
				resource "turso_group" "test" {
					name = "` + name + `"
					primary = "sjc"
					locations = ["sjc", "dfw"]
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("turso_group.test", tfjsonpath.New("id"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("turso_group.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("turso_group.test", tfjsonpath.New("locations"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("sjc"),
						knownvalue.StringExact("dfw"),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("turso_group.test", "id", name),
					resource.TestCheckResourceAttr("turso_group.test", "name", name),
				),
			},

			// ImportState and Update test
			{
				ResourceName:      "turso_group.test",
				ImportStateId:     name,
				ImportState:       true,
				ImportStateVerify: true,

				Config: testAccCreateConfig(`
				resource "turso_group" "test" {
					name = "` + name + `"
					primary = "sjc"
					locations = ["sjc", "dfw", "sea"]
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("turso_group.test", tfjsonpath.New("id"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("turso_group.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("turso_group.test", tfjsonpath.New("locations"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("sjc"),
						knownvalue.StringExact("dfw"),
						knownvalue.StringExact("sea"),
					})),
				},
			},
		},
	})
}
