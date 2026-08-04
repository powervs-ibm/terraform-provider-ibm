---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_snapshot_recovery_location"
description: |-
  Retrieves snapshot recovery location information in the Power Virtual Server cloud.
---

# ibm_pi_snapshot_recovery_location

Retrieve snapshot recovery location information for a Power Systems Virtual Server workspace. For more information, about Power Virtual Server snapshot recovery, see [getting started with IBM Power Systems Virtual Servers](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-getting-started).

## Example Usage

```terraform
data "ibm_pi_snapshot_recovery_location" "ds_snapshot_recovery_location" {
  pi_cloud_instance_id = "49fba6c9-23f8-40bc-9899-aca322ee7d5b"
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

## Attribute Reference

In addition to all argument reference list, you can access the following attribute references after your data source is created.

- `location` - (String) The region zone location of the snapshot recovery site.
- `zonal_snapshot_pools` - (List) List of zonal snapshot pools available at this location.

  Nested scheme for `zonal_snapshot_pools`:
  - `regional_snapshot_locations` - (List) List of regional snapshot locations available from this pool.
  - `storage_pool` - (String) The storage pool name.
