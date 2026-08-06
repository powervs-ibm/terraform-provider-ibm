// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceIBMPIAsyncJobs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIAsyncJobsRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_AsyncJobs: {
				Computed:    true,
				Description: "List of asynchronous jobs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						Attr_ID: {
							Computed:    true,
							Description: "ID of the asynchronous job.",
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
				},
				Type: schema.TypeList,
			},
		},
	}
}

func dataSourceIBMPIAsyncJobsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "(Data) ibm_pi_async_jobs", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	client := instance.NewIBMPIAsyncJobClient(ctx, sess, cloudInstanceID)
	asyncJobs, err := client.GetAll()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetAll failed: %s", err.Error()), "(Data) ibm_pi_async_jobs", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	var clientgenU, _ = uuid.GenerateUUID()
	d.SetId(clientgenU)
	d.Set(Attr_AsyncJobs, flattenAsyncJobs(asyncJobs.AsyncJobs))

	return nil
}

func flattenAsyncJobs(jobs []*models.AsyncJob) []map[string]any {
	result := make([]map[string]any, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		j := map[string]any{
			Attr_Action:           job.Action,
			Attr_ChildJobs:        flattenChildAsyncJobs(job.ChildJobs),
			Attr_CreationDate:     job.CreationDate.String(),
			Attr_ErrorMessage:     job.ErrorMessage,
			Attr_ID:               job.ID,
			Attr_ParentAsyncJobID: job.ParentAsyncJobID,
			Attr_ProgressPercent:  job.ProgressPercent,
			Attr_ResourceID:       job.ResourceID,
			Attr_ResourceType:     job.ResourceType,
			Attr_Status:           job.Status,
		}
		if !job.CompletionDate.IsZero() {
			j[Attr_CompletionDate] = job.CompletionDate.String()
		}
		if !job.LastUpdateDate.IsZero() {
			j[Attr_LastUpdateDate] = job.LastUpdateDate.String()
		}
		result = append(result, j)
	}
	return result
}

func flattenChildAsyncJobs(children []*models.ChildAsyncJob) []map[string]any {
	result := make([]map[string]any, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}
		result = append(result, map[string]any{
			Attr_Action:          child.Action,
			Attr_ID:              child.ID,
			Attr_ProgressPercent: child.ProgressPercent,
			Attr_ResourceID:      child.ResourceID,
			Attr_ResourceType:    child.ResourceType,
			Attr_Status:          child.Status,
		})
	}
	return result
}
