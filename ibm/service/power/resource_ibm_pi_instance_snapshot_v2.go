// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceIBMPIInstanceSnapshotV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceSnapshotV2Create,
		ReadContext:   resourceIBMPIInstanceSnapshotV2Read,
		UpdateContext: resourceIBMPIInstanceSnapshotV2Update,
		DeleteContext: resourceIBMPIInstanceSnapshotV2Delete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_AddRemoteSnapshot: {
				Description:  "Indicates if a remote snapshot should be added. Allowed values are 'no' and 'regional'.",
				ForceNew:     true,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"no", "regional"}, false),
			},
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_Description: {
				Description: "Description of the PVM instance snapshot.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_PVMInstanceID: {
				Description:  "The ID of the PVM instance to snapshot.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_RemoteSnapshot: {
				Description: "Remote snapshot configuration.",
				ForceNew:    true,
				MaxItems:    1,
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Description: {
							Description: "Description of the remote snapshot.",
							Optional:    true,
							Type:        schema.TypeString,
						},
						Attr_Name: {
							Description: "The name of the remote snapshot.",
							Required:    true,
							Type:        schema.TypeString,
						},
						Attr_PropagateUserTags: {
							Description: "Indicates if the user tags should be propagated to the remote snapshot.",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						Attr_UserTags: {
							Description: "User tags for the remote snapshot.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Set:         schema.HashString,
							Type:        schema.TypeSet,
						},
						Attr_WorkspaceCRN: {
							Description: "The CRN of the target workspace for the remote snapshot.",
							Optional:    true,
							Type:        schema.TypeString,
						},
					},
				},
			},
			Arg_SnapshotName: {
				Description:  "The unique name of the snapshot.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_SnapshotType: {
				Description:  "The type of snapshot. Allowed values are 'in-place' and 'zonal'.",
				ForceNew:     true,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"in-place", "zonal"}, false),
			},
			Arg_UserTags: {
				Description: "The user tags attached to this resource.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Set:         schema.HashString,
				Type:        schema.TypeSet,
			},
			Arg_VolumeIDs: {
				Description:      "A list of volume IDs of the instance that will be part of the snapshot. If none are provided, then all the volumes of the instance will be part of the snapshot.",
				DiffSuppressFunc: flex.ApplyOnce,
				Elem:             &schema.Schema{Type: schema.TypeString},
				Optional:         true,
				Set:              schema.HashString,
				Type:             schema.TypeSet,
			},

			// Attributes
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
			Attr_RemotePeerSnapshot: {
				Computed:    true,
				Description: "Details of the remote peer snapshot.",
				MaxItems:    1,
				Type:        schema.TypeList,
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
			},
			Attr_SnapshotID: {
				Computed:    true,
				Description: "ID of the PVM instance snapshot.",
				Type:        schema.TypeString,
			},
			Attr_Status: {
				Computed:    true,
				Description: "Status of the PVM instance snapshot.",
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
			Attr_VolumeSnapshots: {
				Computed:    true,
				Description: "A map of volume snapshots included in the PVM instance snapshot.",
				Type:        schema.TypeMap,
			},
		},
	}
}

