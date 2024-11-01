---
page_title: "XenServer Provider"
subcategory: ""
description: |-
  The XenServer provider can be used to manage and deploy XenServer resources. Before using it, you must configure the provider with the appropriate credentials. Documentation regarding the data sources and resources supported by the XenServer provider can be found in the navigation on the left.
---

# XenServer Provider

The XenServer provider can be used to manage and deploy XenServer resources. Before using it, you must configure the provider with the appropriate credentials. Documentation regarding the data sources and resources supported by the XenServer provider can be found in the navigation on the left.

## Example Usage

```terraform
provider "xenserver" {
  host     = "https://10.1.1.1"
  username = "root"
  password = "test123"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `host` (String) The address of target XenServer host.<br />Can be set by using the environment variable **XENSERVER_HOST**.
- `password` (String, Sensitive) The password of target XenServer host.<br />Can be set by using the environment variable **XENSERVER_PASSWORD**.
- `username` (String) The user name of target XenServer host.<br />Can be set by using the environment variable **XENSERVER_USERNAME**.
