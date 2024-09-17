---
layout: "ibm"
page_title: "IBM : ibm_pi_network_interface"
description: |-
  Manages pi_network_interface.
subcategory: "Power Systems"
---

# ibm_pi_network_interface

Create, update, and delete a network interface.

## Example Usage

```terraform
  resource "ibm_pi_network_interface" "network_interface" {
    pi_cloud_instance_id = "<value of the cloud_instance_id>"
    pi_network_id = "network_id"
    pi_name = "network-interface-name"
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

## Argument Reference

Review the argument references that you can specify for your resource.

- `pi_cloud_instance_id` - (Required, Forces new resource, String) The GUID of the service instance associated with an account.
- `pi_instance_id` - (Optional, String) If supplied populated it attaches to the PVMInstanceID, if empty detaches from PVMInstanceID.
- `pi_ip_address` - (Optional,Forces new resource, String) The requested IP address of this network interface.
- `pi_name` - (Optional, String) Name of the network interface.
- `pi_network_id` - (Required, Forces new resource, String) network id.
- `pi_user_tags` - (Optional, Forces new resource, List) The user tags attached to this resource.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute reference after your resource is created.

- `crn` - (String) The network interface's crn.

- `ip_address` - (String) The ip address of this network interface.

- `mac_address` - (String) The mac address of the network interface.

- `name` - (String) Name of the network interface (not unique or indexable).
- `network_interface_id` - (String) The unique identifier of the network interface.
- `network_security_group_id` - (String) ID of the Network Security Group the network interface will be added to.

- `pvm_instance` - (List) The attached pvm-instances to this network interface.
Nested schema for `pvm_instance`:
  - `href` - (String) Link to pvm-instance resource.
  - `instance_id` - (String) The attahed pvm-instance id.

- `status` - (String) The status of the network address group.