func resourceIBMPIInstanceSnapshotV2Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	pvmInstanceID := d.Get(Arg_PVMInstanceID).(string)
	name := d.Get(Arg_SnapshotName).(string)

	snapshotBody := &models.SnapshotCreateV2{Name: &name}

	if v, ok := d.GetOk(Arg_Description); ok {
		snapshotBody.Description = v.(string)
	}

	if v, ok := d.GetOk(Arg_SnapshotType); ok {
		t := v.(string)
		snapshotBody.Type = &t
	}

	if v, ok := d.GetOk(Arg_AddRemoteSnapshot); ok {
		ars := v.(string)
		snapshotBody.AddRemoteSnapshot = &ars
	}

	volumeIDs := flex.ExpandStringList((d.Get(Arg_VolumeIDs).(*schema.Set)).List())
	if len(volumeIDs) > 0 {
		snapshotBody.VolumeIDs = volumeIDs
	}

	if v, ok := d.GetOk(Arg_UserTags); ok {
		snapshotBody.UserTags = flex.FlattenSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(Arg_RemoteSnapshot); ok {
		rsList := v.([]any)
		if len(rsList) > 0 {
			rs := rsList[0].(map[string]any)
			remoteSnapshot := &models.RemoteSnapshot{}
			if n, ok := rs[Attr_Name].(string); ok && n != "" {
				remoteSnapshot.Name = &n
			}
			if desc, ok := rs[Attr_Description].(string); ok {
				remoteSnapshot.Description = desc
			}
			if put, ok := rs[Attr_PropagateUserTags].(bool); ok {
				remoteSnapshot.PropagateUserTags = &put
			}
			if ut, ok := rs[Attr_UserTags].(*schema.Set); ok {
				remoteSnapshot.UserTags = flex.FlattenSet(ut)
			}
			if wcrn, ok := rs[Attr_WorkspaceCRN].(string); ok && wcrn != "" {
				crn := models.CRN(wcrn)
				remoteSnapshot.WorkspaceCRN = crn
			}
			snapshotBody.RemoteSnapshot = remoteSnapshot
		}
	}

	client := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)
	snapshotResponse, err := client.V2Create(pvmInstanceID, snapshotBody)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("V2Create failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *snapshotResponse.SnapshotID))

	_, err = isWaitForPIInstanceSnapshotV2Available(ctx, client, *snapshotResponse.SnapshotID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("wait for snapshot available failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	if _, ok := d.GetOk(Arg_UserTags); ok {
		if snapshotResponse.Crn != "" {
			oldList, newList := d.GetChange(Arg_UserTags)
			err := flex.UpdateGlobalTagsUsingCRN(oldList, newList, meta, string(snapshotResponse.Crn), "", UserTagType)
			if err != nil {
				log.Printf("Error on update of pi snapshot v2 (%s) pi_user_tags during creation: %s", *snapshotResponse.SnapshotID, err)
			}
		}
	}

	return resourceIBMPIInstanceSnapshotV2Read(ctx, d, meta)
}

func resourceIBMPIInstanceSnapshotV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID, snapshotID, err := splitID(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("splitID failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	client := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)
	snapshotData, err := client.V2Get(snapshotID)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("V2Get failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.Set(Attr_SnapshotID, *snapshotData.SnapshotID)
	d.Set(Attr_CreationDate, snapshotData.CreationDate.String())
	d.Set(Arg_Description, snapshotData.Description)
	d.Set(Attr_JobID, snapshotData.JobID)
	d.Set(Attr_LastUpdateDate, snapshotData.LastUpdateDate.String())
	d.Set(Attr_PercentComplete, snapshotData.PercentComplete)
	d.Set(Attr_PVMInstanceCRN, string(snapshotData.PvmInstanceCRN))
	d.Set(Attr_Status, snapshotData.Status)
	d.Set(Attr_StatusDetail, snapshotData.StatusDetail)
	d.Set(Attr_VolumeSnapshots, snapshotData.VolumeSnapshots)
	if snapshotData.Name != nil {
		d.Set(Arg_SnapshotName, *snapshotData.Name)
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
			log.Printf("Error on get of pi snapshot v2 (%s) pi_user_tags: %s", *snapshotData.SnapshotID, err)
		}
		d.Set(Arg_UserTags, tags)
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

func resourceIBMPIInstanceSnapshotV2Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID, snapshotID, err := splitID(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("splitID failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	client := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)

	if d.HasChange(Arg_SnapshotName) || d.HasChange(Arg_Description) {
		name := d.Get(Arg_SnapshotName).(string)
		description := d.Get(Arg_Description).(string)
		snapshotBody := &models.SnapshotUpdate{Name: &name, Description: &description}

		_, err := client.Update(snapshotID, snapshotBody)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Update failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}

		_, err = isWaitForPIInstanceSnapshotV2Available(ctx, client, snapshotID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("wait for snapshot available failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
	}

	if d.HasChange(Arg_UserTags) {
		if crn, ok := d.GetOk(Attr_CRN); ok {
			oldList, newList := d.GetChange(Arg_UserTags)
			err := flex.UpdateGlobalTagsUsingCRN(oldList, newList, meta, crn.(string), "", UserTagType)
			if err != nil {
				log.Printf("Error on update of pi snapshot v2 (%s) pi_user_tags: %s", snapshotID, err)
			}
		}
	}

	return resourceIBMPIInstanceSnapshotV2Read(ctx, d, meta)
}

func resourceIBMPIInstanceSnapshotV2Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID, snapshotID, err := splitID(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("splitID failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	client := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)
	err = client.Delete(snapshotID)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Delete failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	_, err = isWaitForPIInstanceSnapshotV2Deleted(ctx, client, snapshotID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("wait for snapshot deleted failed: %s", err.Error()), "ibm_pi_instance_snapshot_v2", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId("")
	return nil
}

func isWaitForPIInstanceSnapshotV2Available(ctx context.Context, client *instance.IBMPISnapshotClient, id string, timeout time.Duration) (any, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_InProgress},
		Target:     []string{State_Available},
		Refresh:    isPIInstanceSnapshotV2RefreshFunc(client, id),
		Delay:      30 * time.Second,
		MinTimeout: 2 * time.Minute,
		Timeout:    timeout,
	}
	return stateConf.WaitForStateContext(ctx)
}

func isPIInstanceSnapshotV2RefreshFunc(client *instance.IBMPISnapshotClient, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		snapshotInfo, err := client.V2Get(id)
		if err != nil {
			return nil, "", err
		}
		if snapshotInfo.Status == State_Available && snapshotInfo.PercentComplete == 100 {
			log.Printf("The snapshot v2 is now available")
			return snapshotInfo, State_Available, nil
		}
		return snapshotInfo, State_InProgress, nil
	}
}

func isWaitForPIInstanceSnapshotV2Deleted(ctx context.Context, client *instance.IBMPISnapshotClient, id string, timeout time.Duration) (any, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Retry},
		Target:     []string{State_NotFound},
		Refresh:    isPIInstanceSnapshotV2DeleteRefreshFunc(client, id),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		Timeout:    timeout,
	}
	return stateConf.WaitForStateContext(ctx)
}

func isPIInstanceSnapshotV2DeleteRefreshFunc(client *instance.IBMPISnapshotClient, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		snapshot, err := client.V2Get(id)
		if err != nil {
			log.Printf("The snapshot v2 is not found.")
			return snapshot, State_NotFound, nil
		}
		return snapshot, State_Retry, nil
	}
}

func flattenCopyVolumesV2(list []*models.CopyVolume) []map[string]any {
	result := make([]map[string]any, 0, len(list))
	for _, cv := range list {
		if cv == nil {
			continue
		}
		m := map[string]any{}
		if cv.CopyVolumeCRN != nil {
			m[Attr_CopyVolumeCRN] = string(*cv.CopyVolumeCRN)
		}
		if cv.CopyVolumeID != nil {
			m[Attr_CopyVolumeID] = *cv.CopyVolumeID
		}
		if cv.CopyVolumeName != nil {
			m[Attr_CopyVolumeName] = *cv.CopyVolumeName
		}
		m[Attr_CopyVolumeUserTags] = cv.CopyVolumeUserTags
		if cv.SrcVolumeCRN != nil {
			m[Attr_SrcVolumeCRN] = string(*cv.SrcVolumeCRN)
		}
		result = append(result, m)
	}
	return result
}

func flattenRemotePeerSnapshotV2(rps *models.RemotePeerSnapshot) []map[string]any {
	if rps == nil {
		return nil
	}
	m := map[string]any{
		Attr_StatusDetail: rps.StatusDetail,
	}
	if rps.CompletionDate.String() != "" {
		m[Attr_CompletionDate] = rps.CompletionDate.String()
	}
	if rps.Crn != nil {
		m[Attr_CRN] = string(*rps.Crn)
	}
	if rps.ID != nil {
		m[Attr_ID] = *rps.ID
	}
	if rps.Name != nil {
		m[Attr_Name] = *rps.Name
	}
	if rps.Status != nil {
		m[Attr_Status] = *rps.Status
	}
	m[Attr_CopyVolumes] = flattenCopyVolumesV2(rps.CopyVolumes)
	return []map[string]any{m}
}
