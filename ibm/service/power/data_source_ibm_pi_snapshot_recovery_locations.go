// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceIBMPISnapshotRecoveryLocations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPISnapshotRecoveryLocationsRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_SnapshotRecoveryLocations: {
				Computed:    true,
				Description: "List of snapshot recovery locations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Attr_Location: {
							Computed:    true,
							Description: "The region zone location of the snapshot recovery site.",
							Type:        schema.TypeString,
						},
						Attr_ZonalSnapshotPools: {
							Computed:    true,
							Description: "List of zonal snapshot pools available at this location.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									Attr_RegionalSnapshotLocations: {
										Computed:    true,
										Description: "List of regional snapshot locations available from this pool.",
										Elem:        &schema.Schema{Type: schema.TypeString},
										Type:        schema.TypeList,
									},
									Attr_StoragePool: {
										Computed:    true,
										Description: "The storage pool name.",
										Type:        schema.TypeString,
									},
								},
							},
							Type: schema.TypeList,
						},
					},
				},
				Type: schema.TypeList,
			},
		},
	}
}

func dataSourceIBMPISnapshotRecoveryLocationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "(Data) ibm_pi_snapshot_recovery_locations", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	client := instance.NewIBMPISnapshotRecoveryClient(ctx, sess, cloudInstanceID)
	recoveryLocations, err := client.GetAll()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetAll failed: %s", err.Error()), "(Data) ibm_pi_snapshot_recovery_locations", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	results := make([]map[string]any, 0, len(recoveryLocations.SnapshotRecoveryLocations))
	for _, rl := range recoveryLocations.SnapshotRecoveryLocations {
		if rl == nil {
			continue
		}
		zonalPools := make([]map[string]any, 0, len(rl.ZonalSnapshotPools))
		for _, pool := range rl.ZonalSnapshotPools {
			if pool == nil {
				continue
			}
			p := map[string]any{
				Attr_StoragePool:               pool.StoragePool,
				Attr_RegionalSnapshotLocations: pool.RegionalSnapshotLocations,
			}
			zonalPools = append(zonalPools, p)
		}
		l := map[string]any{
			Attr_Location:           rl.Location,
			Attr_ZonalSnapshotPools: zonalPools,
		}
		results = append(results, l)
	}

	var clientgenU, _ = uuid.GenerateUUID()
	d.SetId(clientgenU)
	d.Set(Attr_SnapshotRecoveryLocations, results)

	return nil
}
