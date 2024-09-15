package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateConfig(`
				data "turso_groups" "test" {}
				data "turso_group" "test" {
					id = data.turso_groups.test.groups[0].name
				}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.turso_groups.test", tfjsonpath.New("groups"), listNotEmpty{}),
					statecheck.ExpectKnownValue("data.turso_groups.test", tfjsonpath.New("groups"), listOfNonNulls{}),
					statecheck.ExpectKnownValue("data.turso_group.test", tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.turso_group.test", tfjsonpath.New("group"), knownvalue.NotNull()),
				},
			},
		},
	})
}
