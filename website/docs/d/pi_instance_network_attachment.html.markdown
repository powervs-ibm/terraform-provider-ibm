---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_instance_network_attachment"
description: |-
  Retrieve information about a VPC Virtual Network Interface attachment on a Power Systems Virtual Server instance.
---

# ibm_pi_instance_network_attachment

Retrieve information about a specific network attachment on a Power Systems Virtual Server instance. For more information about Power Virtual Server instances, see [getting started with IBM Power Systems Virtual Servers](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-getting-started).

## Example Usage

```terraform
data "ibm_pi_instance_network_attachment" "ds_instance_network_attachment" {
  pi_cloud_instance_id     = "49fba6c9-23f8-40bc-9899-aca322ee7d5b"
  pi_instance_id           = "cea6651a-bc0a-4438-9f8a-a0770b112ebb"
  pi_network_attachment_id = "4726d7be-c597-4438-9f8a-cea6651abc0a"
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
- `pi_instance_id` - (Required, String) The unique identifier of the instance.
- `pi_network_attachment_id` - (Required, String) The unique identifier of the network attachment.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute references after your data source is created.

- `id` - (String) The unique identifier of the data source in the form <pi_cloud_instance_id>/<pi_instance_id>/<pi_network_attachment_id>.
- `name` - (String) The name of the network attachment.
- `type` - (String) The type of network that is attached.
- `vpc_virtual_network_interface` - (List) The VPC Virtual Network Interface that is attached.

  Nested scheme for `vpc_virtual_network_interface`:
  - `bandwidth` - (Integer) Bandwidth in Gbps of the network interface.
  - `crn` - (String) CRN of the VPC Virtual Network Interface.
  - `ip_address` - (String) The IP address of the network interface.
  - `mac_address` - (String) The MAC address of the network interface.
  - `name` - (String) The name for the network interface.
