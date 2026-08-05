---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_instance_snapshots_v2"
description: |-
  Retrieves all instance snapshots (v2) in the Power Virtual Server cloud.
---

# ibm_pi_instance_snapshots_v2

Retrieve information about all Power Systems Virtual Server instance snapshots (v2). For more information, about Power Virtual Server instance snapshots, see [getting started with IBM Power Systems Virtual Servers](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-getting-started).

## Example Usage

```terraform
data "ibm_pi_instance_snapshots_v2" "ds_snapshots_v2" {
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

- `instance_snapshots` - (List) List of Power Virtual Machine instance snapshots (v2) within the given cloud instance.

  Nested scheme for `instance_snapshots`:
  - `action` - (String) Action performed on the instance snapshot.
  - `aggregated_snapshot_usage` - (Map) Aggregated usage for the snapshot, keyed by storage pool.
  - `completion_date` - (String) Date of snapshot completion.
  - `copy_volumes` - (List) List of copy volumes created as part of the snapshot.

    Nested scheme for `copy_volumes`:
    - `copy_volume_crn` - (String) The CRN of the copy volume.
    - `copy_volume_id` - (String) The ID of the copy volume.
    - `copy_volume_name` - (String) The name of the copy volume.
    - `copy_volume_user_tags` - (List) User tags for the copy volume.
    - `src_volume_crn` - (String) The CRN of the source volume.

  - `crn` - (String) The CRN of this resource.
  - `creation_date` - (String) Date of snapshot creation.
  - `description` - (String) The description of the snapshot.
  - `id` - (String) The unique identifier of the Power Systems Virtual Machine instance snapshot.
  - `job_id` - (String) The ID of the job associated with the snapshot operation.
  - `last_update_date` - (String) Date of last update.
  - `name` - (String) The name of the Power Systems Virtual Machine instance snapshot.
  - `percent_complete` - (Integer) The snapshot completion percentage.
  - `pvm_instance_crn` - (String) The CRN of the PVM instance associated with the snapshot.
  - `pvm_instance_id` - (String) The ID of the PVM instance associated with the snapshot.
  - `remote_peer_snapshot` - (List) Details of the remote peer snapshot.

    Nested scheme for `remote_peer_snapshot`:
    - `completion_date` - (String) Date of remote snapshot completion.
    - `copy_volumes` - (List) List of copy volumes for the remote snapshot. See `copy_volumes` above for nested schema.
    - `crn` - (String) The CRN of the remote peer snapshot.
    - `id` - (String) The ID of the remote peer snapshot.
    - `name` - (String) The name of the remote peer snapshot.
    - `status` - (String) The status of the remote peer snapshot.
    - `status_detail` - (String) Detailed status information for the remote peer snapshot.

  - `status` - (String) The status of the Power Virtual Machine instance snapshot.
  - `status_detail` - (String) Detailed information for the last PVM instance snapshot action.
  - `type` - (String) The type of the snapshot.
  - `user_tags` - (List) List of user tags attached to the resource.
  - `volume_snapshots` - (Map) A map of volume snapshots included in the Power Virtual Machine instance snapshot.
