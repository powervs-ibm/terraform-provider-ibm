---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_instance_network_attach"
description: |-
  Manages the attachment of a network to an instance in the Power Virtual Server cloud.
---

# ibm_pi_instance_network_attach

Attach and detach a network to an existing Power Systems Virtual Server instance. Networks specified in the `pi_network` block of `ibm_pi_instance` are set when the instance is created. Use this resource to attach a network to an instance that already exists. For more information, about IBM power virtual server cloud, see [getting started with IBM Power Systems Virtual Servers](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-getting-started).

## Example Usage

The following example attaches a network to an instance:

```terraform
resource "ibm_pi_instance_network_attach" "example" {
  pi_cloud_instance_id = "<value of the cloud instance id>"
  pi_instance_id       = "<value of the instance id>"
  pi_network_id        = "<value of the network id>"
}
```

The following example attaches a network with a specific IP address and network security group:

```terraform
resource "ibm_pi_instance_network_attach" "example" {
  pi_cloud_instance_id          = "<value of the cloud instance id>"
  pi_instance_id                = "<value of the instance id>"
  pi_ip_address                 = "192.168.0.10"
  pi_network_id                 = "<value of the network id>"
  pi_network_security_group_ids = ["<value of the network security group id>"]
}
```

### Notes

- Please find [supported Regions](https://cloud.ibm.com/apidocs/power-cloud#endpoint) for endpoints.
- If a Power cloud instance is provisioned at `lon04`, The provider level attributes should be as follows:
  - `region` - `lon`
  - `zone` - `lon04`

Example usage:

  ```terraform
    provider "ibm" {
      region    =   "lon"
      zone      =   "lon04"
    }
  ```

## Timeouts

The `ibm_pi_instance_network_attach` provides the following [timeouts](https://www.terraform.io/docs/language/resources/syntax.html) configuration options:

- **create** - (Default 10 minutes) Used for attaching a network to an instance.
- **delete** - (Default 10 minutes) Used for detaching a network from an instance.

## Argument Reference

Review the argument references that you can specify for your resource.

- `pi_cloud_instance_id` - (Required, String) The GUID of the service instance associated with an account.
- `pi_instance_id` - (Required, String) The unique identifier of the instance.
- `pi_ip_address` - (Optional, String) The requested IP address of the network interface.
- `pi_network_id` - (Required, String) The ID of the network to attach to the instance.
- `pi_network_security_group_ids` - (Optional, Set of String) IDs of the network security groups that the network interface is a member of. A maximum of 1 network security group is allowed. If not specified, the default network security group is used.

Changing any argument replaces the attachment: the network is detached and then attached again.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute reference after your resource is created.

- `external_ip` - (String) The external IP address of the network interface.
- `href` - (String) Link to this PVM instance network.
- `id` - (String) The unique identifier of the network attachment. The ID is composed of `<pi_cloud_instance_id>/<pi_instance_id>/<pi_network_id>`.
- `ip_address` - (String) The IP address of the network interface.
- `mac_address` - (String) The MAC address of the network interface.
- `network_interface_id` - (String) ID of the network interface.
- `network_name` - (String) The name of the network.
- `network_security_group_ids` - (Set of String) IDs of the network security groups that the network interface is a member of.
- `network_security_groups_href` - (List of String) Links to the network security groups that the network interface is a member of.
- `type` - (String) The type of the network.

## Import

The `ibm_pi_instance_network_attach` resource can be imported by using `pi_cloud_instance_id`, `pi_instance_id` and `pi_network_id`.

### Example

```sh
terraform import ibm_pi_instance_network_attach.example d7bec597-4726-451f-8a63-e62e6f19c32c/cea6651a-bc0a-4438-9f8a-a0770bbf3ebb/4726d7be-c597-4438-9f8a-cea6651abc0a
```
