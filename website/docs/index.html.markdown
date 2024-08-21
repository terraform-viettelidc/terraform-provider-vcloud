---
layout: "vcd"
page_title: "Provider: Viettel IDC Cloud"
sidebar_current: "docs-vcd-index"
description: |-
  The Viettel IDC Cloud provider is used to interact with the resources supported by Viettel IDC Cloud. The provider needs to be configured with the proper credentials before it can be used.
---

# Viettel IDC Cloud Provider 3.13

The Viettel IDC Cloud provider is used to interact with the resources supported by Viettel IDC Cloud. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources. Please refer to
[CHANGELOG.md](https://github.com/vmware/terraform-provider-vcd/blob/main/CHANGELOG.md)
to track feature additions.

~> **NOTE:** The Viettel IDC Cloud Provider documentation pages include *v2.x+* or *v3.x+* labels in resource and/or field
descriptions. These labels are designed to show at which provider version a certain feature was introduced.
When upgrading the provider please check for such labels for the resources you are using.

## Supported VCD Versions

The following Cloud Director versions are supported by this provider:

* 10.4
* 10.5

Also Cloud Director Service (CDS) is supported.

## Connecting as Org Admin

The most common - tenant - use case when you set user to organization administrator and when all resources are in a single organization. 

```hcl
# Configure the Viettel IDC Cloud Provider
provider "vcd" {
  user                 = var.vcloud_user
  password             = var.vcloud_pass
  auth_type            = "integrated"
  org                  = var.vcloud_org
  vdc                  = var.vcloud_vdc
  url                  = var.vcloud_url
  max_retry_timeout    = var.vcloud_max_retry_timeout
  allow_unverified_ssl = var.vcloud_allow_unverified_ssl
}

# Create a new network in organization and VDC defined above
resource "vcloud_network_routed" "net" {
  # ...
}
```

## Connecting as Sys Admin

When you want to manage resources across different organizations from a single configuration.

```hcl
# Configure the Viettel IDC Cloud Provider
provider "vcd" {
  user                 = "administrator"
  password             = var.vcloud_pass
  auth_type            = "integrated"
  org                  = "System"
  url                  = var.vcloud_url
  max_retry_timeout    = var.vcloud_max_retry_timeout
  allow_unverified_ssl = var.vcloud_allow_unverified_ssl
}

# Create a new network in some organization and VDC
resource "vcloud_network_routed" "net1" {
  org = "Org1"
  vdc = "Org1VDC"

  # ...
}

# Create a new network in a different organization and VDC
resource "vcloud_network_routed" "net2" {
  org = "Org2"
  vdc = "Org2VDC"

  # ...
}
```

## Connecting as Sys Admin with Default Org and VDC

When you want to manage resources across different organizations but set a default one. 

```hcl
# Configure the Viettel IDC Cloud Provider
provider "vcd" {
  user                 = "administrator"
  password             = var.vcloud_pass
  auth_type            = "integrated"
  sysorg               = "System"
  org                  = var.vcloud_org # Default for resources
  vdc                  = var.vcloud_vdc # Default for resources
  url                  = var.vcloud_url
  max_retry_timeout    = var.vcloud_max_retry_timeout
  allow_unverified_ssl = var.vcloud_allow_unverified_ssl
}

# Create a new network in the default organization and VDC
resource "vcloud_network_routed" "net1" {
  # ...
}

# Create a new network in a specific organization and VDC
resource "vcloud_network_routed" "net2" {
  org = "OrgZ"
  vdc = "OrgZVDC"

  # ...
}
```

## Connecting with authorization or bearer token

You can connect using an authorization token instead of username and password.

```hcl
provider "vcd" {
  user                 = "none"
  password             = "none"
  auth_type            = "token"
  token                = var.token
  sysorg               = "System"
  org                  = var.vcloud_org # Default for resources
  vdc                  = var.vcloud_vdc # Default for resources
  url                  = var.vcloud_url
  max_retry_timeout    = var.vcloud_max_retry_timeout
  allow_unverified_ssl = var.vcloud_allow_unverified_ssl
}

# Create a new network in the default organization and VDC
resource "vcloud_network_routed" "net1" {
  # ...
}
```
When using a token, the fields `user` and `password` will be ignored, but they need to be in the script.

## Connecting with an API token/API token file

With VCD 10.3.1+, you can connect using an API token, as defined in the [documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-A1B3B2FA-7B2C-4EE1-9D1B-188BE703EEDE.html).
The API token is not a bearer token, but one will be created and automatically used by the Terraform provider when an API
token is supplied. You can create an API token file by utilizing the [`vcloud_api_token`][api-token] resource.

#### Example usage (API token)

```hcl
provider "vcd" {
  user                 = "none"
  password             = "none"
  auth_type            = "api_token"
  api_token            = var.api_token
  sysorg               = "System"
  org                  = var.vcloud_org # Default for resources
  vdc                  = var.vcloud_vdc # Default for resources
  url                  = var.vcloud_url
  max_retry_timeout    = var.vcloud_max_retry_timeout
  allow_unverified_ssl = var.vcloud_allow_unverified_ssl
}

# Create a new network in the default organization and VDC
resource "vcloud_network_routed" "net1" {
  # ...
}
```

#### Example usage (API token file)

```hcl
provider "vcd" {
  user                 = "none"
  password             = "none"
  auth_type            = "api_token_file"
  api_token            = "token.json"
  sysorg               = "System"
  org                  = var.vcloud_org # Default for resources
  vdc                  = var.vcloud_vdc # Default for resources
  url                  = var.vcloud_url
  max_retry_timeout    = var.vcloud_max_retry_timeout
  allow_unverified_ssl = var.vcloud_allow_unverified_ssl
}

# Create a new network in the default organization and VDC
resource "vcloud_network_routed" "net1" {
  # ...
}
```

The file containing the API token needs to be readable and writable, in `json` format with the API key. e.g:
```json
{"refresh_token":"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}
```


Note that when connecting with API tokens you can't create or modify users, roles, global roles, or rights bundles.

## Connecting with a Service Account API token

With VCD 10.4.0+, similar to API token file, you can connect using a service account API token, as 
defined in the 
[documentation](https://blogs.vmware.com/cloudprovider/2022/07/cloud-director-service-accounts.html). 
Because a new API token is provided on every authentication request, 
the user is required to provide a readable+writable file in `json` 
format with the current API key. e.g:
```json
{"refresh_token":"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}
```
Note that the file will be rewritten at every usage, and the updated file will have additional fields, such as
```json
{
  "token_type": "Service Account",
  "refresh_token": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "updated_by": "terraform-provider-viettelidc/v3.9.0 (darwin/arm64; isProvider:true)",
  "updated_on": "2023-04-18T14:33:07+02:00"
 }
```

The API token file is **sensitive data** and it's up to the user to secure it.

~> **NOTE:** The service account needs to be in `Active Stage` and 
it's up to the user to provide the initial API token. A service account 
can be created using the [`service_account`][service-account] resource, 
also it can be done using a sample shell script for creating, authorizing 
and activating a VCD Service Account can be found in the 
[repository][service-account-script]

```hcl
provider "vcd" {
  auth_type                  = "service_account_token_file"
  service_account_token_file = "token.json"
  sysorg                     = "System"
  org                        = var.vcloud_org # Default for resources
  vdc                        = var.vcloud_vdc # Default for resources
  url                        = var.vcloud_url
  max_retry_timeout          = var.vcloud_max_retry_timeout
  allow_unverified_ssl       = var.vcloud_allow_unverified_ssl
}
```

## Shell script to obtain a bearer token
To obtain a bearer token you can use this sample shell script:

```sh
#!/bin/bash
user=$1
password=$2
org=$3
IP=$4

if [ -z "$IP" ]
then
    echo "Syntax $0 user password organization hostname_or_IP_address"
    exit 1
fi

options=""
os=$(uname -s)
is_linux=$(echo "$os" | grep -i linux)
if [ -n "$is_linux" ]
then
  options="-w 0"
fi

auth=$(echo -n "$user@$org:$password" |base64 $options)

curl -I -k --header "Accept: application/*;version=32.0" \
    --header "Authorization: Basic $auth" \
    --request POST https://$IP/api/sessions
```

If successful, the output of this command will include lines like the following:

```
X-VCLOUD-AUTHORIZATION: 08a321735de84f1d9ec80c3b3e18fa8b
X-VMWARE-VCLOUD-ACCESS-TOKEN: eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhZG1pbmlzdHJhdG9yIiwiaXNzIjoiYTkzYzlkYjktNzQ3MS0zMTkyLThkMDktYThmN2VlZGE4NWY5QGY5MDZlODE1LTM0NjgtNGQ0ZS04MmJlLTcyYzFjMmVkMTBiMyIsImV4cCI6MTYwNzUxMjgyOCwidmVyc2lvbiI6InZjbG91ZF8xLjAiLCJqdGkiOiJjY2IwZjIwN2JjY2Y0NmYwYmEwNTcyNzgxZDQyNDg2MyJ9.SMjp5wsSd7CXGMdlj-weeCRdr5AazA74pwwx2w3Eqh3RdzyiEMvQfWQAuPAQjM1oOsEUnFOg2u0gYsnIyQg_p7kzXKPQwPNz3BPi0tm2DxxQtQVhOBRXCqUJ9OmRlMVu7FZZ6gKD4GhpbTkZyKMN_IgOFkkt8iXs1-weNZw5TmyVHeWiJdV0JFM45CV47jQNdQMy4OSsU-CqE2VVLOK83oJhRnlnc3O4OAAIfuVZ4SLWqgi1lIoc2vbZv0HYeWO7L_2pGfmja8CVzVhPrgIGEoDhXnvO29z1ToEXRnyMKh9cisiRkhUISpsh4aHRGUUzaZYeOejVX3PAO9aCX3iYWA

The string after `X-VCLOUD-AUTHORIZATION:` is the old (deprecated) token.
The string after `X-VMWARE-VCLOUD-ACCESS-TOKEN` is the bearer token
```

Either token will grant the same abilities as the account used to run the above script. Note, however, that the deprecated
token may not work in recent VCD versions.

Using a token produced by an org admin to run a task that requires a system administrator will fail.

## Connecting with SAML user using Microsoft Active Directory Federation Services (ADFS) and setting custom Relaying Party Trust Identifier

Take special attention to `user`, `use_saml_adfs` and `saml_rpt_id` fields.

```hcl
# Configure the Viettel IDC Cloud Provider
provider "vcd" {
  user      = "test@contoso.com"
  password  = var.vcloud_pass
  sysorg    = "my-org"
  auth_type = "saml_adfs"
  # If `saml_adfs_rpt_id` is not specified - VCD SAML Entity ID will be used automatically
  saml_adfs_rpt_id     = "my-custom-rpt-id"
  org                  = var.vcloud_org # Default for resources
  vdc                  = var.vcloud_vdc # Default for resources
  url                  = var.vcloud_url
  max_retry_timeout    = var.vcloud_max_retry_timeout
  allow_unverified_ssl = var.vcloud_allow_unverified_ssl
}
```

## Argument Reference

The following arguments are used to configure the Viettel IDC Cloud Provider:

* `user` - (Required) This is the username for Cloud Director API operations. Can also be specified
  with the `vcloud_USER` environment variable. *v2.0+* `user` may be "administrator" (set `org` or
  `sysorg` to "System" in this case). 
  *v2.9+* When using with SAML and ADFS - username format must be in Active Directory format -
  `user@contoso.com` or `contoso.com\user` in combination with `use_saml_adfs` option.
  
* `password` - (Required) This is the password for Cloud Director API operations. Can
  also be specified with the `vcloud_PASSWORD` environment variable.

* `auth_type` - (Optional) `integrated`, `token`, `api_token`, `service_account_token_file` or `saml_adfs`. 
  Default is `integrated`. Can also be set with `vcloud_AUTH_TYPE` environment variable. 
  * `integrated` - VCD local users and LDAP users (provided LDAP is configured for Organization).
  * `saml_adfs` allows to use SAML login flow with Active Directory Federation
  Services (ADFS) using "/adfs/services/trust/13/usernamemixed" endpoint. Please note that
  credentials for ADFS should be formatted as `user@contoso.com` or `contoso.com\user`. 
  `saml_adfs_rpt_id` can be used to specify a different RPT ID.
  * `token` allows to specify token in `token` field.
  * `api_token` allows to specify an API token.
  * `api_token_file` allows to specify a file containing an API token.
  * `service_account_token_file` allows to specify a file containing a service account's token.
  
* `token` - (Optional; *v2.6+*) This is the bearer token that can be used instead of username
   and password (in combination with field `auth_type=token`). When this is set, username and
   password will be ignored, but should be left in configuration either empty or with any custom
   values. A token can be specified with the `vcloud_TOKEN` environment variable.
   Both a (deprecated) authorization token or a bearer token (*v3.1+*) can be used in this field.

* `api_token` - (Optional; *v3.5+*) This is the API token that a System or organization administrator can create and 
   distribute to users. It is used instead of username and password (in combination with `auth_type=api_token`). When
   this field is filled, username and password are ignored. An API token can also be specified with the `vcloud_API_TOKEN`
   environment variable. This token requires at least VCD 10.3.1. There are restrictions to its use, as defined in
   [the documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-A1B3B2FA-7B2C-4EE1-9D1B-188BE703EEDE.html)

* `api_token_file` - (Optional; *v3.10+*)) Same as `api_token`, only provided 
   as a JSON file. Can also be specified with the `vcloud_API_TOKEN_FILE` environment variable.
 
* `service_account_token_file` - (Optional; *v3.9+, VCD 10.4+*) This is the file that contains a Service Account API token. The
   path to the file could be provided as absolute or relative to the working directory. It is used instead of username
   and password (in combination with `auth_type=service_account_token_file`. The file can also be specified with the 
   `vcloud_SA_TOKEN_FILE` environment variable. There are restrictions to its use, as defined in 
   [the documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-8CD3C8BE-3187-4769-B960-3E3315492C16.html)

* `allow_service_account_token_file` - (Optional; *v3.9+, VCD 10.4+*) When using `auth_type=service_account_token_file`,
  if set to `true`, will suppress a warning to the user about the service account token file containing *sensitive information*.
  Can also be set with `vcloud_ALLOW_SA_TOKEN_FILE`.

* `saml_adfs_rpt_id` - (Optional) When using `auth_type=saml_adfs` VCD SAML entity ID will be used
  as Relaying Party Trust Identifier (RPT ID) by default. If a different RPT ID is needed - one can
  set it using this field. It can also be set with `vcloud_SAML_ADFS_RPT_ID` environment variable.

* `org` - (Required) This is the Cloud Director Org on which to run API
  operations. Can also be specified with the `vcloud_ORG` environment
  variable.  
  *v2.0+* `org` may be set to "System" when connection as Sys Admin is desired
  (set `user` to "administrator" in this case).  
  Note: `org` value is case sensitive.
  
* `sysorg` - (Optional; *v2.0+*) - Organization for user authentication. Can also be
   specified with the `vcloud_SYS_ORG` environment variable. Set `sysorg` to "System" and
   `user` to "administrator" to free up `org` argument for setting a default organization
   for resources to use.
   
* `url` - (Required) This is the URL for the Cloud Director API endpoint. e.g.
  https://server.domain.com/api. Can also be specified with the `vcloud_URL` environment variable.
  
* `vdc` - (Optional) This is the virtual datacenter within Cloud Director to run
  API operations against. If not set the plugin will select the first virtual
  datacenter available to your Org. Can also be specified with the `vcloud_VDC` environment
  variable.
  
* `max_retry_timeout` - (Optional) This provides you with the ability to specify the maximum
  amount of time (in seconds) you are prepared to wait for interactions on resources managed
  by Cloud Director to be successful. If a resource action fails, the action will be retried
  (as long as it is still within the `max_retry_timeout` value) to try and ensure success.
  Defaults to 60 seconds if not set.
  Can also be specified with the `vcloud_MAX_RETRY_TIMEOUT` environment variable.
  
* `maxRetryTimeout` - (Deprecated) Use `max_retry_timeout` instead.

* `allow_unverified_ssl` - (Optional) Boolean that can be set to true to
  disable SSL certificate verification. This should be used with care as it
  could allow an attacker to intercept your auth token. If omitted, default
  value is false. Can also be specified with the
  `vcloud_ALLOW_UNVERIFIED_SSL` environment variable.

* `logging` - (Optional; *v2.0+*) Boolean that enables API calls logging from upstream library `go-vcloud-director`. 
   The logging file will record all API requests and responses, plus some debug information that is part of this 
   provider. Logging can also be activated using the `vcloud_API_LOGGING` environment variable.

* `logging_file` - (Optional; *v2.0+*) The name of the log file (when `logging` is enabled). By default is 
  `go-vcloud-director` and it can also be changed using the `vcloud_API_LOGGING_FILE` environment variable.
  
* `import_separator` - (Optional; *v2.5+*) The string to be used as separator with `terraform import`. By default
  it is a dot (`.`).

* `ignore_metadata_changes` - (Optional; Experimental; *v3.10+*) Use one or more of these blocks to ignore specific metadata entries from being changed by this Terraform provider
  after creation or when they were created outside Terraform.
  See ["Ignore Metadata Changes"](#ignore-metadata-changes) for more details.

## Ignore metadata changes

=> This is an **EXPERIMENTAL FEATURE** that may change in a future release.

One or more `ignore_metadata_changes` blocks can be optionally set in the provider configuration, which will allow to ignore specific `metadata_entry`
items during all Terraform operations. This is useful, for example, to avoid removing metadata entries that were created
by an external actor, or after they were created by Terraform.

~> Note that this feature is only considered when using the `metadata_entry` argument in the resources and data sources that support
it. In other words, to ignore metadata when the deprecated `metadata` argument is used, please use the native Terraform `lifecycle.ignore_changes` block.

~> Be aware that setting a `metadata_entry` in the Terraform configuration that matches any `ignore_metadata_changes` can produce inconsistent
results, as the metadata will be stored in state but nothing will be done in VCD. Using `ignore_metadata_changes` with matching metadata entries
in the code is NOT recommended. In the event that it contains such conflict, though, the ensuing action can be controlled with
`conflict_action`, which can be `error`, `warn` or `none`.

The available sub-attributes for `ignore_metadata_changes` are:

* `resource_type` - (Optional) Specifies the resource type which metadata needs to be ignored. If set, the resource type must be one of:
  *"vcloud_catalog"*, *"vcloud_catalog_item"*, *"vcloud_catalog_media"*, *"vcloud_catalog_vapp_template"*, *"vcloud_independent_disk"*, *"vcloud_network_direct"*,
  *"vcloud_network_isolated"*, *"vcloud_network_isolated_v2"*, *"vcloud_network_routed"*, *"vcloud_network_routed_v2"*, *"vcloud_org"*, *"vcloud_org_vdc"*, *"vcloud_provider_vdc"*,
  *"vcloud_rde" (v3.11+)*, *"vcloud_storage_profile"*, *"vcloud_vapp"*, *"vcloud_vapp_vm"* or *"vcloud_vm"*, which are the resources compatible with `metadata_entry`.
* `resource_name`- (Optional) Specifies the name of the entity in VCD which metadata needs to be ignored. This attribute can be used with
   any kind of `resource_type`, except for *vcloud_storage_profile* which **cannot be filtered by name**.
* `key_regex`- (Optional) A regular expression that can filter out metadata keys that match. Either `key_regex` or `value_regex` are required on each block. 
* `value_regex`- (Optional) A regular expression that can filter out metadata values that match. Either `key_regex` or `value_regex` are required on each block.
* `conflict_action` - (Optional) Defines what to do if a conflict exists between a `metadata_entry` that is managed
  by Terraform, and it matches the criteria defined in the `ignore_metadata_changes` block, as the metadata will be stored in state but nothing will be done in VCD.
  If the value is `error`, when this happens, any read operation (like a Plan or Refresh) will fail. When the value is `warn`, it will just give a warning but the operation will continue,
  and with the `none` value nothing will be shown. Defaults to `error`.

~> The `conflict_action` mechanism will be evaluated on every read, including `terraform destroy`, as it will trigger a refresh before deleting
resources. To avoid this situation, we can use the `-refresh=false` option.

Note that these attributes **are evaluated as a logical `and`**. This means that the snippet below would ignore all metadata entries
that belong to the specific Organization named "client1" **and** which keys match the regular expression `[Ee]nvironment`:

```hcl
provider "vcd" {
  # ...
  ignore_metadata_changes {
    resource_type = "vcloud_org"
    resource_name = "client1"
    key_regex     = "[Ee]nvironment"
    # Setting this value to 'warn' will make all 'metadata_entry' entries that
    # are managed by Terraform and that are ignored to give a warning to the user.
    conflict_action = "warn"
  }
}
```

We can have more than one block, to ignore more entries:

```hcl
provider "vcd" {
  # ...

  # Filters all metadata with key "Environment" or "environment" in all VCD objects with any name.
  ignore_metadata_changes {
    key_regex = "^[Ee]nvironment$"
  }

  # Filters all metadata with key "NiceMetadataKey" in all VCD objects named "SpecificName".
  ignore_metadata_changes {
    resource_name = "SpecificName"
    key_regex     = "^NiceMetadataKey$"
  }

  # Filters all metadata with values "Yes" in the Organization named "Tatooine".
  ignore_metadata_changes {
    resource_type = "vcloud_org"
    resource_name = "Tatooine"
    value_regex   = "^Yes$"
  }
}

resource "vcloud_org" "my_org" {
  name = "MyOrg"
  # ...

  # This entry will be added, if this Organization has other metadata entries that
  # match the ones defined in the Provider `ignore_metadata_changes` blocks, they will not be
  # deleted.
  metadata_entry {
    key         = "OneKey"
    value       = "OneValue"
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }
}
```

Note that this argument **does not affect metadata of the [data source filters](/providers/vmware/vcd/latest/docs/guides/data_source_filters)**.

## Connection Cache (*2.0+*)

Cloud Director connection calls can be expensive, and if a definition file contains several resources, it may trigger 
multiple connections. There is a cache engine, disabled by default, which can be activated by the `vcloud_CACHE` 
environment variable. When enabled, the provider will not reconnect, but reuse an active connection for up to 20 
minutes, and then connect again.

[service-account]: /providers/vmware/vcd/latest/docs/resources/service_account
[service-account-script]: https://github.com/vmware/terraform-provider-vcd/blob/main/scripts/create_service_account.sh
[api-token]: /providers/vmware/vcd/latest/docs/resource/api_token

