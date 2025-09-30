// Copyright IBM Corp. 2025 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM/go-sdk-core/v5/core"

	"github.com/IBM-Cloud/power-go-client/power/models"
)

func ResourceIBMPIInstanceVpmenVolumes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceVpmenVolumesCreate,
		ReadContext:   resourceIBMPIInstanceVpmenVolumesRead,
		DeleteContext: resourceIBMPIInstanceVpmenVolumesDelete,
		Importer:      &schema.ResourceImporter{},

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
			Arg_Volume: {
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
				ForceNew: true,
				MaxItems: 1,
				MinItems: 1,
				Required: true,
				Type:     schema.TypeList,
			},

			// Attributes
			Attr_Volumes: vpmemVolumeSchema(),
		},
	}
}

func resourceIBMPIInstanceVpmenVolumesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	body.VpmemVolume = resourceIBMPIInstanceVpmenVolumesMapToVpMemVolumeCreate(d.Get(Arg_Volume + ".0").(map[string]any))
	volumes, err := client.CreatePvmVpmemVolumes(pvmInstanceID, body)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("CreatePvmVpmemVolumes failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	id := fmt.Sprintf("%s/%s", cloudInstanceID, pvmInstanceID)
	for _, vol := range volumes.Volumes {
		id += "/" + *vol.VolumeID
	}

	d.SetId(id)

	return resourceIBMPIInstanceVpmenVolumesRead(ctx, d, meta)
}

func resourceIBMPIInstanceVpmenVolumesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("SepIdParts failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
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
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetAllPvmVpmemVolumes failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	volumes := []map[string]any{}
	if vpmemVolumes.Volumes != nil {
		for _, volume := range vpmemVolumes.Volumes {
			vpemVol := dataSourceIBMPIVPMEMVolumeToMap(volume, meta)
			volumes = append(volumes, vpemVol)
		}
	}
	d.Set(Attr_Volumes, volumes)

	return nil
}

func resourceIBMPIInstanceVpmenVolumesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("SepIdParts failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	client := instance.NewIBMPIVPMEMClient(ctx, sess, parts[0])
	for i := 2; i < len(parts); i++ {
		err := client.DeletePvmVpmemVolume(parts[1], parts[i])
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("DeletePvmVpmemVolume failed: %s", err.Error()), "ibm_pi_instance_vpmem_volumes", "create")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
	}

	d.SetId("")

	return nil
}

func resourceIBMPIInstanceVpmenVolumesMapToVpMemVolumeCreate(modelMap map[string]interface{}) *models.VPMemVolumeCreate {
	model := &models.VPMemVolumeCreate{}
	model.Name = core.StringPtr(modelMap[Attr_Name].(string))
	model.Size = core.Int64Ptr(int64(modelMap[Attr_Size].(int)))
	return model
}
