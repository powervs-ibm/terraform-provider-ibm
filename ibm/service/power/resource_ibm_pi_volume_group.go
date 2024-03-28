// Copyright IBM Corp. 2022 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_volume_groups"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPIVolumeGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIVolumeGroupCreate,
		ReadContext:   resourceIBMPIVolumeGroupRead,
		UpdateContext: resourceIBMPIVolumeGroupUpdate,
		DeleteContext: resourceIBMPIVolumeGroupDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			Arg_CloudInstanceID: {
				Description: "Cloud Instance ID - This is the service_instance_id.",
				Required:    true,
				Type:        schema.TypeString,
			},
			Arg_VolumeGroupName: {
				ConflictsWith: []string{Arg_VolumeGroupConsistencyGroupName},
				Description:   "Volume Group Name to create",
				Optional:      true,
				Type:          schema.TypeString,
			},
			Arg_VolumeGroupConsistencyGroupName: {
				ConflictsWith: []string{Arg_VolumeGroupName},
				Description:   "The name of consistency group at storage controller level",
				Optional:      true,
				Type:          schema.TypeString,
			},
			Arg_VolumeIds: {
				Description: "List of volumes to add in volume group",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Set:         schema.HashString,
				Type:        schema.TypeSet,
			},

			// Computed Attributes
			Attr_VolumeGroupID: {
				Computed:    true,
				Description: "Volume Group ID",
				Type:        schema.TypeString,
			},
			Attr_VolumeGroupStatus: {
				Computed:    true,
				Description: "Volume Group Status",
				Type:        schema.TypeString,
			},
			Attr_ReplicationStatus: {
				Computed:    true,
				Description: "Volume Group Replication Status",
				Type:        schema.TypeString,
			},
			Attr_ConsistencyGroupName: {
				Computed:    true,
				Description: "Consistency Group Name if volume is a part of volume group",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceIBMPIVolumeGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	vgName := d.Get(Arg_VolumeGroupName).(string)
	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	body := &models.VolumeGroupCreate{
		Name: vgName,
	}

	volids := flex.ExpandStringList((d.Get(Arg_VolumeIds).(*schema.Set)).List())
	body.VolumeIDs = volids

	if v, ok := d.GetOk(Arg_VolumeGroupConsistencyGroupName); ok {
		body.ConsistencyGroupName = v.(string)
	}

	client := instance.NewIBMPIVolumeGroupClient(ctx, sess, cloudInstanceID)
	vg, err := client.CreateVolumeGroup(body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *vg.ID))

	_, err = isWaitForIBMPIVolumeGroupAvailable(ctx, client, *vg.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIBMPIVolumeGroupRead(ctx, d, meta)
}

func resourceIBMPIVolumeGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, vgID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVolumeGroupClient(ctx, sess, cloudInstanceID)

	vg, err := client.GetDetails(vgID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("volume_group_id", vg.ID)
	d.Set("volume_group_status", vg.Status)
	d.Set("consistency_group_name", vg.ConsistencyGroupName)
	d.Set("replication_status", vg.ReplicationStatus)
	d.Set(Arg_VolumeGroupName, vg.Name)
	d.Set(Arg_VolumeIds, vg.VolumeIDs)
	d.Set("status_description_errors", flattenVolumeGroupStatusDescription(vg.StatusDescription.Errors))

	return nil
}

func resourceIBMPIVolumeGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, vgID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVolumeGroupClient(ctx, sess, cloudInstanceID)
	if d.HasChanges(Arg_VolumeIds) {
		old, new := d.GetChange(Arg_VolumeIds)
		oldList := old.(*schema.Set)
		newList := new.(*schema.Set)
		body := &models.VolumeGroupUpdate{
			AddVolumes:    flex.ExpandStringList(newList.Difference(oldList).List()),
			RemoveVolumes: flex.ExpandStringList(oldList.Difference(newList).List()),
		}
		err := client.UpdateVolumeGroup(vgID, body)
		if err != nil {
			return diag.FromErr(err)
		}
		_, err = isWaitForIBMPIVolumeGroupAvailable(ctx, client, vgID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIBMPIVolumeGroupRead(ctx, d, meta)
}
func resourceIBMPIVolumeGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, vgID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVolumeGroupClient(ctx, sess, cloudInstanceID)

	volids := flex.ExpandStringList((d.Get(Arg_VolumeIds).(*schema.Set)).List())
	if len(volids) > 0 {
		body := &models.VolumeGroupUpdate{
			RemoveVolumes: volids,
		}
		err = client.UpdateVolumeGroup(vgID, body)
		if err != nil {
			return diag.FromErr(err)
		}
		_, err = isWaitForIBMPIVolumeGroupAvailable(ctx, client, vgID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = client.DeleteVolumeGroup(vgID)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = isWaitForIBMPIVolumeGroupDeleted(ctx, client, vgID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
func isWaitForIBMPIVolumeGroupAvailable(ctx context.Context, client *instance.IBMPIVolumeGroupClient, id string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for Volume Group (%s) to be available.", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", Attr_VolumeProvisioning},
		Target:     []string{Attr_VolumeProvisioningDone},
		Refresh:    isIBMPIVolumeGroupRefreshFunc(client, id),
		Delay:      10 * time.Second,
		MinTimeout: 2 * time.Minute,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isIBMPIVolumeGroupRefreshFunc(client *instance.IBMPIVolumeGroupClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vg, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}

		if vg.Status == "available" {
			return vg, Attr_VolumeProvisioningDone, nil
		}

		return vg, Attr_VolumeProvisioning, nil
	}
}

func isWaitForIBMPIVolumeGroupDeleted(ctx context.Context, client *instance.IBMPIVolumeGroupClient, id string, timeout time.Duration) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleting", "updating"},
		Target:     []string{"deleted"},
		Refresh:    isIBMPIVolumeGroupDeleteRefreshFunc(client, id),
		Delay:      10 * time.Second,
		MinTimeout: 2 * time.Minute,
		Timeout:    timeout,
	}
	return stateConf.WaitForStateContext(ctx)
}

func isIBMPIVolumeGroupDeleteRefreshFunc(client *instance.IBMPIVolumeGroupClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vg, err := client.Get(id)
		if err != nil {
			uErr := errors.Unwrap(err)
			switch uErr.(type) {
			case *p_cloud_volume_groups.PcloudVolumegroupsGetNotFound:
				log.Printf("[DEBUG] volume-group does not exist while deleteing %v", err)
				return vg, "deleted", nil
			}
			return nil, "", err
		}
		if vg == nil {
			return vg, "deleted", nil
		}
		return vg, "deleting", nil
	}
}
