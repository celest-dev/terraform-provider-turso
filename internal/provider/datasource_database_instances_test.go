package provider

import (
	"fmt"
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
				},
			},
		},
	})
}

type listNotEmpty struct{}

func (listNotEmpty) CheckValue(v any) error {
	val, ok := v.([]any)

	if !ok {
		return fmt.Errorf("expected []any value for ListNotEmpty check, got: %T", v)
	}

	if len(val) == 0 {
		return fmt.Errorf("expected non-empty list for ListNotEmpty check, but list was empty")
	}

	return nil
}

func (listNotEmpty) String() string {
	return "non-empty list"
}
