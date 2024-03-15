// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"
)

func TestAccIBMPiHostgroupDataSourceBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPiHostgroupDataSourceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_pi_hostgroup.testacc_ds_hostgroup", "id"),
				),
			},
		},
	})
}

func testAccCheckIBMPiHostgroupDataSourceConfigBasic() string {
	return fmt.Sprintf(`
		data "ibm_pi_hostgroup" "testacc_ds_hostgroup" {
			pi_cloud_instance_id = "%s"
			pi_hostgroup_id  = "%s"
		}`, acc.Pi_cloud_instance_id, acc.Pi_hostgroup_id)
}
