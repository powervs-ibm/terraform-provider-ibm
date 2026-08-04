// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPISnapshotRecoveryLocationDataSourceBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPISnapshotRecoveryLocationDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_pi_snapshot_recovery_location.testacc_snapshot_recovery_location", "id"),
					resource.TestCheckResourceAttrSet("data.ibm_pi_snapshot_recovery_location.testacc_snapshot_recovery_location", "location"),
				),
			},
		},
	})
}

func testAccCheckIBMPISnapshotRecoveryLocationDataSourceConfig() string {
	return fmt.Sprintf(`
		data "ibm_pi_snapshot_recovery_location" "testacc_snapshot_recovery_location" {
			pi_cloud_instance_id = "%s"
		}`, acc.Pi_cloud_instance_id)
}
