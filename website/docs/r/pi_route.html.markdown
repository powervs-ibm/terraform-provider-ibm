---
subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_route"
description: |-
  Manages a route in a routing table in the Power Virtual Server cloud.
---

# ibm_pi_route

Create, update or delete a route.

## Example usage

The following example enables you to create a route:

```terraform
resource "ibm_pi_route" "route" {
	pi_cloud_instance_id = "<cloud-instance-id>"
	pi_name              = "test-route"
	pi_next_hop          = "<next-hop-ip>"
	pi_destination       = "<destination-cidr>"
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

ibm_pi_placement_group provides the following [timeouts](https://www.terraform.io/docs/language/resources/syntax.html) configuration options:

- **create** - (Default 60 minutes) Used for creating a placement group.
- **delete** - (Default 60 minutes) Used for deleting a placement group.

## Argument reference

Review the argument references that you can specify for your resource.

- `pi_cloud_instance_id` - (Required, String) The GUID of the service instance associated with an account.
- `pi_placement_group_name`  - (Required, String) The name of the placement group.
- `pi_placement_group_policy` - (Required, String) The value of the group's affinity policy. Valid values are `affinity` and `anti-affinity`.

## Attribute reference

 In addition to all argument reference list, you can access the following attribute reference after your resource is created.

- `id` - (String) The unique identifier of the placement group.
- `members` - (List of strings) The list of server instances IDs that are members of the placement group.
- `placement_group_id` - (String) The placement group ID.

## Import

The `ibm_pi_placement_group` resource can be imported by using `power_instance_id` and `placement_group_id`.

### Example

```bash
terraform import ibm_pi_placement_group.example d7bec597-4726-451f-8a63-e62e6f19c32c/b17a2b7f-77ab-491c-811e-495f8d4c8947
```
