---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_async_job"
description: |-
  Retrieves an asynchronous job in the Power Virtual Server cloud.
---

# ibm_pi_async_job

Retrieve information about a Power Systems Virtual Server asynchronous job. For more information, about Power Virtual Server, see [getting started with IBM Power Systems Virtual Servers](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-getting-started).

## Example Usage

```terraform
data "ibm_pi_async_job" "ds_async_job" {
  pi_cloud_instance_id = "49fba6c9-23f8-40bc-9899-aca322ee7d5b"
  pi_async_job_id      = "8fba9c34-12a4-5678-bcde-9f0e1234ab56"
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

- `pi_async_job_id` - (Required, String) The unique identifier of the asynchronous job.
- `pi_cloud_instance_id` - (Required, String) The GUID of the service instance associated with an account.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute references after your data source is created.

- `action` - (String) The action being performed in the job.
- `child_jobs` - (List) Array of child jobs.

  Nested scheme for `child_jobs`:
  - `action` - (String) Action of the child job.
  - `id` - (String) ID of the child job.
  - `progress_percent` - (Integer) Progress percentage of the child job.
  - `resource_id` - (String) ID of the resource being acted upon.
  - `resource_type` - (String) Type of resource being acted upon.
  - `status` - (String) Status of the child job.

- `completion_date` - (String) Date the job completed.
- `creation_date` - (String) Date the job was created.
- `error_message` - (String) Detailed information of error encountered during job processing.
- `last_update_date` - (String) Date the job was last updated.
- `parent_async_job_id` - (String) ID of the parent async job.
- `progress_percent` - (Integer) Percentage of the job that has completed.
- `resource_id` - (String) ID of the resource being acted upon in the job.
- `resource_type` - (String) Type of resource being acted upon in the job.
- `status` - (String) Status of the job.
