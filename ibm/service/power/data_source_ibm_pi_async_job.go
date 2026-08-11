// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceIBMPIAsyncJob() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIAsyncJobRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_AsyncJobID: {
				Description:  "The ID of the asynchronous job.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_Action: {
				Computed:    true,
				Description: "The action being performed in the job.",
				Type:        schema.TypeString,
			},
			Attr_ChildJobs: {
				Computed:    true,
				Description: "Array of child jobs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Action: {
							Computed:    true,
							Description: "Action of the child job.",
							Type:        schema.TypeString,
						},
						Attr_ID: {
							Computed:    true,
							Description: "ID of the child job.",
							Type:        schema.TypeString,
						},
						Attr_ProgressPercent: {
							Computed:    true,
							Description: "Progress percentage of the child job.",
							Type:        schema.TypeInt,
						},
						Attr_ResourceID: {
							Computed:    true,
							Description: "ID of the resource being acted upon.",
							Type:        schema.TypeString,
						},
						Attr_ResourceType: {
							Computed:    true,
							Description: "Type of resource being acted upon.",
							Type:        schema.TypeString,
						},
						Attr_Status: {
							Computed:    true,
							Description: "Status of the child job.",
							Type:        schema.TypeString,
						},
					},
				},
				Type: schema.TypeList,
			},
			Attr_CompletionDate: {
				Computed:    true,
				Description: "Date the job completed.",
				Type:        schema.TypeString,
			},
			Attr_CreationDate: {
				Computed:    true,
				Description: "Date the job was created.",
				Type:        schema.TypeString,
			},
			Attr_ErrorMessage: {
				Computed:    true,
				Description: "Detailed information of error encountered during job processing.",
				Type:        schema.TypeString,
			},
			Attr_LastUpdateDate: {
				Computed:    true,
				Description: "Date the job was last updated.",
				Type:        schema.TypeString,
			},
			Attr_ParentAsyncJobID: {
				Computed:    true,
				Description: "ID of the parent async job.",
				Type:        schema.TypeString,
			},
			Attr_ProgressPercent: {
				Computed:    true,
				Description: "Percentage of the job that has completed.",
				Type:        schema.TypeInt,
			},
			Attr_ResourceID: {
				Computed:    true,
				Description: "ID of the resource being acted upon in the job.",
				Type:        schema.TypeString,
			},
			Attr_ResourceType: {
				Computed:    true,
				Description: "Type of resource being acted upon in the job.",
				Type:        schema.TypeString,
			},
			Attr_Status: {
				Computed:    true,
				Description: "Status of the job.",
				Type:        schema.TypeString,
			},
		},
	}
}

func dataSourceIBMPIAsyncJobRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "(Data) ibm_pi_async_job", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	asyncJobID := d.Get(Arg_AsyncJobID).(string)
	client := instance.NewIBMPIAsyncJobClient(ctx, sess, cloudInstanceID)
	asyncJob, err := client.Get(asyncJobID)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Get failed: %s", err.Error()), "(Data) ibm_pi_async_job", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(*asyncJob.ID)
	d.Set(Attr_Action, asyncJob.Action)
	d.Set(Attr_ChildJobs, flattenChildAsyncJobs(asyncJob.ChildJobs))
	if !asyncJob.CompletionDate.IsZero() {
		d.Set(Attr_CompletionDate, asyncJob.CompletionDate.String())
	}
	d.Set(Attr_CreationDate, asyncJob.CreationDate.String())
	d.Set(Attr_ErrorMessage, asyncJob.ErrorMessage)
	if !asyncJob.LastUpdateDate.IsZero() {
		d.Set(Attr_LastUpdateDate, asyncJob.LastUpdateDate.String())
	}
	d.Set(Attr_ParentAsyncJobID, asyncJob.ParentAsyncJobID)
	d.Set(Attr_ProgressPercent, asyncJob.ProgressPercent)
	d.Set(Attr_ResourceID, asyncJob.ResourceID)
	d.Set(Attr_ResourceType, asyncJob.ResourceType)
	d.Set(Attr_Status, asyncJob.Status)

	return nil
}
