---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "xenserver_pif_configure Resource - xenserver"
subcategory: ""
description: |-
  PIF configuration resource which is used to update the existing PIF parameters.
  Noted that no new PIF will be deployed when terraform apply is executed. Additionally, when it comes to terraform destroy, it actually has no effect on this resource.
---

# xenserver_pif_configure (Resource)

PIF configuration resource which is used to update the existing PIF parameters. 

 Noted that no new PIF will be deployed when `terraform apply` is executed. Additionally, when it comes to `terraform destroy`, it actually has no effect on this resource.

## Example Usage

```terraform
data "xenserver_pif" "pif_eth1" {
  device = "eth1"
}

# Update single PIF configuration
resource "xenserver_pif_configure" "pif_update" {
  uuid = data.xenserver_pif.pif_eth1.data_items[0].uuid
  disallow_unplug = true
  interface = {
    mode = "Static"
    ip = "10.62.49.185"
    netmask = "255.255.240.0"
  }
}

# Update multiple PIFs configuration
locals {
  pif_data = tomap({for element in data.xenserver_pif.pif_eth1.data_items: element.uuid => element})
}

resource "xenserver_pif_configure" "pif_update" {
  for_each = local.pif_data
  uuid     = each.key
  interface = {
    mode = "DHCP"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `uuid` (String) The UUID of the PIF.

### Optional

- `disallow_unplug` (Boolean) Set to `true` if you want to prevent this PIF from being unplugged.
- `interface` (Attributes) The IP interface of the PIF. Currently only support IPv4. (see [below for nested schema](#nestedatt--interface))

### Read-Only

- `id` (String) The test ID of the PIF.

<a id="nestedatt--interface"></a>
### Nested Schema for `interface`

Required:

- `mode` (String) The protocol define the primary address of this PIF, for example, `"None"`, `"DHCP"`, `"Static"`.

Optional:

- `dns` (String) Comma separated list of the IP addresses of the DNS servers to use.
- `gateway` (String) The IP gateway.
- `ip` (String) The IP address.
- `name_label` (String) The name of the interface in IP Address Configuration.
- `netmask` (String) The IP netmask.

## Import

Import is supported using the following syntax:

```shell
terraform import xenserver_pif_configure.pif_update 00000000-0000-0000-0000-000000000000

# when use 'for_each' in resource
terraform  import  xenserver_pif_configure.pif_update[\"{each.key}\"] 00000000-0000-0000-0000-000000000000
```