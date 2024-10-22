package xenserver

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func pifResource(eth_index string) string {
	return fmt.Sprintf(`
// configure eth1 PIF IP
data "xenserver_pif" "pif_data" {
  device = "eth%s"
}

resource "xenserver_pif_configure" "pif1" {
  uuid = data.xenserver_pif.pif_data.data_items[0].uuid
  interface = {
    mode = "DHCP"
  }
}

resource "xenserver_pif_configure" "pif2" {
  uuid = data.xenserver_pif.pif_data.data_items[1].uuid
  interface = {
    mode = "DHCP"
  }
}

resource "xenserver_pif_configure" "pif3" {
  uuid = data.xenserver_pif.pif_data.data_items[2].uuid
  interface = {
    mode = "DHCP"
  }
}

data "xenserver_pif" "pif" {
    device = "eth%s"
}
`, eth_index, eth_index)
}

func managementNetwork(index string) string {
	return fmt.Sprintf(`
	management_network = data.xenserver_pif.pif.data_items[%s].network
	`, index)
}

func testPoolResource(name_label string,
	name_description string,
	storage_location string,
	management_network string,
	supporter_params string,
	eject_supporter string) string {
	return fmt.Sprintf(`
resource "xenserver_sr_nfs" "nfs" {
	name_label       = "NFS"
	version          = "3"
	storage_location = "%s"
}

data "xenserver_host" "supporter" {
  is_coordinator = false
}

resource "xenserver_pool" "pool" {
    name_label   = "%s"
	name_description = "%s"
    default_sr = xenserver_sr_nfs.nfs.uuid
	%s
	%s
	%s
}
`, storage_location,
		name_label,
		name_description,
		management_network,
		supporter_params,
		eject_supporter)
}

func testJoinSupporterParams(supporterHost string, supporterUsername string, supporterPassowd string) string {
	return fmt.Sprintf(`
	join_supporters = [
		{
		    host = "%s"
			username = "%s"
			password = "%s"
		}
    ]
`, supporterHost, supporterUsername, supporterPassowd)
}

func ejectSupporterParams(index string) string {
	return fmt.Sprintf(`
	eject_supporters = [
		data.xenserver_host.supporter.data_items[%s].uuid
	]
`, index)
}

func TestAccPoolResource(t *testing.T) {
	// skip test if TEST_POOL is not set
	if os.Getenv("TEST_POOL") == "" {
		t.Skip("Skipping TestAccPoolResource test due to TEST_POOL not set")
	}

	storageLocation := os.Getenv("NFS_SERVER") + ":" + os.Getenv("NFS_SERVER_PATH")
	joinSupporterParams := testJoinSupporterParams(os.Getenv("SUPPORTER_HOST"), os.Getenv("SUPPORTER_USERNAME"), os.Getenv("SUPPORTER_PASSWORD"))
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for Default SR and Pool Join
			{
				Config: providerConfig + testPoolResource("Test Pool A",
					"Test Pool Join",
					storageLocation,
					"",
					joinSupporterParams,
					""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("xenserver_pool.pool", "name_label", "Test Pool A"),
					resource.TestCheckResourceAttr("xenserver_pool.pool", "name_description", "Test Pool Join"),
					resource.TestCheckResourceAttrSet("xenserver_pool.pool", "default_sr"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "xenserver_pool.pool",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"join_supporters"},
			},
			// Update and Read testing For Pool eject supporter
			{
				Config: providerConfig + testPoolResource("Test Pool B",
					"Test Pool Eject",
					storageLocation,
					"",
					"",
					ejectSupporterParams("1")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("xenserver_pool.pool", "name_label", "Test Pool B"),
					resource.TestCheckResourceAttr("xenserver_pool.pool", "name_description", "Test Pool Eject"),
				),
			},
			// Update and Read testing For Pool Management Network
			{
				Config: providerConfig + pifResource("3") + testPoolResource("Test Pool C",
					"Test Pool Management Network",
					storageLocation,
					managementNetwork("2"),
					"",
					""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("xenserver_pool.pool", "name_label", "Test Pool C"),
					resource.TestCheckResourceAttr("xenserver_pool.pool", "name_description", "Test Pool Management Network"),
					resource.TestCheckResourceAttrSet("xenserver_pool.pool", "management_network"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})

	// sleep 30s to wait for supporters and management network back to enable
	time.Sleep(30 * time.Second)
}