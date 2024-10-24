// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"strings"
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
		CreateContext: resourceIBMPIVirtualSerialNumberCreate,
		ReadContext:   resourceIBMPIVirtualSerialNumberRead,
		UpdateContext: resourceIBMPIVirtualSerialNumberUpdate,
		DeleteContext: resourceIBMPIVirtualSerialNumberUpdate,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_AssignVirtualSerialNumber: {
				Description: "Virtual Serial Number information",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Description: {
							Description: "Description of the Virtual Serial Number",
							Optional:    true,
							Type:        schema.TypeString,
						},
						Attr_Serial: {
							Description:      "Provide an existing reserved Virtual Serial Number or specify 'auto-assign' for auto generated Virtual Serial Number.",
							DiffSuppressFunc: supressVSNDiffAutoAssign,
							ForceNew:         true,
							Required:         true,
							Type:             schema.TypeString,
						},
					},
				},
				MaxItems:     1,
				Optional:     true,
				RequiredWith: []string{Arg_PVMInstanceId},
				Type:         schema.TypeList,
			},
			Arg_CloudInstanceID: {
				Description: "This is the Power Instance id that is assigned to the account",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			Arg_Description: {
				ConflictsWith: []string{Arg_AssignVirtualSerialNumber},
				Computed:      true,
				Description:   "Description of virtual serial number.",
				Optional:      true,
				Type:          schema.TypeString,
			},
			Arg_PVMInstanceId: {
				Computed:      true,
				ConflictsWith: []string{Arg_Serial},
				Description:   "PVM Instance to attach VSN to",
				Optional:      true,
				Type:          schema.TypeString,
			},
			Arg_RetainVirtualSerialNumber: {
				ConflictsWith: []string{Arg_Serial},
				Description:   "Indicates whether to retain virtual serial number after unassigning from PVM instance.",
				Optional:      true,
				RequiredWith:  []string{Arg_PVMInstanceId},
				Type:          schema.TypeBool,
			},
			Arg_Serial: {
				ConflictsWith:    []string{Arg_AssignVirtualSerialNumber},
				Computed:         true,
				Description:      "Virtual serial number.",
				DiffSuppressFunc: flex.ApplyOnce,
				Optional:         true,
				Type:             schema.TypeString,
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

	serialString := ""
	if serial, ok := d.GetOk(Arg_Serial); ok {
		vsnArg := serial.(string)
		vsn, err := client.Get(vsnArg)
		if err != nil {
			return diag.FromErr(err)
		}
		serialString = *vsn.Serial
	}

	if pvmInstanceId, ok := d.GetOk(Arg_PVMInstanceId); ok {
		pvmInstanceIdArg := pvmInstanceId.(string)
		instanceClient := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
		ins, err := instanceClient.Get(pvmInstanceIdArg)
		if err != nil {
			diag.FromErr(err)
		}
		status := *ins.Status
		restartInstance := false
		if strings.ToLower(status) != State_Shutoff {
			err = stopLparForResourceChange(ctx, instanceClient, pvmInstanceIdArg, d)
			if err != nil {
				return diag.FromErr(err)
			}
			restartInstance = true
		}
		serialNumber := d.Get(Arg_AssignVirtualSerialNumber + ".0." + Attr_Serial).(string)
		addBody := &models.AddServerVirtualSerialNumber{
			Serial: &serialNumber,
		}
		if desc, ok := d.GetOk(Arg_AssignVirtualSerialNumber + ".0." + Attr_Description); ok {
			description := desc.(string)
			addBody.Description = description
		}
		err = client.PVMInstanceAttachVSN(pvmInstanceIdArg, addBody)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = isWaitForPIInstanceStopped(ctx, instanceClient, pvmInstanceIdArg, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}

		if restartInstance {
			err = startLparAfterResourceChange(ctx, instanceClient, pvmInstanceIdArg, d)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		vsns, err := client.GetAll(&pvmInstanceIdArg)
		if err != nil {
			return diag.FromErr(err)
		}
		serialString = *vsns[0].Serial
	}

	id := cloudInstanceID + "/" + serialString
	d.SetId(id)

	return resourceIBMPIVirtualSerialNumberRead(ctx, d, meta)
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

	client := instance.NewIBMPIVSNClient(ctx, sess, cloudInstanceID)

	if serial, ok := d.GetOk(Arg_Serial); ok {
		serialNumber := serial.(string)
		vsn, err := client.Get(serialNumber)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set(Arg_Description, vsn.Description)
		if vsn.PvmInstanceID != nil {
			d.Set(Arg_PVMInstanceId, vsn.PvmInstanceID)
		}
		d.Set(Arg_Serial, vsn.Serial)
	}

	if pvmInstanceId, ok := d.GetOk(Arg_PVMInstanceId); ok {
		pvmInstanceIdArg := pvmInstanceId.(string)
		vsns, err := client.GetAll(&pvmInstanceIdArg)
		if err != nil {
			return diag.FromErr(err)
		}
		if len(vsns) < 1 {
			return diag.Errorf("get of serial numbers assigned to %s found 0 items", pvmInstanceIdArg)
		}
		vsn := vsns[0]
		d.Set(Arg_Description, vsn.Description)
		d.Set(Arg_Serial, vsn.Serial)
		d.Set(Arg_AssignVirtualSerialNumber, flattenVirtualSerialNumberToListSerialType(vsn))
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
	client := instance.NewIBMPIVSNClient(ctx, sess, cloudInstanceID)

	if v, ok := d.GetOk(Arg_Serial); ok {
		serialNumber := v.(string)
		err = client.Delete(serialNumber)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk(Arg_PVMInstanceId); ok {
		pvmInstanceId := v.(string)
		retainVSN := d.Get(Arg_RetainVirtualSerialNumber).(bool)
		deleteBody := &models.DeleteServerVirtualSerialNumber{
			RetainVSN: retainVSN,
		}
		err = client.PVMInstanceDeleteVSN(pvmInstanceId, deleteBody)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")

	return nil
}

func resourceIBMPIVirtualSerialNumberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	client := instance.NewIBMPIVSNClient(ctx, sess, cloudInstanceID)

	if _, ok := d.GetOk(Arg_AssignVirtualSerialNumber); !ok && d.HasChange(Arg_Description) {
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

	if d.HasChange(Arg_PVMInstanceId) {
		oldId, newId := d.GetChange(Arg_PVMInstanceId)
		oldIdString, newIdString := oldId.(string), newId.(string)
		instanceClient := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
		retainVSN := d.Get(Arg_RetainVirtualSerialNumber).(bool)

		// Old instance
		ins, err := instanceClient.Get(oldIdString)
		if err != nil {
			diag.FromErr(err)
		}
		status := *ins.Status
		restartInstance := false
		if strings.ToLower(status) != State_Shutoff {
			err = stopLparForResourceChange(ctx, instanceClient, oldIdString, d)
			if err != nil {
				return diag.FromErr(err)
			}
			restartInstance = true
		}

		detachBody := &models.DeleteServerVirtualSerialNumber{
			RetainVSN: retainVSN,
		}
		err = client.PVMInstanceDeleteVSN(oldIdString, detachBody)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = isWaitForPIInstanceStopped(ctx, instanceClient, oldIdString, d.Timeout(schema.TimeoutUpdate))

		if restartInstance {
			err = startLparAfterResourceChange(ctx, instanceClient, oldIdString, d)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// New instance
		ins, err = instanceClient.Get(newIdString)
		if err != nil {
			diag.FromErr(err)
		}
		status = *ins.Status
		restartInstance = false
		if strings.ToLower(status) != State_Shutoff {
			err = stopLparForResourceChange(ctx, instanceClient, newIdString, d)
			if err != nil {
				return diag.FromErr(err)
			}
			restartInstance = true
		}

		serial := d.Get(Arg_AssignVirtualSerialNumber + ".0." + Attr_Serial).(string)
		addBody := &models.AddServerVirtualSerialNumber{
			Serial: &serial,
		}
		if v, ok := d.GetOk(Arg_AssignVirtualSerialNumber + ".0." + Attr_Description); ok {
			description := v.(string)
			addBody.Description = description
		}
		err = client.PVMInstanceAttachVSN(newIdString, addBody)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = isWaitForPIInstanceStopped(ctx, instanceClient, newIdString, d.Timeout(schema.TimeoutUpdate))

		if restartInstance {
			err = startLparAfterResourceChange(ctx, instanceClient, newIdString, d)
			if err != nil {
				return diag.FromErr(err)
			}
		}

	}

	if !d.HasChange(Arg_PVMInstanceId) && d.HasChange(Arg_AssignVirtualSerialNumber+".0."+Attr_Description) {
		pvmInstanceId := d.Get(Arg_PVMInstanceId).(string)
		description := d.Get(Arg_AssignVirtualSerialNumber + ".0." + Attr_Description).(string)
		updateBody := &models.UpdateServerVirtualSerialNumber{
			Description: &description,
		}
		_, err = client.PVMInstanceUpdateVSN(pvmInstanceId, updateBody)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIBMPIVirtualSerialNumberRead(ctx, d, meta)
}

func flattenVirtualSerialNumberToListSerialType(vsn *models.VirtualSerialNumber) []map[string]interface{} {
	v := make([]map[string]interface{}, 1)
	v[0] = map[string]interface{}{
		Attr_Description: vsn.Description,
		Attr_Serial:      vsn.Serial,
	}
	return v
}
