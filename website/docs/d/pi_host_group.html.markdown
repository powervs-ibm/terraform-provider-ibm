---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: ibm_pi_hostgroup"
description: |-
  Manages a hostgroup in
---

# ibm_pi_hostgroup

Provides a read-only data source to retrieve information about a hostgroup you can use in Power Systems Virtual Server. For more information, about ower Systems Virtual Server hostgroup, see [hostgroups](https://cloud.ibm.com/apidocs/power-cloud#endpoint).

## Example usage

```terraform
data "ibm_pi_hostgroup" "ds_hostgroup" {
    pi_cloud_instance_id    = "<value of the cloud_instance_id>"
    pi_hostgroup_id         = "<value of the hostgroup_id>"
}
```

## Notes

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

You can specify the following arguments for this data source.

- `pi_cloud_instance_id` - (Required, String) The GUID of the service instance associated with an account.

- `hostgroup_id` - (Required, Forces new resource, String) Hostgroup ID.

## Attribute Reference

In addition to all argument reference list, you can access the following attribute references after your data source is created.

- `id` - The unique identifier of the hostgroup.
- `creation_date` - (String) Date/Time of hostgroup creation.

- `hosts` - (List) List of hosts.

- `name` - (String) Name of the hostgroup.

- `primary` - (String) Name of the workspace owning the hostgroup.

- `secondaries` - (List) Names of workspaces the hostgroup has been shared with.
