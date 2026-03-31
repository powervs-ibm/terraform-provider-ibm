// Copyright IBM Corp. 2025 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPIInstanceVpmemVolumes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceVpmemVolumesCreate,
		ReadContext:   resourceIBMPIInstanceVpmemVolumesRead,
		UpdateContext: resourceIBMPIInstanceVpmemVolumesUpdate,
		DeleteContext: resourceIBMPIInstanceVpmemVolumesDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},
		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v any) error {
				return flex.ResourcePowerUserTagsCustomizeDiff(diff)
			},
			// When volumes are renamed, propagate only the name change into the
			// computed Attr_Volumes so Terraform shows a precise diff (old→new)
			// rather than marking every volume as (known after apply).
			func(_ context.Context, diff *schema.ResourceDiff, v any) error {
				if !diff.HasChange(Arg_VPMEMVolumes) {
					return nil
				}
				old, new := diff.GetChange(Arg_VPMEMVolumes)
				oldList := old.([]any)
				newList := new.([]any)

				if len(oldList) != 0 && len(oldList) != len(newList) {
					opErr := flex.FmtErrorf("Number of vPMEM cannot be changed after resource is created. Please recreate resource to change that number ")
					tfErr := flex.TerraformErrorf(opErr, fmt.Sprintf("operation failed: %s", opErr.Error()), "ibm_pi_instance_vpmem_volumes", "update")
					log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
					return opErr
				}

				// TypeList preserves index order, so old[i] always corresponds to new[i].
				renameMap := make(map[string]string) // old name -> new name
				for i, v := range newList {
					if i >= len(oldList) {
						break
					}
					newVol := v.(map[string]any)
					oldVol := oldList[i].(map[string]any)
					newName := newVol[Attr_Name].(string)
					oldName := oldVol[Attr_Name].(string)
					oldSize := oldVol[Attr_Size].(int)
					newSize := newVol[Attr_Size].(int)

					if oldSize != newSize {
						opErr := flex.FmtErrorf("%s cannot be updated", Attr_Size)
						tfErr := flex.TerraformErrorf(opErr, fmt.Sprintf("operation failed: %s", opErr.Error()), "ibm_pi_instance_vpmem_volumes", "update")
						log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
						return opErr
					}
					if oldName != newName {
						renameMap[oldName] = newName
					}
				}
				if len(renameMap) == 0 {
					return nil
				}

				// Apply renames to the current Attr_Volumes state.
				currentSet := diff.Get(Attr_Volumes).(*schema.Set)
				updated := make([]map[string]any, 0, currentSet.Len())
				for _, elem := range currentSet.List() {
					vol := elem.(map[string]any)
					vpmem := make(map[string]any, len(vol))
					maps.Copy(vpmem, vol)
					if name, ok := vol[Attr_Name].(string); ok {
						if newName, ok := renameMap[name]; ok {
							vpmem[Attr_Name] = newName
						}
					}
					updated = append(updated, vpmem)
				}
				return diff.SetNew(Attr_Volumes, updated)
			},
		),
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description: "This is the Power Instance id that is assigned to the account",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			Arg_PVMInstanceID: {
				Description: "PCloud PVM Instance ID.",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			Arg_UserTags: {
				Description: "List of user tags.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				ForceNew:    true,
				Optional:    true,
				Set:         schema.HashString,
				Type:        schema.TypeSet,
			},
			Arg_VPMEMVolumes: {
				Description: "Description of volume(s) to create.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Name: {
							Description: "Volume base name.",
							Required:    true,
							Type:        schema.TypeString,
						},
						Attr_Size: {
							Description: "Volume size (GB).",
							Required:    true,
							Type:        schema.TypeInt,
						},
						Attr_VolumeID: {
							Computed:    true,
							Description: "Volume ID.",
							Type:        schema.TypeString,
						},
					},
				},
				MaxItems: 4,
				MinItems: 1,
				Required: true,
				Type:     schema.TypeList,
			},

			// Attributes
			Attr_Volumes: vpmemVolumeSchema(),
		},
	}
}

