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

- `pi_cloud_instance_id` - (Required, String) The GUID of the service instance associated with an account. 
- `pi_description` - (Optional, String) Desired description for virtual serial number.
- `pi_serial` - (Required, String) Virtual serial number.


## Attribute reference
 In addition to all argument reference list, you can access the following attribute reference after your resource is created.

- `id` - (String) The unique identifier of the virtual serial number. Composed of `<cloud instance id>/<virtual serial number>`
- `pvm_instance_id` - (String) ID of the PVM instance the virtual serial number is assigned to.

## Import

The `ibm_virtual_serial_number` resource can be imported by using `power_instance_id` and `serial`.

**Example**

```bash
$ terraform import ibm_pi_virtual_serial_number.example power_instance_id/virtual_serial_number
```
