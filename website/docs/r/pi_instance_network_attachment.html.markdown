---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_instance_network_attachment"
description: |-
  Manages a VPC Virtual Network Interface attachment to an instance in the Power Virtual Server cloud.
---

# ibm_pi_instance_network_attachment

Attach, update, and detach a VPC Virtual Network Interface (VNI) to a Power Systems Virtual Server instance. For more information, about IBM power virtual server cloud, see [getting started with IBM Power Systems Virtual Servers](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-getting-started).

## Example Usage

The following example enables you to attach a VPC Virtual Network Interface to an instance:

```terraform
resource "ibm_pi_instance_network_attachment" "example" {
  pi_cloud_instance_id = "<value of the service instance id>"
  pi_instance_id        = "<value of the instance id>"
  pi_name               = "<name of the network attachment>"
  vpc_virtual_network_interface {
    crn       = "<crn of the VPC Virtual Network Interface>"
    bandwidth = 10
  }
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

The `ibm_pi_instance_network_attachment` provides the following [timeouts](https://www.terraform.io/docs/language/resources/syntax.html) configuration options:

- **create** - (Default 10 minutes) Used for attaching a VPC Virtual Network Interface to an instance.
- **update** - (Default 10 minutes) Used for updating a network attachment.
- **delete** - (Default 10 minutes) Used for detaching a VPC Virtual Network Interface from an instance.

## Argument Reference

Review the argument references that you can specify for your resource.

- `pi_cloud_instance_id` - (Required, Forces new resource, String) The GUID of the service instance associated with an account.
- `pi_instance_id` - (Required, Forces new resource, String) The unique identifier of the instance to attach the network attachment to.
- `pi_name` - (Optional, String) The name of this network attachment. Names must be unique within the PVM instance the network attachment resides in. If unspecified, the name will be a hyphenated list of randomly-selected words.
- `vpc_virtual_network_interface` - (Required, Forces new resource, List) The VPC Virtual Network Interface to attach.

  Nested scheme for `vpc_virtual_network_interface`:
  - `bandwidth` - (Optional, Integer) Bandwidth in Gbps for the network interface. Valid range is 1 to 100. Default is 1.
  - `crn` - (Required, Forces new resource, String) CRN of the VPC Virtual Network Interface to attach.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute reference after your resource is created.

- `id` - (String) The unique identifier of the instance network attachment. The ID is composed of `<pi_cloud_instance_id>/<pi_instance_id>/<network_attachment_id>`.
- `network_attachment_id` - (String) ID of the network attachment.
- `type` - (String) The type of network that is attached.
- `vpc_virtual_network_interface` - (List) The VPC Virtual Network Interface that is attached.

  Nested scheme for `vpc_virtual_network_interface`:
  - `ip_address` - (String) The IP address of the network interface.
  - `mac_address` - (String) The MAC address of the network interface.

## Import

The `ibm_pi_instance_network_attachment` can be imported by using `pi_cloud_instance_id`, `pi_instance_id` and `network_attachment_id`.

### Example

```sh
terraform import ibm_pi_instance_network_attachment.example d7bec597-4726-451f-8a63-e62e6f19c32c/cea6651a-bc0a-4438-9f8a-a0770bbf3ebb/4726d7be-c597-4438-9f8a-cea6651abc0a
```
