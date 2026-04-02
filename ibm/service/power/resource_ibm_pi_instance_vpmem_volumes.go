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
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
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

				// Build size->name maps for old and new to detect renames.
				oldBySize := make(map[int]string)
				for _, v := range old.(*schema.Set).List() {
					vol := v.(map[string]any)
					oldBySize[vol[Attr_Size].(int)] = vol[Attr_Name].(string)
				}
				renameMap := make(map[string]string) // old name -> new name
				for _, v := range new.(*schema.Set).List() {
					vol := v.(map[string]any)
					size := vol[Attr_Size].(int)
					newName := vol[Attr_Name].(string)
					if oldName, ok := oldBySize[size]; ok && oldName != newName {
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
					entry := make(map[string]any, len(vol))
					maps.Copy(entry, vol)
					if name, ok := vol[Attr_Name].(string); ok {
						if newName, ok := renameMap[name]; ok {
							entry[Attr_Name] = newName
						}
					}
					updated = append(updated, entry)
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
					},
				},
				MaxItems: 4,
				MinItems: 1,
				Required: true,
				Type:     schema.TypeSet,
			},

			// Attributes
			Attr_Volumes: vpmemVolumeSchema(),
		},
	}
}

func resourceIBMPIInstanceVpmemVolumesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		vpmemList = v.(*schema.Set).List()
	}
	var vpmemVolumes []*models.VPMemVolumeCreate
	for _, v := range vpmemList {
		vol := v.(map[string]interface{})
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

func resourceIBMPIInstanceVpmemVolumesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	volumes := []map[string]any{}
	if vpmemVolumes.Volumes != nil {
		for _, volume := range vpmemVolumes.Volumes {
			volumes = append(volumes, dataSourceIBMPIVPMEMVolumeToMap(volume, meta))
		}
	}
	if err := d.Set(Attr_Volumes, volumes); err != nil {
		log.Printf("[WARN] ibm_pi_instance_vpmem_volumes read: d.Set(%s) failed: %s", Attr_Volumes, err)
	}

	return nil
}

func resourceIBMPIInstanceVpmemVolumesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		// Build nameToID from prior Attr_Volumes state.
		volumeList, _ := d.GetChange(Attr_Volumes)
		nameToID := make(map[string]string)
		for _, v := range volumeList.(*schema.Set).List() {
			vol := v.(map[string]any)
			nameToID[vol[Attr_Name].(string)] = vol[Attr_VolumeID].(string)
		}

		old, new := d.GetChange(Arg_VPMEMVolumes)
		// TypeSet ordering is hash-based so index comparison is unreliable.
		// detect renames via set-difference matched by size.
		oldBySize := make(map[int]string) // size -> old name
		for _, v := range old.(*schema.Set).List() {
			vol := v.(map[string]any)
			oldBySize[vol[Attr_Size].(int)] = vol[Attr_Name].(string)
		}

		var newNames []string
		for _, v := range new.(*schema.Set).List() {
			vol := v.(map[string]any)
			newName := vol[Attr_Name].(string)
			size := vol[Attr_Size].(int)
			oldName, ok := oldBySize[size]
			if !ok || oldName == newName {
				continue
			}
			volID := nameToID[oldName]
			if volID == "" {
				continue
			}
			if err := client.UpdatePvmVpmemVolume(parts[1], volID, &models.VPMemVolumeUpdate{
				Name: flex.PtrToString(newName),
			}); err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("UpdatePvmVpmemVolume failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "update")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
			newNames = append(newNames, newName)
		}

		if len(newNames) > 0 {
			if _, err := isWaitForVpmemUpdated(ctx, client, parts[1], newNames, d.Timeout(schema.TimeoutUpdate)); err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVpmemUpdated failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "update")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return tfErr.GetDiag()
			}
		}
	}

	return resourceIBMPIInstanceVpmemVolumesRead(ctx, d, meta)
}

func resourceIBMPIInstanceVpmemVolumesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceIBMPIInstanceVpmemVolumesMapToVpMemVolumeCreate(modelMap map[string]interface{}) *models.VPMemVolumeCreate {
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
	return func() (interface{}, string, error) {
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

func isWaitForVpmemUpdated(ctx context.Context, client *instance.IBMPIVPMEMClient, instanceID string, newNames []string, timeout time.Duration) (*models.VPMemVolumes, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Updating},
		Target:     []string{State_Completed},
		Refresh:    isVpmemUpdateRefreshFunc(client, instanceID, newNames),
		Delay:      Timeout_Delay,
		MinTimeout: Retry_Delay,
		Timeout:    timeout,
	}

	result, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}
	vols, _ := result.(*models.VPMemVolumes)
	return vols, nil
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
