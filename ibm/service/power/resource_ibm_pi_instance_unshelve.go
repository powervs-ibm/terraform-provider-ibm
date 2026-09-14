// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPIInstanceUnshelve() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceUnshelveCreate,
		ReadContext:   resourceIBMPIInstanceUnshelveRead,
		DeleteContext: resourceIBMPIInstanceUnshelveDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "PI Cloud instance id",
			},
			Arg_InstanceID: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "PVM instance ID",
			},
			Arg_DeploymentTarget: {
				Description: "The dedicated host or host group where the instance is to be placed.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_ID: {
							Description: "The uuid of the host group or host.",
							Required:    true,
							Type:        schema.TypeString,
						},
						Attr_Type: {
							Description:  "The deployment target type. Supported values are `host` and `hostGroup`.",
							Required:     true,
							Type:         schema.TypeString,
							ValidateFunc: validate.ValidateAllowedStringValues([]string{Host, HostGroup}),
						},
					},
				},
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				Type:     schema.TypeSet,
			},
			Arg_Memory: {
				ConflictsWith: []string{Arg_SAPProfileID},
				Description:   "Amount of memory to allocate (in GiB).",
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeFloat,
			},
			Arg_PreferredProcessorCompatibilityMode: {
				Description: "Preferred processor compatibility mode.",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_Processors: {
				ConflictsWith: []string{Arg_SAPProfileID},
				Description:   "Number of processors to allocate.",
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeFloat,
			},
			Arg_SAPProfileID: {
				ConflictsWith: []string{Arg_Processors, Arg_Memory},
				Description:   "SAP profile ID to use when unshelving the instance. Only valid for instances already using an SAP profile.",
				ForceNew:      true,
				Optional:      true,
				Type:          schema.TypeString,
			},
			Arg_SharedProcessorPool: {
				Description: "The shared processor pool on which the instance is to be placed.",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_SysType: {
				Description: "System type used to host the instance.",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_VirtualCoresAssigned: {
				Description: "Number of virtual cores to allocate.",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			Arg_VirtualSerialNumber: {
				ConflictsWith: []string{Arg_SAPProfileID},
				Description:   "Virtual Serial Number information to assign when unshelving the instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_SoftwareTier: {
							Description:  "Software tier. Enum: [\"P05\", \"P10\", \"P20\", \"P30\"].",
							Required:     true,
							Type:         schema.TypeString,
							ValidateFunc: validate.ValidateAllowedStringValues([]string{"P05", "P10", "P20", "P30"}),
						},
					},
				},
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				Type:     schema.TypeList,
			},
			// Attributes
			Attr_Status: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the PVM instance",
			},
		},
	}
}

func resourceIBMPIInstanceUnshelveCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_unshelve", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	id := d.Get(Arg_InstanceID).(string)

	body := &models.PVMInstanceUnshelve{}
	if v, ok := d.GetOk(Arg_Memory); ok {
		memory := v.(float64)
		body.Memory = &memory
	}
	if v, ok := d.GetOk(Arg_Processors); ok {
		processors := v.(float64)
		body.Processors = &processors
	}
	if v, ok := d.GetOk(Arg_PreferredProcessorCompatibilityMode); ok {
		ppcm := v.(string)
		body.PreferredProcessorCompatibilityMode = &ppcm
	}
	if v, ok := d.GetOk(Arg_SAPProfileID); ok {
		sapProfileID := v.(string)
		body.SapProfileID = &sapProfileID
	}
	if v, ok := d.GetOk(Arg_SharedProcessorPool); ok {
		spp := v.(string)
		body.SharedProcessorPool = &spp
	}
	if v, ok := d.GetOk(Arg_SysType); ok {
		sysType := v.(string)
		body.SysType = &sysType
	}
	if v, ok := d.GetOk(Arg_VirtualCoresAssigned); ok {
		assigned := int64(v.(int))
		body.VirtualCores = &models.VirtualCores{Assigned: &assigned}
	}
	if v, ok := d.GetOk(Arg_DeploymentTarget); ok {
		body.DeploymentTarget = expandDeploymentTarget(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk(Arg_VirtualSerialNumber); ok {
		body.VirtualSerialNumber = expandUnshelveVirtualSerialNumber(v.([]any))
	}

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)

	log.Printf("Calling the IBM PI Unshelve on the instance %s", id)
	_, err = client.Unshelve(id, body)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Unshelve failed: %s", err.Error()), "ibm_pi_instance_unshelve", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	_, err = isWaitForPIInstanceUnshelved(ctx, client, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForPIInstanceUnshelved failed: %s", err.Error()), "ibm_pi_instance_unshelve", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, id))

	return resourceIBMPIInstanceUnshelveRead(ctx, d, meta)
}

func resourceIBMPIInstanceUnshelveRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_unshelve", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID, id, err := splitID(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("splitID failed: %s", err.Error()), "ibm_pi_instance_unshelve", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	powervmdata, err := client.Get(id)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Get failed: %s", err.Error()), "ibm_pi_instance_unshelve", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.Set(Arg_CloudInstanceID, cloudInstanceID)
	d.Set(Arg_InstanceID, id)
	d.Set(Attr_Status, powervmdata.Status)

	return nil
}

func resourceIBMPIInstanceUnshelveDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId("")
	return nil
}

func expandUnshelveVirtualSerialNumber(vsnList []any) *models.UnshelveVirtualSerialNumber {
	vsnItemMap := vsnList[0].(map[string]any)
	softwareTier := models.SoftwareTier(vsnItemMap[Attr_SoftwareTier].(string))
	return &models.UnshelveVirtualSerialNumber{
		SoftwareTier: &softwareTier,
	}
}

func isWaitForPIInstanceUnshelved(ctx context.Context, client *instance.IBMPIInstanceClient, id string, timeout time.Duration) (any, error) {
	log.Printf("Waiting for the instance %s to be unshelved", id)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Pending, State_Shelved},
		Target:     []string{State_Active, State_Error},
		Refresh:    isPIInstanceUnshelveRefreshFunc(client, id),
		Delay:      30 * time.Second,
		MinTimeout: 1 * time.Minute,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isPIInstanceUnshelveRefreshFunc(client *instance.IBMPIInstanceClient, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		pvm, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}

		status := strings.ToLower(*pvm.Status)
		if status == State_Active {
			return pvm, State_Active, nil
		}

		if status == State_Error {
			if pvm.Fault != nil {
				return pvm, status, fmt.Errorf("failed to unshelve the instance: %s", pvm.Fault.Message)
			}
			return pvm, status, fmt.Errorf("failed to unshelve the instance")
		}

		return pvm, State_Pending, nil
	}
}
