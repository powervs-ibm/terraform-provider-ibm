---
layout: "ibm"
page_title: "IBM : ibm_pi_network_interface"
description: |-
  Get information about pi_network_interface
subcategory: "Power Systems"
---

# ibm_pi_network_interface

Retrieves information about a network interface.

## Example Usage

```terraform
    data "ibm_pi_network_interface" "network_interface" {
        pi_cloud_instance_id = "<value of the cloud_instance_id>"
        pi_network_id = "network_id"
        pi_network_interface_id = "network_interface_id"
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

You can specify the following arguments for this data source.

- `pi_cloud_instance_id` - (Required, String) The GUID of the service instance associated with an account.
- `pi_network_id` - (Required, String) network id.
- `pi_network_interface_id` - (Required, String) network interface id.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute references after your data source is created.

- `id` - The unique identifier of the network interface.
- `crn` - (String) The network interface's crn.

- `ip_address` - (String) The ip address of this network interface.

- `mac_address` - (String) The mac address of the network interface.

- `name` - (String) Name of the network interface (not unique or indexable).

- `pvm_instance` - (List) The attached pvm-instances to this network interface.
   Nested schema for `pvm_instance`:
  - `href` - (String) Link to pvm-instance resource.
  - `pvm_instance_id` - (String) The attahed pvm-instance id.

- `status` - (String) The status of the network address group.
