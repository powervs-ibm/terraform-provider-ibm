// Copyright IBM Corp. 2022 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceIBMPIInstanceAction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceActionCreate,
		ReadContext:   resourceIBMPIInstanceActionRead,
		UpdateContext: resourceIBMPIInstanceActionUpdate,
		DeleteContext: resourceIBMPIInstanceActionDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
		},

		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
			_, hasChapName := diff.GetOk(Arg_ChapName)
			_, hasChapSecret := diff.GetOk(Arg_ChapSecret)
			if (hasChapName || hasChapSecret) && diff.Get(Arg_Action).(string) != Action_Start {
				return fmt.Errorf("%s and %s can only be set when %s is \"%s\"", Arg_ChapName, Arg_ChapSecret, Arg_Action, Action_Start)
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_Action: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validate.ValidateAllowedStringValues(
					[]string{Action_Dumprestart, Action_HardReboot, Action_ImmediateShutdown,
						Action_ResetState, Action_Start, Action_Stop, Action_SoftReboot}),
				Description: "PVM instance action type",
			},
			Arg_ChapName: {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{Arg_ChapSecret},
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\x00-\x7F]{1,32}$`), "must be 1 to 32 ASCII characters"),
				Description:  "CHAP authentication name for iSCSI D-mode IPL. Can only be set when pi_action is start.",
			},
			Arg_ChapSecret: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{Arg_ChapName},
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\x00-\x7F]{12,32}$`), "must be 12 to 32 ASCII characters"),
				Description:  "CHAP authentication secret for iSCSI D-mode IPL. Can only be set when pi_action is start.",
			},
			Arg_CloudInstanceID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PI Cloud instance id",
			},
			Arg_HealthStatus: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{OK, Warning}),
				Default:      OK,
				Description:  "Set the health status of the PVM instance to connect it faster",
			},
			Arg_InstanceID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PVM instance ID",
			},
			// Attributes
			Attr_HealthStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The PVM's health status value",
			},
			Attr_Progress: {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The progress of an operation",
			},
			Attr_Status: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the PVM instance",
			},
		},
	}
}

func resourceIBMPIInstanceActionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	adiag := takeInstanceAction(ctx, d, meta, d.Timeout(schema.TimeoutCreate))
	if adiag != nil {
		return adiag
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	id := d.Get(Arg_InstanceID).(string)
	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, id))

	return resourceIBMPIInstanceActionRead(ctx, d, meta)
}

func resourceIBMPIInstanceActionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_action", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID, id, err := splitID(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("splitID failed: %s", err.Error()), "ibm_pi_instance_action", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	client := instance.NewIBMPIInstanceClient(context.Background(), sess, cloudInstanceID)
	powervmdata, err := client.Get(id)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Get failed: %s", err.Error()), "ibm_pi_instance_action", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.Set(Attr_Status, powervmdata.Status)
	d.Set(Attr_Progress, powervmdata.Progress)
	if powervmdata.Health != nil {
		d.Set(Attr_HealthStatus, powervmdata.Health.Status)
	}

	return nil
}

func resourceIBMPIInstanceActionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if d.HasChange(Arg_Action) {
		adiag := takeInstanceAction(ctx, d, meta, d.Timeout(schema.TimeoutUpdate))
		if adiag != nil {
			return adiag
		}
	}
	return resourceIBMPIInstanceActionRead(ctx, d, meta)
}

func resourceIBMPIInstanceActionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// There is no delete or unset concept for instance action
	d.SetId("")
	return nil
}

func takeInstanceAction(ctx context.Context, d *schema.ResourceData, meta any, timeout time.Duration) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_action", "create/update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	id := d.Get(Arg_InstanceID).(string)
	action := d.Get(Arg_Action).(string)
	targetHealthStatus := d.Get(Arg_HealthStatus).(string)

	var targetStatus string
	switch action {
	case Action_ImmediateShutdown, Action_Stop:
		targetStatus = State_Shutoff
	case Action_ResetState:
		targetStatus = State_Active
		targetHealthStatus = Critical
	default:
		// action is "dumprestart", "hard-reboot", "start", or "soft-reboot"
		targetStatus = State_Active
	}

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)

	// special case for action "start", "stop", "immediate-shutdown"
	// skip calling action if instance is already in desired state
	if action == Action_Start || action == Action_Stop || action == Action_ImmediateShutdown {
		pvm, err := client.Get(id)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Get failed: %s", err.Error()), "ibm_pi_instance_action", "create/update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}

		if strings.ToLower(*pvm.Status) == targetStatus && pvm.Health != nil && (pvm.Health.Status == targetHealthStatus || pvm.Health.Status == OK) {
			log.Printf("[DEBUG] skipping as action %s not needed on the instance %s", action, id)
			return nil
		}
	}

	body := &models.PVMInstanceAction{Action: &action}
	if action == Action_Start {
		if v, ok := d.GetOk(Arg_ChapName); ok {
			body.ChapName = v.(string)
		}
		if v, ok := d.GetOk(Arg_ChapSecret); ok {
			body.ChapPassword = strfmt.Password(v.(string))
		}
	}
	log.Printf("Calling the IBM PI Action %s on the instance %s", action, id)

	err = client.Action(id, body)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Action failed: %s", err.Error()), "ibm_pi_instance_action", "create/update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	log.Printf("Executed the action on the instance")

	log.Printf("Calling the check for %s opertion to check for status %s", action, targetStatus)
	_, err = isWaitForPIInstanceActionStatus(ctx, client, id, timeout, targetStatus, targetHealthStatus)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForPIInstanceActionStatus failed: %s", err.Error()), "ibm_pi_instance_action", "create/update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	return nil
}

func isWaitForPIInstanceActionStatus(ctx context.Context, client *instance.IBMPIInstanceClient, id string, timeout time.Duration, targetStatus, targetHealthStatus string) (any, error) {
	log.Printf("Waiting for the action to be performed on the instance %s", id)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Pending},
		Target:     []string{targetStatus, State_Error, ""},
		Refresh:    isPIActionRefreshFunc(client, id, targetStatus, targetHealthStatus),
		Delay:      30 * time.Second,
		MinTimeout: 2 * time.Minute,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isPIActionRefreshFunc(client *instance.IBMPIInstanceClient, id, targetStatus, targetHealthStatus string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		log.Printf("Waiting for the target status to be [ %s ]", targetStatus)
		pvm, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}

		if strings.ToLower(*pvm.Status) == targetStatus && (pvm.Health.Status == targetHealthStatus || pvm.Health.Status == OK) {
			log.Printf("The health status is now %s", pvm.Health.Status)
			return pvm, targetStatus, nil
		}

		if strings.ToLower(*pvm.Status) == State_Error {
			if pvm.Fault != nil {
				err = fmt.Errorf("failed to perform the action on the instance: %s", pvm.Fault.Message)
			} else {
				err = fmt.Errorf("failed to perform the action on the instance")
			}
			return pvm, *pvm.Status, err
		}

		return pvm, State_Pending, nil
	}
}
