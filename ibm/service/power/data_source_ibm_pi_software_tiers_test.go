// Copyright IBM Corp. 2022 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPISoftwareTiersDataSourceBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPISoftwareTiersDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_pi_software_tiers.test", "id"),
				),
			},
		},
	})
}

func testAccCheckIBMPISoftwareTiersDataSourceConfig() string {
	return fmt.Sprintf(`
		data "ibm_pi_software_tiers" "test" {
			pi_cloud_instance_id = "%s"
		}`, acc.Pi_cloud_instance_id)
}