func resourceIBMPIInstanceVpmemVolumesCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	pvmInstanceID := d.Get(Arg_PVMInstanceID).(string)
	client := instance.NewIBMPIVPMEMClient(ctx, sess, cloudInstanceID)
	var body = &models.VPMemVolumeAttach{}
	if tags, ok := d.GetOk(Arg_UserTags); ok {
		body.UserTags = flex.FlattenSet(tags.(*schema.Set))
	}

	var vpmemList []any
	if v, ok := d.GetOk(Arg_VPMEMVolumes); ok {
		vpmemList = v.([]any)
	}

	var vpmemVolumes []*models.VPMemVolumeCreate
	for _, v := range vpmemList {
		vol := v.(map[string]any)
		vpmemVolume := resourceIBMPIInstanceVpmemVolumesMapToVpMemVolumeCreate(vol)
		vpmemVolumes = append(vpmemVolumes, vpmemVolume)
	}

	body.VpmemVolumes = vpmemVolumes
	volumes, err := client.CreatePvmVpmemVolumes(pvmInstanceID, body)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("CreatePvmVpmemVolumes failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	id := fmt.Sprintf("%s/%s", cloudInstanceID, pvmInstanceID)
	for _, vol := range volumes.Volumes {
		id += "/" + *vol.UUID
		_, err = isWaitForVpmemAvailable(ctx, client, pvmInstanceID, *vol.UUID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVpmemAvailable failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
	}
	d.SetId(id)

	return resourceIBMPIInstanceVpmemVolumesRead(ctx, d, meta)
}

func resourceIBMPIInstanceVpmemVolumesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("SepIdParts failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	client := instance.NewIBMPIVPMEMClient(ctx, sess, parts[0])
	vpmemVolumes, err := client.GetAllPvmVpmemVolumes(parts[1])
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), NotFound) {
			d.SetId("")
			return nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetAllPvmVpmemVolumes failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	volIDMap := make(map[string]string)
	if vpmemVolumes.Volumes != nil {
		for _, vol := range vpmemVolumes.Volumes {
			if vol.Name != nil && vol.UUID != nil {
				volIDMap[*vol.Name] = *vol.UUID
			}
		}
	}
	vpmemList := d.Get(Arg_VPMEMVolumes).([]any)
	updatedVpmem := make([]map[string]any, 0, len(vpmemList))
	for _, v := range vpmemList {
		vol := v.(map[string]any)
		vpmem := map[string]any{
			Attr_Name: vol[Attr_Name],
			Attr_Size: vol[Attr_Size],
		}
		if id, ok := volIDMap[vol[Attr_Name].(string)]; ok {
			vpmem[Attr_VolumeID] = id
		}
		updatedVpmem = append(updatedVpmem, vpmem)
	}
	d.Set(Arg_VPMEMVolumes, updatedVpmem)

	volumes := []map[string]any{}
	if vpmemVolumes.Volumes != nil {
		for _, volume := range vpmemVolumes.Volumes {
			volumes = append(volumes, dataSourceIBMPIVPMEMVolumeToMap(volume, meta))
		}
	}
	d.Set(Attr_Volumes, volumes)

	return nil
}

