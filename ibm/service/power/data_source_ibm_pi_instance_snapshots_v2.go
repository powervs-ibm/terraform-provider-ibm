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

func DataSourceIBMPIInstanceSnapshotsV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIInstanceSnapshotsV2Read,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_InstanceSnapshots: {
				Computed:    true,
				Description: "List of Power Virtual Machine instance snapshots (v2) within the given cloud instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Action: {
							Computed:    true,
							Description: "Action performed on the instance snapshot.",
							Type:        schema.TypeString,
						},
						Attr_AggregatedSnapshotUsage: {
							Computed:    true,
							Description: "Aggregated usage for the snapshot, keyed by storage pool.",
							Elem:        &schema.Schema{Type: schema.TypeFloat},
							Type:        schema.TypeMap,
						},
						Attr_CompletionDate: {
							Computed:    true,
							Description: "Date of snapshot completion.",
							Type:        schema.TypeString,
						},
						Attr_CopyVolumes: {
							Computed:    true,
							Description: "List of copy volumes created as part of the snapshot.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									Attr_CopyVolumeCRN: {
										Computed:    true,
										Description: "The CRN of the copy volume.",
										Type:        schema.TypeString,
									},
									Attr_CopyVolumeID: {
										Computed:    true,
										Description: "The ID of the copy volume.",
										Type:        schema.TypeString,
									},
									Attr_CopyVolumeName: {
										Computed:    true,
										Description: "The name of the copy volume.",
										Type:        schema.TypeString,
									},
									Attr_CopyVolumeUserTags: {
										Computed:    true,
										Description: "User tags for the copy volume.",
										Elem:        &schema.Schema{Type: schema.TypeString},
										Set:         schema.HashString,
										Type:        schema.TypeSet,
									},
									Attr_SrcVolumeCRN: {
										Computed:    true,
										Description: "The CRN of the source volume.",
										Type:        schema.TypeString,
									},
								},
							},
							Type: schema.TypeList,
						},
						Attr_CRN: {
							Computed:    true,
							Description: "The CRN of this resource.",
							Type:        schema.TypeString,
						},
						Attr_CreationDate: {
							Computed:    true,
							Description: "Date of snapshot creation.",
							Type:        schema.TypeString,
						},
						Attr_Description: {
							Computed:    true,
							Description: "The description of the snapshot.",
							Type:        schema.TypeString,
						},
						Attr_ID: {
							Computed:    true,
							Description: "The unique identifier of the Power Systems Virtual Machine instance snapshot.",
							Type:        schema.TypeString,
						},
						Attr_JobID: {
							Computed:    true,
							Description: "The ID of the job associated with the snapshot operation.",
							Type:        schema.TypeString,
						},
						Attr_LastUpdateDate: {
							Computed:    true,
							Description: "Date of last update.",
							Type:        schema.TypeString,
						},
						Attr_Name: {
							Computed:    true,
							Description: "The name of the Power Systems Virtual Machine instance snapshot.",
							Type:        schema.TypeString,
						},
						Attr_PercentComplete: {
							Computed:    true,
							Description: "The snapshot completion percentage.",
							Type:        schema.TypeInt,
						},
						Attr_PVMInstanceCRN: {
							Computed:    true,
							Description: "The CRN of the PVM instance associated with the snapshot.",
							Type:        schema.TypeString,
						},
						Attr_PVMInstanceID: {
							Computed:    true,
							Description: "The ID of the PVM instance associated with the snapshot.",
							Type:        schema.TypeString,
						},
						Attr_RemotePeerSnapshot: {
							Computed:    true,
							Description: "Details of the remote peer snapshot.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									Attr_CompletionDate: {
										Computed:    true,
										Description: "Date of remote snapshot completion.",
										Type:        schema.TypeString,
									},
									Attr_CopyVolumes: {
										Computed:    true,
										Description: "List of copy volumes for the remote snapshot.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												Attr_CopyVolumeCRN: {
													Computed:    true,
													Description: "The CRN of the copy volume.",
													Type:        schema.TypeString,
												},
												Attr_CopyVolumeID: {
													Computed:    true,
													Description: "The ID of the copy volume.",
													Type:        schema.TypeString,
												},
												Attr_CopyVolumeName: {
													Computed:    true,
													Description: "The name of the copy volume.",
													Type:        schema.TypeString,
												},
												Attr_CopyVolumeUserTags: {
													Computed:    true,
													Description: "User tags for the copy volume.",
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
													Type:        schema.TypeSet,
												},
												Attr_SrcVolumeCRN: {
													Computed:    true,
													Description: "The CRN of the source volume.",
													Type:        schema.TypeString,
												},
											},
										},
										Type: schema.TypeList,
									},
									Attr_CRN: {
										Computed:    true,
										Description: "The CRN of the remote peer snapshot.",
										Type:        schema.TypeString,
									},
									Attr_ID: {
										Computed:    true,
										Description: "The ID of the remote peer snapshot.",
										Type:        schema.TypeString,
									},
									Attr_Name: {
										Computed:    true,
										Description: "The name of the remote peer snapshot.",
										Type:        schema.TypeString,
									},
									Attr_Status: {
										Computed:    true,
										Description: "The status of the remote peer snapshot.",
										Type:        schema.TypeString,
									},
									Attr_StatusDetail: {
										Computed:    true,
										Description: "Detailed status information for the remote peer snapshot.",
										Type:        schema.TypeString,
									},
								},
							},
							Type: schema.TypeList,
						},
						Attr_Status: {
							Computed:    true,
							Description: "The status of the Power Virtual Machine instance snapshot.",
							Type:        schema.TypeString,
						},
						Attr_StatusDetail: {
							Computed:    true,
							Description: "Detailed information for the last PVM instance snapshot action.",
							Type:        schema.TypeString,
						},
						Attr_Type: {
							Computed:    true,
							Description: "The type of the snapshot.",
							Type:        schema.TypeString,
						},
						Attr_UserTags: {
							Computed:    true,
							Description: "List of user tags attached to the resource.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Type:        schema.TypeSet,
						},
						Attr_VolumeSnapshots: {
							Computed:    true,
							Description: "A map of volume snapshots included in the Power Virtual Machine instance snapshot.",
							Type:        schema.TypeMap,
						},
					},
				},
				Type: schema.TypeList,
			},
		},
	}
}

