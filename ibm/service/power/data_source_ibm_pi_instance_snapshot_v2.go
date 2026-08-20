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

func DataSourceIBMPIInstanceSnapshotV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIInstanceSnapshotV2Read,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_SnapshotID: {
				Description:  "The unique identifier of the Power Systems Virtual Machine instance snapshot.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
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
				Description: "Creation date of the snapshot.",
				Type:        schema.TypeString,
			},
			Attr_Description: {
				Computed:    true,
				Description: "The description of the snapshot.",
				Type:        schema.TypeString,
			},
			Attr_JobID: {
				Computed:    true,
				Description: "The ID of the job associated with the snapshot operation.",
				Type:        schema.TypeString,
			},
			Attr_LastUpdateDate: {
				Computed:    true,
				Description: "The last updated date of the snapshot.",
				Type:        schema.TypeString,
			},
			Attr_Name: {
				Computed:    true,
				Description: "The name of the snapshot.",
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
	}
}

func dataSourceIBMPIInstanceSnapshotV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "(Data) ibm_pi_instance_snapshot_v2", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	snapshotID := d.Get(Arg_SnapshotID).(string)

	client := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)
	snapshotData, err := client.V2Get(snapshotID)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("V2Get failed: %s", err.Error()), "(Data) ibm_pi_instance_snapshot_v2", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(*snapshotData.SnapshotID)
	d.Set(Attr_Action, snapshotData.Action)
	d.Set(Attr_CreationDate, snapshotData.CreationDate.String())
	d.Set(Attr_Description, snapshotData.Description)
	d.Set(Attr_JobID, snapshotData.JobID)
	d.Set(Attr_LastUpdateDate, snapshotData.LastUpdateDate.String())
	d.Set(Attr_PercentComplete, snapshotData.PercentComplete)
	d.Set(Attr_PVMInstanceCRN, string(snapshotData.PvmInstanceCRN))
	d.Set(Attr_PVMInstanceID, snapshotData.PvmInstanceID)
	d.Set(Attr_Status, snapshotData.Status)
	d.Set(Attr_StatusDetail, snapshotData.StatusDetail)
	d.Set(Attr_VolumeSnapshots, snapshotData.VolumeSnapshots)
	if snapshotData.Name != nil {
		d.Set(Attr_Name, *snapshotData.Name)
	}
	if snapshotData.Type != nil {
		d.Set(Attr_Type, *snapshotData.Type)
	}

	if snapshotData.CompletionDate.String() != "" {
		d.Set(Attr_CompletionDate, snapshotData.CompletionDate.String())
	}

	if snapshotData.Crn != nil {
		d.Set(Attr_CRN, string(*snapshotData.Crn))
		tags, err := flex.GetGlobalTagsUsingCRN(meta, string(*snapshotData.Crn), "", UserTagType)
		if err != nil {
			log.Printf("Error on get of pi instance snapshot v2 (%s) user_tags: %s", *snapshotData.SnapshotID, err)
		}
		d.Set(Attr_UserTags, tags)
	}

	if snapshotData.AggregatedSnapshotUsage != nil {
		usage := make(map[string]any, len(snapshotData.AggregatedSnapshotUsage))
		for k, v := range snapshotData.AggregatedSnapshotUsage {
			usage[k] = v
		}
		d.Set(Attr_AggregatedSnapshotUsage, usage)
	}

	d.Set(Attr_CopyVolumes, flattenCopyVolumesV2(snapshotData.CopyVolumes))

	if snapshotData.RemotePeerSnapshot != nil {
		d.Set(Attr_RemotePeerSnapshot, flattenRemotePeerSnapshotV2(snapshotData.RemotePeerSnapshot))
	}

	return nil
}
