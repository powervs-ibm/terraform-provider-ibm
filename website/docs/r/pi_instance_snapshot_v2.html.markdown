---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_instance_snapshot_v2"
description: |-
  Manages instance snapshots (v2) in the Power Virtual Server cloud.
---

# ibm_pi_instance_snapshot_v2

Manages instance snapshots (v2) in the Power Virtual Server Cloud. For more information, about snapshots in the Power Virtual Server, see [snapshots for PVM Instance](https://cloud.ibm.com/apidocs/power-cloud#pcloud-v2-pvminstances-snapshots-post).

## Example Usage

The following example enables you to create a snapshot:

```terraform
resource "ibm_pi_instance_snapshot_v2" "testacc_snapshot_v2" {
  pi_cloud_instance_id = "<value of the cloud_instance_id>"
  pi_pvm_instance_id   = "<value of the pvm_instance_id>"
  pi_snapshot_name     = "test-snapshot-v2"
  pi_description       = "Testing snapshot v2 for instance"
  pi_snapshot_type     = "in-place"
  pi_volume_ids        = ["volumeid1", "volumeid2"]
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

The `ibm_pi_instance_snapshot_v2` provides the following [Timeouts](https://www.terraform.io/docs/language/resources/syntax.html) configuration options:

- **create** - (Default 60 minutes) Used for Creating snapshot.
- **update** - (Default 60 minutes) Used for Updating snapshot.
- **delete** - (Default 10 minutes) Used for Deleting snapshot.

## Argument Reference

Review the argument references that you can specify for your resource.

- `pi_add_remote_snapshot` - (Optional, String, Forces new resource) Indicates if a remote snapshot should be added. Allowed values are `no` and `regional`.
- `pi_cloud_instance_id` - (Required, String, Forces new resource) The GUID of the service instance associated with an account.
- `pi_description` - (Optional, String) Description of the PVM instance snapshot.
- `pi_pvm_instance_id` - (Required, String, Forces new resource) The ID of the PVM instance to snapshot.
- `pi_remote_snapshot` - (Optional, List, Forces new resource) Remote snapshot configuration. Maximum items: 1.

  Nested scheme for `pi_remote_snapshot`:
  - `description` - (Optional, String) Description of the remote snapshot.
  - `name` - (Required, String) The name of the remote snapshot.
  - `propagate_user_tags` - (Optional, Bool) Indicates if the user tags should be propagated to the remote snapshot.
  - `user_tags` - (Optional, List) User tags for the remote snapshot.
  - `workspace_crn` - (Optional, String) The CRN of the target workspace for the remote snapshot.

- `pi_snapshot_name` - (Required, String) The unique name of the snapshot.
- `pi_snapshot_type` - (Optional, String, Forces new resource) The type of snapshot. Allowed values are `in-place` and `zonal`.
- `pi_user_tags` - (Optional, List) The user tags attached to this resource.
- `pi_volume_ids` - (Optional, List) A list of volume IDs of the instance that will be part of the snapshot. If none are provided, then all the volumes of the instance will be part of the snapshot.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute reference after your resource is created.

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
- `creation_date` - (String) Creation date of the snapshot.
- `id` - (String) The unique identifier of the snapshot. The ID is composed of `<pi_cloud_instance_id>/<snapshot_id>`.
- `job_id` - (String) The ID of the job associated with the snapshot operation.
- `last_update_date` - (String) The last updated date of the snapshot.
- `percent_complete` - (Integer) The snapshot completion percentage.
- `pvm_instance_crn` - (String) The CRN of the PVM instance associated with the snapshot.
- `remote_peer_snapshot` - (List) Details of the remote peer snapshot.

  Nested scheme for `remote_peer_snapshot`:
  - `completion_date` - (String) Date of remote snapshot completion.
  - `copy_volumes` - (List) List of copy volumes for the remote snapshot. See `copy_volumes` above for nested schema.
  - `crn` - (String) The CRN of the remote peer snapshot.
  - `id` - (String) The ID of the remote peer snapshot.
  - `name` - (String) The name of the remote peer snapshot.
  - `status` - (String) The status of the remote peer snapshot.
  - `status_detail` - (String) Detailed status information for the remote peer snapshot.

- `snapshot_id` - (String) ID of the PVM instance snapshot.
- `status` - (String) Status of the PVM instance snapshot.
- `status_detail` - (String) Detailed information for the last PVM instance snapshot action.
- `type` - (String) The type of the snapshot.
- `volume_snapshots` - (Map) A map of volume snapshots included in the PVM instance snapshot.

## Import

The `ibm_pi_instance_snapshot_v2` resource can be imported by using `pi_cloud_instance_id` and `snapshot_id`.

### Example

```bash
terraform import ibm_pi_instance_snapshot_v2.example d7bec597-4726-451f-8a63-e62e6f19c32c/cea6651a-bc0a-4438-9f8a-a0770bbf3ebb
```
