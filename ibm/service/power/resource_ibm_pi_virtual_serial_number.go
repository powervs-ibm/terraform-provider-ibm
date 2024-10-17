// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPIVirtualSerialNumber() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceCreate,
		ReadContext:   resourceIBMPIInstanceRead,
		UpdateContext: resourceIBMPIInstanceUpdate,
		DeleteContext: resourceIBMPIInstanceDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description: "This is the Power Instance id that is assigned to the account",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			Arg_Description: {
				Description: "Description of virtual serial number.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_Serial: {
				Description: "Virtual serial number.",
				Required:    true,
				Type:        schema.TypeString,
			},

			// Attributes
			Attr_PVMInstanceID: {
				Computed:    true,
				Description: "PVM instance ID virtual serial number is assigned to.",
				Type:        schema.TypeBool,
			},
		},
	}
}

func resourceIBMPIVirtualSerialNumberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	client := instance.NewIBMPIVSNClient(ctx, sess, cloudInstanceID)

	vsnArg := d.Get(Arg_Serial).(string)
	vsn, err := client.Get(vsnArg)
	if err != nil {
		return diag.FromErr(err)
	}

	id := cloudInstanceID + "/" + *vsn.Serial
	d.SetId(id)

	return resourceIBMPICaptureRead(ctx, d, meta)

}

func resourceIBMPIVirtualSerialNumberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	idArr, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := idArr[0]
	serial := idArr[1]

	client := instance.NewIBMPIVSNClient(ctx, sess, cloudInstanceID)
	vsn, err := client.Get(serial)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set(Arg_Description, vsn.Description)
	if vsn.PvmInstanceID != nil {
		d.Set(Attr_PVMInstanceID, vsn.PvmInstanceID)
	}

	return nil
}

func resourceIBMPIVirtualSerialNumberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	idArr, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := idArr[0]
	serial := idArr[1]

	client := instance.NewIBMPIVSNClient(ctx, sess, cloudInstanceID)
	err = client.Delete(serial)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceIBMPIVirtualSerialNumberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange(Arg_VirtualSerialNumber) || d.HasChange(Arg_CloudInstanceID) {
		cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
		client := instance.NewIBMPIVSNClient(ctx, sess, cloudInstanceID)

		vsnArg := d.Get(Arg_Serial).(string)
		vsn, err := client.Get(vsnArg)
		if err != nil {
			return diag.FromErr(err)
		}

		id := cloudInstanceID + "/" + *vsn.Serial
		d.SetId(id)
	} else {
		if d.HasChange(Arg_Description) {
			cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
			client := instance.NewIBMPIVSNClient(ctx, sess, cloudInstanceID)

			newDescription := d.Get(Arg_Description).(string)
			updateBody := &models.UpdateVirtualSerialNumber{
				Description: &newDescription,
			}

			vsnArg := d.Get(Arg_Serial).(string)

			_, err = client.Update(vsnArg, updateBody)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceIBMPIVirtualSerialNumberRead(ctx, d, meta)
}
