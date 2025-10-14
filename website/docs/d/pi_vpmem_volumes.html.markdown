---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM : ibm_pi_vpmem_volumes"
description: |-
  Get information about vPMEM volumes
---

# ibm_pi_vpmem_volumes

Retrieves information about vPMEM volumes in Power Systems Virtual Server cloud. For more information, about managin a volume, see [moving data to the cloud](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-moving-data-to-the-cloud).

## Example Usage

```terraform
data "ibm_pi_vpmem_volumes" "vpmem_volumes" {
    pi_cloud_instance_id = "a1b2c3d4-e5f6-7g8h-9i0j-1k2l3m4n5o6p"
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

- `id` - (String) The unique identifier of the vpmem_volumes.
- `volumes` - (List) List of vPMEM volumes.
    Nested schema for `volumes`:
  - `creation_date` - (String) The date and time when the volume was created.
  - `crn` - (String) The CRN for this resource.
  - `error_code` - (String) Error code for the vPMEM volume.
  - `href` - (String) Link to vPMEM volume resource.
  - `name` - (String) Volume Name.
  - `pvm_instance_id` - (String) PVM Instance ID which the volume is attached to.
  - `reason` - (String) Reason for error.
  - `size` - (Float) Volume Size (GB).
  - `status` - (String) Status of the volume.
  - `updated_date` - (String) The date and time when the volume was updated.
  - `user_tags` - (List) List of user tags.
  - `volume_id` - (String) Volume ID.
