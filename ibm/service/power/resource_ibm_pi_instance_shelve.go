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
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPIInstanceShelve() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceShelveCreate,
		ReadContext:   resourceIBMPIInstanceShelveRead,
		DeleteContext: resourceIBMPIInstanceShelveDelete,
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
			// Attributes
			Attr_Status: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the PVM instance",
			},
		},
	}
}

func resourceIBMPIInstanceShelveCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_shelve", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	id := d.Get(Arg_InstanceID).(string)

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	err = client.Shelve(id)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Shelve failed: %s", err.Error()), "ibm_pi_instance_shelve", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	_, err = isWaitForPIInstanceShelved(ctx, client, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForPIInstanceShelved failed: %s", err.Error()), "ibm_pi_instance_shelve", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, id))

	return resourceIBMPIInstanceShelveRead(ctx, d, meta)
}

func resourceIBMPIInstanceShelveRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_shelve", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID, id, err := splitID(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("splitID failed: %s", err.Error()), "ibm_pi_instance_shelve", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	powervmdata, err := client.Get(id)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Get failed: %s", err.Error()), "ibm_pi_instance_shelve", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.Set(Arg_CloudInstanceID, cloudInstanceID)
	d.Set(Arg_InstanceID, id)
	d.Set(Attr_Status, powervmdata.Status)

	return nil
}

func resourceIBMPIInstanceShelveDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId("")
	return nil
}

func isWaitForPIInstanceShelved(ctx context.Context, client *instance.IBMPIInstanceClient, id string, timeout time.Duration) (any, error) {
	log.Printf("Waiting for the instance %s to be shelved", id)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Pending, State_Active},
		Target:     []string{State_Shelved, State_Error},
		Refresh:    isPIInstanceShelveRefreshFunc(client, id),
		Delay:      30 * time.Second,
		MinTimeout: 1 * time.Minute,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isPIInstanceShelveRefreshFunc(client *instance.IBMPIInstanceClient, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		pvm, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}

		status := strings.ToLower(*pvm.Status)
		if status == State_Shelved {
			return pvm, State_Shelved, nil
		}

		if status == State_Error {
			if pvm.Fault != nil {
				return pvm, status, fmt.Errorf("failed to shelve the instance: %s", pvm.Fault.Message)
			}
			return pvm, status, fmt.Errorf("failed to shelve the instance")
		}

		return pvm, State_Pending, nil
	}
}