func resourceIBMPIInstanceVpmemVolumesUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("SepIdParts failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	client := instance.NewIBMPIVPMEMClient(ctx, sess, parts[0])

	if d.HasChange(Arg_VPMEMVolumes) {
		old, new := d.GetChange(Arg_VPMEMVolumes)
		oldList := old.([]any)
		newList := new.([]any)

		if len(oldList) != 0 && len(oldList) != len(newList) {
			opErr := flex.FmtErrorf("Number of vPMEM cannot be changed after resource is created. Please recreate resource to change that number ")
			tfErr := flex.TerraformErrorf(opErr, fmt.Sprintf("operation failed: %s", opErr.Error()), "ibm_pi_instance_vpmem_volumes", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}

		// TypeList preserves index order: old[i] and new[i] are the same volume.
		var updatedNames []string
		for i, v := range newList {
			newVol := v.(map[string]any)
			oldVol := oldList[i].(map[string]any)
			newName := newVol[Attr_Name].(string)
			oldName := oldVol[Attr_Name].(string)
			oldSize := oldVol[Attr_Size].(int)
			newSize := newVol[Attr_Size].(int)
			volID := oldVol[Attr_VolumeID].(string)

			if oldSize != newSize {
				opErr := flex.FmtErrorf("%s cannot be updated", Attr_Size)
				tfErr := flex.TerraformErrorf(opErr, fmt.Sprintf("operation failed: %s", opErr.Error()), "ibm_pi_instance_vpmem_volumes", "update")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
			if newName == oldName || volID == "" {
				continue
			}
			var vpmemVolumeUpdate models.VPMemVolumeUpdate
			vpmemVolumeUpdate.Name = &newName
			err := client.UpdatePvmVpmemVolume(parts[1], volID, &vpmemVolumeUpdate)
			if err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("UpdatePvmVpmemVolume failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "update")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
			updatedNames = append(updatedNames, newName)
		}

		if len(updatedNames) > 0 {
			if _, err := isWaitForVpmemUpdated(ctx, client, parts[1], updatedNames, d.Timeout(schema.TimeoutUpdate)); err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVpmemUpdated failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "update")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
		}
	}

	return resourceIBMPIInstanceVpmemVolumesRead(ctx, d, meta)
}

func resourceIBMPIInstanceVpmemVolumesDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("SepIdParts failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	client := instance.NewIBMPIVPMEMClient(ctx, sess, parts[0])
	for i := 2; i < len(parts); i++ {
		err := client.DeletePvmVpmemVolume(parts[1], parts[i])
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("DeletePvmVpmemVolume failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "delete")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
		_, err = isWaitForVpmemDeleted(ctx, client, parts[1], parts[i], d.Timeout(schema.TimeoutDelete))
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVpmemDeleted failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "delete")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
	}

	d.SetId("")

	return nil
}

func resourceIBMPIInstanceVpmemVolumesMapToVpMemVolumeCreate(modelMap map[string]any) *models.VPMemVolumeCreate {
	model := &models.VPMemVolumeCreate{}
	model.Name = core.StringPtr(modelMap[Attr_Name].(string))
	model.Size = core.Int64Ptr(int64(modelMap[Attr_Size].(int)))
	return model
}

func isWaitForVpmemAvailable(ctx context.Context, client *instance.IBMPIVPMEMClient, instanceID, volID string, timeout time.Duration) (any, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Configuring},
		Target:     []string{State_Active, State_Error},
		Refresh:    isVpmemRefreshFunc(client, instanceID, volID),
		Delay:      Timeout_Delay,
		MinTimeout: Retry_Delay,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isVpmemRefreshFunc(client *instance.IBMPIVPMEMClient, instanceID, volID string) retry.StateRefreshFunc {
	return func() (any, string, error) {

		vpmemVol, err := client.GetPvmVpmemVolume(instanceID, volID)
		if err != nil {
			return nil, "", flex.FmtErrorf("[ERROR] error getting vpmem %s", err)
		}

		if strings.ToLower(*vpmemVol.Status) == State_Active {
			return vpmemVol, State_Active, nil
		}
		if strings.ToLower(*vpmemVol.Status) == State_Error {
			return vpmemVol, *vpmemVol.Status, flex.FmtErrorf("[ERROR] vpmem is in error state: %s", err)
		}

		return vpmemVol, State_Configuring, nil
	}
}

func isWaitForVpmemUpdated(ctx context.Context, client *instance.IBMPIVPMEMClient, instanceID string, newNames []string, timeout time.Duration) (any, error) {

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Updating},
		Target:     []string{State_Completed},
		Refresh:    isVpmemUpdateRefreshFunc(client, instanceID, newNames),
		Delay:      Timeout_Delay,
		MinTimeout: Retry_Delay,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isVpmemUpdateRefreshFunc(client *instance.IBMPIVPMEMClient, instanceID string, newNames []string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		vpmemVolumes, err := client.GetAllPvmVpmemVolumes(instanceID)
		if err != nil {
			return nil, "", flex.FmtErrorf("[ERROR] error getting vpmem volumes: %s", err)
		}
		numFound := 0
		for _, vpmemVolume := range vpmemVolumes.Volumes {
			if slices.Contains(newNames, *vpmemVolume.Name) {
				numFound++
			}
		}
		if numFound == len(newNames) {
			return vpmemVolumes, State_Completed, nil
		}
		return vpmemVolumes, State_Updating, nil
	}
}

func isWaitForVpmemDeleted(ctx context.Context, client *instance.IBMPIVPMEMClient, instanceID, volID string, timeout time.Duration) (any, error) {

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Retry, State_Deleting},
		Target:     []string{State_NotFound},
		Refresh:    isVpmemDeleteRefreshFunc(client, instanceID, volID),
		Delay:      Timeout_Delay,
		MinTimeout: Retry_Delay,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isVpmemDeleteRefreshFunc(client *instance.IBMPIVPMEMClient, instanceID, volID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		vpmemVol, err := client.GetPvmVpmemVolume(instanceID, volID)
		if err != nil {
			return vpmemVol, State_NotFound, nil
		}
		return vpmemVol, State_Deleting, nil
	}
}
