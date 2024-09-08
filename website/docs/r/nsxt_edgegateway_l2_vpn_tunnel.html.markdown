---
layout: "vcd"
page_title: "VMware Cloud Director: vcloud_nsxt_edgegateway_l2_vpn_tunnel"
sidebar_current: "docs-vcd-resource-nsxt-edgegateway-l2-vpn-tunnel"
description: |-
  Provides a resource to manage NSX-T Edge Gateway L2 VPN Tunnel sessions and their configurations.
---

# vcd\_nsxt\_edgegateway\_l2\_vpn\_tunnel

Supported in provider *v3.11+* and VCD *10.4+* with NSX-T.

Provides a resource to manage NSX-T Edge Gateway L2 VPN Tunnel sessions and their configurations.
<a id="example-usage"></a>
## Example Usage (Both server and client tunnel sessions connecting two Edge Gateways)

```hcl
data "vcloud_org_vdc" "existing" {
  name = "existing-vdc"
}

data "vcloud_nsxt_edgegateway" "server-testing" {
  owner_id = data.vcloud_org_vdc.existing.id
  name     = "server-testing"
}

data "vcloud_nsxt_edgegateway" "client-testing" {
  owner_id = data.vcloud_org_vdc.existing.id
  name     = "client-testing"
}

resource "vcloud_nsxt_edgegateway_l2_vpn_tunnel" "server-session" {
  org             = "datacloud"
  edge_gateway_id = data.vcloud_nsxt_edgegateway.server-testing.id

  name        = "server-session"
  description = "example description"

  session_mode              = "SERVER"
  enabled                   = true
  connector_initiation_mode = "ON_DEMAND"

  # must be sub-allocated on the Edge Gateway
  local_endpoint_ip  = "10.10.50.2"
  tunnel_interface   = "192.168.0.1/24"
  remote_endpoint_ip = "1.2.2.3"

  stretched_network {
    network_id = data.vcloud_network_routed_v2.test_network_server.id
  }

  pre_shared_key = "secret_passphrase"
}

resource "vcloud_nsxt_edgegateway_l2_vpn_tunnel" "client-session" {
  org = "datacloud"

  # Note that this is a different Edge Gateway, as one Edge Gateway
  # can function only in SERVER or CLIENT mode
  edge_gateway_id = data.vcloud_nsxt_edgegateway.client-testing.id

  name        = "client-session"
  description = "example description"

  session_mode = "CLIENT"
  enabled      = true

  # must be sub-allocated on the Edge Gateway
  local_endpoint_ip  = "101.22.30.3"
  remote_endpoint_ip = "1.2.2.3"

  stretched_network {
    network_id = data.vcloud_network_routed_v2.test_network_client.id
    # CLIENT mode sessions need to define a tunnel ID for every stretched network
    tunnel_id = 1
  }

  stretched_network {
    network_id = data.vcloud_network_routed_v2.test_network_client_other.id
    tunnel_id  = 2
  }

  # Be aware that if there are changes in the `server-session`, the peer_code
  # will be updated as well, so `terraform apply` needs to run twice
  peer_code = vcloud_nsxt_edgegateway_l2_vpn_tunnel.server-session.peer_code
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at 
  provider level. Useful when connected as sysadmin working across different organisations
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-T only). 
  Can be looked up using [`vcloud_nsxt_edgegateway`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source
* `name` - (Required) The name of the tunnel.
* `description` - (Optional) The description of the tunnel.
* `session_mode` - (Required) Mode of the tunnel session (SERVER or CLIENT).
* `enabled` - (Optional) State of the `SERVER` mode session, always set to `true` for `CLIENT` 
  mode sessions. Default is `true`.
* `connector_initiation_mode` - (Optional) Mode in which the connection is formed. 
  Required for `SERVER` mode sessions. One of:
	* `INITIATOR` - Local endpoint initiates tunnel setup and will also respond to 
  incoming tunnel setup requests from the peer gateway.
	* `RESPOND_ONLY` - Local endpoint shall only respond to incoming tunnel setup 
  requests, it shall not initiate the tunnel setup.
	* `ON_DEMAND` - In this mode local endpoint will initiate tunnel creation once 
  first packet matching the policy rule is received, and will also respond to 
  incoming initiation requests.
* `local_endpoint_ip` - (Required) The IP address corresponding to the Edge 
  Gateway the tunnel is being configured on. The IP must be sub-allocated 
  on the Edge Gateway.
* `remote_endpoint_ip` - (Required) The IP address of the remote endpoint, which 
corresponds to the device on the remote site terminating the VPN tunnel.
* `tunnel_interface` - (Optional) The network CIDR block over which the session 
  interfaces. Relevant only for `SERVER` mode sessions. If not provided, Cloud 
  Director will attempt to automatically allocate a tunnel interface.
* `pre_shared_key` - (Optional) The key that is used for authenticating the 
  connection. Required for `SERVER` mode sessions.
* `peer_code` - (Optional) Encoded string that contains the whole configuration 
  of a `SERVER` mode session including the pre-shared key so it is user's 
  responsibility to secure it. Computed for `SERVER` mode sessions, required for 
  `CLIENT` mode sessions. See [example](#example-usage) 
  for a solution implemented fully in Terraform.
* `stretched_network` - (Optional) One or more stretched networks for the tunnel. 
  See [`stretched_network`](#stretched-network) for more detail.

## Stretched network

* `network_id` - (Required) Network ID of a routed network on the Edge Gateway. 
  Can be looked up using [`vcloud_network_routed_v2`](/providers/vmware/vcd/latest/docs/data-sources/network_routed_v2) 
  datasource.
* `tunnel_id` - (Optional) Tunnel ID of the network on the tunnel. Required for 
  `CLIENT` mode sessions, computed for `SERVER` mode sessions.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing L2 VPN Tunnel configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is below:

```hcl
data "vcloud_org" "my_org" {
  name = "my-org"
}

data "vcloud_org_vdc" "my-vdc-or-vdc-group" {
  name = "my-vdc"
  org  = "my-org"
}

data "vcloud_nsxt_edgegateway" "my-edge-gateway" {
  name     = "my-edge-gateway"
  owner_id = data.vcloud_org_vdc.my-vdc-or-vdc-group.id
}

resource "vcloud_nsxt_edgegateway_l2_vpn_tunnel" "imported" {
  org  = "my-org"
  name = "my-tunnel"
}
```

```
terraform import vcloud_nsxt_edgegateway_l2_vpn_tunnel.imported my-org.my-vdc-or-vdc-group.my-edge-gateway.l2_vpn_tunnel
```

The above would import the `l2_vpn_tunnel` L2 VPN Tunnel that is defined in
`my-edge-gateway` NSX-T Edge Gateway. Edge Gateway should be located in `my-vdc-or-vdc-group` VDC or
VDC Group in Org `my-org`

[docs-import]: https://www.terraform.io/docs/import/
