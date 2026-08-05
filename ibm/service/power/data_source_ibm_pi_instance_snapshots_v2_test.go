// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPIInstanceSnapshotsV2DataSource_basic(t *testing.T) {
	snapshotResData := "data.ibm_pi_instance_snapshots_v2.testacc_ds_snapshots_v2"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceSnapshotsV2DataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(snapshotResData, "id"),
				),
			},
		},
	})
}

func testAccCheckIBMPIInstanceSnapshotsV2DataSourceConfig() string {
	return fmt.Sprintf(`
		data "ibm_pi_instance_snapshots_v2" "testacc_ds_snapshots_v2" {
			pi_cloud_instance_id = "%s"
		}`, acc.Pi_cloud_instance_id)
}
