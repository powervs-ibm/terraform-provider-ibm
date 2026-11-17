// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPIInstanceNetworkAttachmentsDataSource_basic(t *testing.T) {
	ds := "data.ibm_pi_instance_network_attachments.testacc_ds_instance_network_attachments"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceNetworkAttachmentsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ds, "id"),
				),
			},
		},
	})
}

func testAccCheckIBMPIInstanceNetworkAttachmentsDataSourceConfig() string {
	return fmt.Sprintf(`
		data "ibm_pi_instance_network_attachments" "testacc_ds_instance_network_attachments" {
			pi_cloud_instance_id = "%[1]s"
  			pi_instance_id       = "%[2]s"
		}`, acc.Pi_cloud_instance_id, acc.Pi_instance_id)
}
