---

subcategory: "Power Systems"
layout: "ibm"
page_title: "IBM: pi_virtual_serial_number"
description: |-
  Manages an EXISTING virtual serial number in IBM Power
---

# ibm_pi_virtual_serial_number
Get, update or delete an existing virtual serial number.

## Example usage
The following example enables you to create a shared processor pool placement group with a group policy of affinity:

```terraform
resource "ibm_pi_virtual_serial_number" "testacc_virtual_serial_number" {
  pi_serial   = "<existing virtual serial number>"
  pi_cloud_instance_id      = "<value of the cloud_instance_id>"
  pi_description = "<desired description for virtual serial number>"
}
```

**Note**
* Please find [supported Regions](https://cloud.ibm.com/apidocs/power-cloud#endpoint) for endpoints.
* If a Power cloud instance is provisioned at `lon04`, The provider level attributes should be as follows:
  * `region` - `lon`
  * `zone` - `lon04`
  
  Example usage:

  ```terraform
    provider "ibm" {
      region    =   "lon"
      zone      =   "lon04"
    }
  ```

**Note**
* This resource CANNOT create a virtual serial number. It can only be used to manage an existing virtual serial number. To create a virtual serial number and assign it, please use the `pi_instance` resource.

## Timeouts

ibm_pi_virtual_serial_number provides the following [timeouts](https://www.terraform.io/docs/language/resources/syntax.html) configuration options:

- **create** - (Default 10 minutes) Used for getting an existing virtual serial number.
- **update** - (Default 10 minutes) Used for updating a virtual serial number.
- **delete** - (Default 10 minutes) Used for deleting a reserved virtual serial number.

## Argument reference
Review the argument references that you can specify for your resource. 

- `pi_assign_virtual_serial_number` - (Optional, List) Information of virtual serial number to assign to a power instance. Required with `pi_instance_id`.
  - `description` - (String, Optional) Description of virtual serial number.
  - `serial` - (String, Required) Provide an existing reserved Virtual Serial Number or specify 'auto-assign' for auto generated Virtual Serial Number.
      
    ~> **Note** When set to "auto-assign", changes to `serial` outside of terraform will not be detected. In addition, if a new generated virtual serial number is needed,
    the old serial must be deleted before a new one is generated.
- `pi_cloud_instance_id` - (Required, String) The GUID of the service instance associated with an account. 
- `pi_description` - (Optional, String) Desired description for virtual serial number.
- `pi_instance_id` - (Optional, String) Power instance ID to assign created or existing virtual serial number to. Conflicts with `pi_serial`
- `pi_retain_virtual_serial_number` - (Optional, Boolean) Indicates whether to reserve or delete virtual serial number when detached from power instance during delete. Required with `pi_instance_id`
- `pi_serial` - (Required, String) Virtual serial number of existing serial. Conflicts with `pi_instance_id`


## Attribute reference
 In addition to all argument reference list, you can access the following attribute reference after your resource is created.

- `id` - (String) The unique identifier of the virtual serial number. Composed of `<cloud instance id>/<virtual serial number>`

## Import

The `ibm_virtual_serial_number` resource can be imported by using `pi_cloud_instance_id` and `serial`.

**Example**

```bash
$ terraform import ibm_pi_virtual_serial_number.example d7bec597-4726-451f-8a63-e62e6f19c32c/VS0762Y
```
