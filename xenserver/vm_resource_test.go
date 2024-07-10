package xenserver

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccVMResourceConfig(name_label string, memory int64, vcpu int64) string {
	return fmt.Sprintf(`
data "xenserver_sr" "sr" {
  name_label = "Local storage"
}

resource "xenserver_vdi" "vdi" {
  name_label       = "local-storage-vdi"
  sr_uuid          = data.xenserver_sr.sr.data_items[0].uuid
  virtual_size     = 100 * 1024 * 1024 * 1024
}

resource "xenserver_vm" "test_vm" {
  name_label = "%s"
  template_name = "Windows 11"
  static_mem_max = %d * 1024 * 1024 * 1024
  vcpus         = %d
  hard_drive = [
    { 
      vdi_uuid = xenserver_vdi.vdi.uuid,
      bootable = true,
      mode = "RW"
    },
  ]
  other_config = {
  	"flag" = "1"
  }
}
`, name_label, memory, vcpu)
}

func TestAccVMResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccVMResourceConfig("test vm 1", 4, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "name_label", "test vm 1"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "template_name", "Windows 11"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "static_mem_min", "4294967296"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "static_mem_max", "4294967296"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "dynamic_mem_min", "4294967296"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "dynamic_mem_max", "4294967296"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "vcpus", "4"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "cores_per_socket", "1"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "hard_drive.#", "1"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "hard_drive.0.%", "4"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "hard_drive.0.mode", "RW"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "hard_drive.0.bootable", "true"),
					resource.TestCheckResourceAttrSet("xenserver_vm.test_vm", "hard_drive.0.vbd_ref"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "other_config.%", "1"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "other_config.flag", "1"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("xenserver_vm.test_vm", "uuid"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "xenserver_vm.test_vm",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{},
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccVMResourceConfig("test vm 2", 3, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "name_label", "test vm 2"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "template_name", "Windows 11"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "static_mem_min", "3221225472"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "static_mem_max", "3221225472"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "dynamic_mem_min", "3221225472"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "dynamic_mem_max", "3221225472"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "vcpus", "3"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "cores_per_socket", "1"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "other_config.%", "1"),
					resource.TestCheckResourceAttr("xenserver_vm.test_vm", "other_config.flag", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
