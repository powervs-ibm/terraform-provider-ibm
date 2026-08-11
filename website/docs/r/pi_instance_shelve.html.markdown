---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_instance_shelve"
description: |-
  Shelves a PVM instance.
---

# ibm_pi_instance_shelve

Shelves a [Power Systems Virtual Server instance](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-creating-power-virtual-server) to reduce resource consumption. The instance configuration is preserved and can be restored later with the [`ibm_pi_instance_unshelve`](/docs/providers/ibm/r/pi_instance_unshelve.html) resource.

## Example Usage

```terraform
resource "ibm_pi_instance_shelve" "example" {
  pi_cloud_instance_id = "d7bec597-4726-451f-8a63-e62e6f19c32c"
  pi_instance_id       = "cea6651a-bc0a-4438-9f8a-a0770b112ebb"
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

The `ibm_pi_instance_shelve` provides the following [timeouts](https://www.terraform.io/docs/language/resources/syntax.html) configuration options:

- **create** - (Default 15 minutes) Used for shelving the instance.

## Argument References

Review the argument references that you can specify for your resource.

- `pi_cloud_instance_id` - (Required, String, ForceNew) The GUID of the service instance associated with an account.
- `pi_instance_id` - (Required, String, ForceNew) The ID of the pvm instance to shelve.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute reference after your resource is created.

- `id` - (String) The unique identifier of the instance. The ID is composed of `<pi_cloud_instance_id>/<pi_instance_id>`.
- `status` - (String) The status of the instance.

## Import

The `ibm_pi_instance_shelve` can be imported using `pi_cloud_instance_id` and `pi_instance_id`.

### Example

```bash
terraform import ibm_pi_instance_shelve.example d7bec597-4726-451f-8a63-e62e6f19c32c/cea6651a-bc0a-4438-9f8a-a0770b112ebb
```
