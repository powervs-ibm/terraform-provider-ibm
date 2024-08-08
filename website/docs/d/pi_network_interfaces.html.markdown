---
layout: "ibm"
page_title: "IBM : ibm_pi_network_interfaces"
description: |-
  Get information about pi_network_interfaces
subcategory: "Power Systems"
---

# ibm_pi_network_interfaces

Retrieve information about network interfaces.

## Example Usage

```terraform
    data "ibm_pi_network_interfaces" "network_interfaces" {
        pi_cloud_instance_id = "<value of the cloud_instance_id>"
        pi_network_id = "network_id"
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

Review the argument references that you can specify for your data source.

- `pi_cloud_instance_id` - (Required, String) The GUID of the service instance associated with an account.

- `pi_network_id` - (Required, String) network id.

## Attribute Reference

In addition to all argument reference listed, you can access the following attribute references after your data source is created.

- `id` - The unique identifier of the pi_network_interfaces.
- `interfaces` - (List) network Interfaces.
  Nested schema for `interfaces`:
  - `crn` - (String) The network interface's crn.
  - `id` - (String) The unique network interface id.
  - `ip_address` - (String) The ip address of this network interface.
  - `mac_address` - (String) The mac address of the network interface.
  - `name` - (String) Name of the network interface (not unique or indexable).
  - `pvm_instance` - (List) The attached pvm-instances to this network interface.
    Nested schema for `pvm_instance`:
    - `href` - (String) Link to pvm-instance resource.
    - `pvm_instance_id` - (String) The attahed pvm-instance id.
  - `status` - (String) The status of the network address group.
