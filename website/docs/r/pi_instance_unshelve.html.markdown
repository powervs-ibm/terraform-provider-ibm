---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_instance_unshelve"
description: |-
  Unshelves a previously shelved PVM instance.
---

# ibm_pi_instance_unshelve

Unshelves a previously shelved [Power Systems Virtual Server instance](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-creating-power-virtual-server), shelved with the [`ibm_pi_instance_shelve`](/docs/providers/ibm/r/pi_instance_shelve.html) resource, to restore it to active state.

## Example Usage

```terraform
resource "ibm_pi_instance_unshelve" "example" {
  pi_cloud_instance_id = "d7bec597-4726-451f-8a63-e62e6f19c32c"
  pi_instance_id       = "cea6651a-bc0a-4438-9f8a-a0770b112ebb"
  pi_memory            = 2
  pi_processors        = 1
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

The `ibm_pi_instance_unshelve` provides the following [timeouts](https://www.terraform.io/docs/language/resources/syntax.html) configuration options:

- **create** - (Default 15 minutes) Used for unshelving the instance.

## Argument References

Review the argument references that you can specify for your resource.

- `pi_cloud_instance_id` - (Required, String, ForceNew) The GUID of the service instance associated with an account.
- `pi_deployment_target` - (Optional, List, ForceNew) The dedicated host or host group where the instance is to be placed. Max items: 1.

  Nested scheme for `pi_deployment_target`:
  - `id` - (Required, String) The uuid of the host group or host.
  - `type` - (Required, String) The deployment target type. Supported values are `host` and `hostGroup`.
- `pi_instance_id` - (Required, String, ForceNew) The ID of the pvm instance to unshelve.
- `pi_memory` - (Optional, Float, ForceNew) Amount of memory to allocate (in GiB). Conflicts with `pi_sap_profile_id`.
- `pi_preferred_processor_compatibility_mode` - (Optional, String, ForceNew) Preferred processor compatibility mode.
- `pi_processors` - (Optional, Float, ForceNew) Number of processors to allocate. Conflicts with `pi_sap_profile_id`.
- `pi_sap_profile_id` - (Optional, String, ForceNew) SAP profile ID to use when unshelving the instance. Only valid for instances already using an SAP profile. Conflicts with `pi_processors` and `pi_memory`.
- `pi_shared_processor_pool` - (Optional, String, ForceNew) The shared processor pool on which the instance is to be placed.
- `pi_sys_type` - (Optional, String, ForceNew) System type used to host the instance.
- `pi_virtual_cores_assigned` - (Optional, Integer, ForceNew) Number of virtual cores to allocate.
- `pi_virtual_serial_number` - (Optional, List, ForceNew) Virtual Serial Number information to assign when unshelving the instance. Max items: 1. Conflicts with `pi_sap_profile_id`.

  Nested scheme for `pi_virtual_serial_number`:
  - `software_tier` - (Required, String) Software tier for virtual serial number. Allowed values are: ["P05", "P10", "P20", "P30"].

## Attribute Reference

In addition to all argument reference list, you can access the following attribute reference after your resource is created.

- `id` - (String) The unique identifier of the instance. The ID is composed of `<pi_cloud_instance_id>/<pi_instance_id>`.
- `status` - (String) The status of the instance.

## Import

The `ibm_pi_instance_unshelve` can be imported using `pi_cloud_instance_id` and `pi_instance_id`.

### Example

```bash
terraform import ibm_pi_instance_unshelve.example d7bec597-4726-451f-8a63-e62e6f19c32c/cea6651a-bc0a-4438-9f8a-a0770b112ebb
```
