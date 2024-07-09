locals {
  env_vars = { for tuple in regexall("export\\s*(\\S*)\\s*=\\s*(\\S*)\\s*", file("../../.env")) : tuple[0] => tuple[1] }
}

terraform {
  required_providers {
    xenserver = {
      source = "xenserver/xenserver"
    }
  }
}

provider "xenserver" {
  host     = local.env_vars["XENSERVER_HOST"]
  username = local.env_vars["XENSERVER_USERNAME"]
  password = local.env_vars["XENSERVER_PASSWORD"]
}

data "xenserver_sr" "sr" {
  name_label = "Local storage"
}

resource "xenserver_vdi" "vdi1" {
  name_label   = "local-storage-vdi-1"
  sr_uuid      = data.xenserver_sr.sr.data_items[0].uuid
  virtual_size = 100 * 1024 * 1024 * 1024
}
resource "xenserver_vdi" "vdi2" {
  name_label   = "local-storage-vdi-2"
  sr_uuid      = data.xenserver_sr.sr.data_items[0].uuid
  virtual_size = 100 * 1024 * 1024 * 1024
}

data "xenserver_network" "network" {}

resource "xenserver_vm" "vm" {
  name_label    = "A test virtual-machine"
  template_name = "Windows 11"
  hard_drive = [
    {
      vdi_uuid = xenserver_vdi.vdi1.id,
      bootable = true,
      mode     = "RW"
    },
    {
      vdi_uuid = xenserver_vdi.vdi2.id,
      bootable = false,
      mode     = "RO"
    },
  ]
  network_interface = [
    # {
    #   network_uuid = data.xenserver_network.network.data_items[0].uuid,
    #   mtu          = 1500,
    #   mac          = "00:11:22:33:44:55",
    #   other_config = {
    #     ethtool-gso = "off"
    #     ethtool-ufo = "off"
    #     ethtool-tso = "off"
    #     ethtool-sg = "off"
    #     ethtool-tx = "off"
    #     ethtool-rx = "off"
    #   }
    # },
    {
      other_config = {
        ethtool-gso = "off"

      }
      mac          = "00:11:22:33:44:55",
      network_uuid = data.xenserver_network.network.data_items[1].uuid,
    },
  ]
}

output "vm_out" {
  value = xenserver_vm.vm
}


