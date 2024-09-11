---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "xenserver_sr_smb Resource - xenserver"
subcategory: ""
description: |-
  Provides an SMB storage repository resource.
---

# xenserver_sr_smb (Resource)

Provides an SMB storage repository resource.

## Example Usage

```terraform
resource "xenserver_sr_smb" "smb_test" {
  name_label       = "SMB storage"
  name_description = "A test SMB storage repository"
  storage_location = "\\\\server\\path"
  username         = "username"
  password         = "password"
}

resource "xenserver_sr_smb" "smb_test" {
  name_label       = "SMB storage"
  storage_location = <<-EOF
    \\server\path
EOF
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name_label` (String) The name of the SMB storage repository.
- `storage_location` (String) The server and server path of the SMB storage repository.<br />Follow the format `"\\\\server\\path"`.

-> **Note:** `storage_location` is not allowed to be updated.

### Optional

- `name_description` (String) The description of the SMB storage repository, default to be `""`.
- `password` (String, Sensitive) The password of the SMB storage repository. Used when creating the SR.

-> **Note:** This password will be stored in terraform state file, follow document [Sensitive values in state](https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables#sensitive-values-in-state) to protect your sensitive data.
- `username` (String) The username of the SMB storage repository. Used when creating the SR.

### Read-Only

- `id` (String) The test ID of the SMB storage repository.
- `uuid` (String) The UUID of the SMB storage repository.

## Import

Import is supported using the following syntax:

```shell
terraform import xenserver_sr_smb.smb_test 00000000-0000-0000-0000-000000000000
```