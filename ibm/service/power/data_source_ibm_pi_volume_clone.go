// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
)

func DataSourceIBMPIVolumeClone() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIVolumeCloneRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_VolumeCloneTaskID: {
				Description:  "The ID of the volume clone task.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_ClonedVolumes: clonedVolumesSchema(),
			Attr_FailureReason: {
				Computed:    true,
				Description: "The reason the clone volumes task has failed.",
				Type:        schema.TypeString,
			},
			Attr_PercentComplete: {
				Computed:    true,
				Description: "The completion percentage of the volume clone task.",
				Type:        schema.TypeInt,
			},
			Attr_Status: {
				Computed:    true,
				Description: "The status of the volume clone task.",
				Type:        schema.TypeString,
			},
		},
	}
}

func dataSourceIBMPIVolumeCloneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	client := instance.NewIBMPICloneVolumeClient(ctx, sess, cloudInstanceID)
	volClone, err := client.Get(d.Get(Arg_VolumeCloneTaskID).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(d.Get(Arg_VolumeCloneTaskID).(string))
	if volClone.Status != nil {
		d.Set(Attr_Status, *volClone.Status)
	}
	d.Set(Attr_FailureReason, volClone.FailedReason)
	if volClone.PercentComplete != nil {
		d.Set(Attr_PercentComplete, *volClone.PercentComplete)
	}
	d.Set(Attr_ClonedVolumes, flattenClonedVolumes(volClone.ClonedVolumes))

	return nil
}
