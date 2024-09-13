package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceGroup(t *testing.T) {
	name := randomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateConfig(`
				resource "turso_group" "test" {
					name = "` + name + `"
					primary = "sjc"
					locations = ["sjc", "dfw"]
				}
				data "turso_group" "test" {
					id = turso_group.test.id
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.turso_group.test", tfjsonpath.New("id"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("data.turso_group.test", tfjsonpath.New("group"), knownvalue.NotNull()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.turso_group.test", "id", name),
				),
			},
		},
	})
}
