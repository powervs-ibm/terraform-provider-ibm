// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
)

func ResourceIBMPISnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPISnapshotCreate,
		ReadContext:   resourceIBMPISnapshotRead,
		UpdateContext: resourceIBMPISnapshotUpdate,
		DeleteContext: resourceIBMPISnapshotDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			Attr_SnapshotName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name of the snapshot",
			},
			Arg_InstanceName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Instance name / id of the pvm",
			},
			Arg_InstanceVolumeIds: {
				Type:             schema.TypeSet,
				Optional:         true,
				Elem:             &schema.Schema{Type: schema.TypeString},
				Set:              schema.HashString,
				DiffSuppressFunc: flex.ApplyOnce,
				Description:      "List of PI volumes",
			},
			Arg_CloudInstanceID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: " Cloud Instance ID - This is the service_instance_id.",
			},
			Arg_Description: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the PVM instance snapshot",
			},

			// Computed Attributes
			Attr_SnapshotID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the PVM instance snapshot",
			},
			Attr_Status: {
				Type:     schema.TypeString,
				Computed: true,
			},
			Attr_CreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			Attr_LastUpdateDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			Attr_VolumeSnapshots: {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceIBMPISnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	instanceid := d.Get(Arg_InstanceName).(string)
	volids := flex.ExpandStringList((d.Get(Arg_InstanceVolumeIds).(*schema.Set)).List())
	name := d.Get(Attr_SnapshotName).(string)

	var description string
	if v, ok := d.GetOk("pi_description"); ok {
		description = v.(string)
	}

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)

	snapshotBody := &models.SnapshotCreate{Name: &name, Description: description}

	if len(volids) > 0 {
		snapshotBody.VolumeIDs = volids
	} else {
		log.Printf("no volumeids provided. Will snapshot the entire instance")
	}

	snapshotResponse, err := client.CreatePvmSnapShot(instanceid, snapshotBody)
	if err != nil {
		log.Printf("[DEBUG]  err %s", err)
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *snapshotResponse.SnapshotID))

	pisnapclient := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)
	_, err = isWaitForPIInstanceSnapshotAvailable(ctx, pisnapclient, *snapshotResponse.SnapshotID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIBMPISnapshotRead(ctx, d, meta)
}

func resourceIBMPISnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Calling the Snapshot Read function post create")
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, snapshotID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	snapshot := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)
	snapshotdata, err := snapshot.Get(snapshotID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set(Attr_SnapshotName, snapshotdata.Name)
	d.Set(Attr_SnapshotID, *snapshotdata.SnapshotID)
	d.Set(Attr_Status, snapshotdata.Status)
	d.Set(Attr_CreationDate, snapshotdata.CreationDate.String())
	d.Set(Attr_VolumeSnapshots, snapshotdata.VolumeSnapshots)
	d.Set(Attr_LastUpdateDate, snapshotdata.LastUpdateDate.String())

	return nil
}

func resourceIBMPISnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	log.Printf("Calling the IBM Power Snapshot  update call")
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, snapshotID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)

	if d.HasChange(Attr_SnapshotName) || d.HasChange(Attr_Description) {
		name := d.Get(Attr_SnapshotName).(string)
		description := d.Get(Attr_Description).(string)
		snapshotBody := &models.SnapshotUpdate{Name: name, Description: description}

		_, err := client.Update(snapshotID, snapshotBody)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = isWaitForPIInstanceSnapshotAvailable(ctx, client, snapshotID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIBMPISnapshotRead(ctx, d, meta)
}

func resourceIBMPISnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, snapshotID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPISnapshotClient(ctx, sess, cloudInstanceID)
	snapshot, err := client.Get(snapshotID)
	if err != nil {
		// snapshot does not exist
		d.SetId("")
		return nil
	}

	log.Printf("The snapshot  to be deleted is in the following state .. %s", snapshot.Status)

	err = client.Delete(snapshotID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = isWaitForPIInstanceSnapshotDeleted(ctx, client, snapshotID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
func isWaitForPIInstanceSnapshotAvailable(ctx context.Context, client *instance.IBMPISnapshotClient, id string, timeout time.Duration) (interface{}, error) {

	log.Printf("Waiting for PIInstance Snapshot (%s) to be available and active ", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"in_progress", "BUILD"},
		Target:     []string{"available", "ACTIVE"},
		Refresh:    isPIInstanceSnapshotRefreshFunc(client, id),
		Delay:      30 * time.Second,
		MinTimeout: 2 * time.Minute,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isPIInstanceSnapshotRefreshFunc(client *instance.IBMPISnapshotClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		snapshotInfo, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}

		//if pvm.Health.Status == helpers.PIInstanceHealthOk {
		if snapshotInfo.Status == "available" && snapshotInfo.PercentComplete == 100 {
			log.Printf("The snapshot is now available")
			return snapshotInfo, "available", nil

		}
		return snapshotInfo, "in_progress", nil
	}
}

// Delete Snapshot

func isWaitForPIInstanceSnapshotDeleted(ctx context.Context, client *instance.IBMPISnapshotClient, id string, timeout time.Duration) (interface{}, error) {

	log.Printf("Waiting for (%s) to be deleted.", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", State_Deleting},
		Target:     []string{"Not Found"},
		Refresh:    isPIInstanceSnapshotDeleteRefreshFunc(client, id),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isPIInstanceSnapshotDeleteRefreshFunc(client *instance.IBMPISnapshotClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		snapshot, err := client.Get(id)
		if err != nil {
			log.Printf("The snapshot is not found.")
			return snapshot, State_NotFound, nil
		}
		return snapshot, State_NotFound, nil

	}
}