func dataSourceIBMPIInstanceSnapshotsV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "(Data) ibm_pi_instance_snapshots_v2", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	client := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)
	snapshotData, err := client.V2GetAll()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("V2GetAll failed: %s", err.Error()), "(Data) ibm_pi_instance_snapshots_v2", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	var clientgenU, _ = uuid.GenerateUUID()
	d.SetId(clientgenU)
	d.Set(Attr_InstanceSnapshots, flattenSnapshotsV2(snapshotData.Snapshots, meta))

	return nil
}

func flattenSnapshotsV2(list []*models.SnapshotV2, meta any) []map[string]any {
	result := make([]map[string]any, 0, len(list))
	for _, s := range list {
		if s == nil {
			continue
		}
		l := map[string]any{
			Attr_Action:          s.Action,
			Attr_Description:     s.Description,
			Attr_ID:              *s.SnapshotID,
			Attr_JobID:           s.JobID,
			Attr_LastUpdateDate:  s.LastUpdateDate.String(),
			Attr_CreationDate:    s.CreationDate.String(),
			Attr_PercentComplete: s.PercentComplete,
			Attr_PVMInstanceCRN:  string(s.PvmInstanceCRN),
			Attr_PVMInstanceID:   s.PvmInstanceID,
			Attr_Status:          s.Status,
			Attr_StatusDetail:    s.StatusDetail,
			Attr_VolumeSnapshots: s.VolumeSnapshots,
		}
		if s.Name != nil {
			l[Attr_Name] = *s.Name
		}
		if s.Type != nil {
			l[Attr_Type] = *s.Type
		}

		if s.CompletionDate.String() != "" {
			l[Attr_CompletionDate] = s.CompletionDate.String()
		}

		if s.Crn != nil {
			l[Attr_CRN] = string(*s.Crn)
			tags, err := flex.GetGlobalTagsUsingCRN(meta, string(*s.Crn), "", UserTagType)
			if err != nil {
				log.Printf("Error on get of pi instance snapshot v2 (%s) user_tags: %s", *s.SnapshotID, err)
			}
			l[Attr_UserTags] = tags
		}

		if s.AggregatedSnapshotUsage != nil {
			usage := make(map[string]any, len(s.AggregatedSnapshotUsage))
			for k, v := range s.AggregatedSnapshotUsage {
				usage[k] = v
			}
			l[Attr_AggregatedSnapshotUsage] = usage
		}

		l[Attr_CopyVolumes] = flattenCopyVolumesV2(s.CopyVolumes)

		if s.RemotePeerSnapshot != nil {
			l[Attr_RemotePeerSnapshot] = flattenRemotePeerSnapshotV2(s.RemotePeerSnapshot)
		}

		result = append(result, l)
	}
	return result
}
