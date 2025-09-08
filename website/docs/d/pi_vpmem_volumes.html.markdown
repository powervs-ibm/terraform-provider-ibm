---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM : ibm_pi_vpmem_volumes"
description: |-
  Get information about pi_vpmem_volumes
---

# ibm_pi_vpmem_volumes

Retrieves information about vPMEM volumes in Power Systems Virtual Server cloud. For more information, about managin a volume, see [moving data to the cloud](https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-moving-data-to-the-cloud).

## Example Usage

```terraform
data "ibm_pi_vpmem_volumes" "vpmem_volumes" {
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

- `id` - The unique identifier of the vpmem_volumes.
- `volumes` - (List) List of vPMEM volumes.
    Nested schema for `volumes`:
  - `created_at` - (String) Time when the volume was created.
  - `crn` - (String) The CRN for this resource.
  - `href` - (String) Link to vPMEM volume resource.
  - `name` - (String) Volume Name.
  - `pvm_instance_id` - (String) PVM Instance ID which the volume is attached to.
  - `size` - (Float) Volume Size (GB).
  - `status` - (String) Status of the volume.
  - `user_tags` - (List) List of user tags.
  - `volume_id` - (String) Volume ID.
