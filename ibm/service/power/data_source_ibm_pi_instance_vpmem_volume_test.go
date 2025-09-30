// Copyright IBM Corp. 2025 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"
)

func TestAccIBMPIInstanceVpmemVolumeDataSourceBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceVpmemVolumeDataSourceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ibm_pi_instance_vpmem_volume.instance_vpmem_volume_instance", "id"),
					resource.TestCheckResourceAttrSet("data.ibm_pi_instance_vpmem_volume.instance_vpmem_volume_instance", "created_at"),
					resource.TestCheckResourceAttrSet("data.ibm_pi_instance_vpmem_volume.instance_vpmem_volume_instance", "crn"),
					resource.TestCheckResourceAttrSet("data.ibm_pi_instance_vpmem_volume.instance_vpmem_volume_instance", "href"),
					resource.TestCheckResourceAttrSet("data.ibm_pi_instance_vpmem_volume.instance_vpmem_volume_instance", "name"),
					resource.TestCheckResourceAttrSet("data.ibm_pi_instance_vpmem_volume.instance_vpmem_volume_instance", "size"),
					resource.TestCheckResourceAttrSet("data.ibm_pi_instance_vpmem_volume.instance_vpmem_volume_instance", "status"),
					resource.TestCheckResourceAttrSet("data.ibm_pi_instance_vpmem_volume.instance_vpmem_volume_instance", "volume_id"),
				),
			},
		},
	})
}

func testAccCheckIBMPIInstanceVpmemVolumeDataSourceConfigBasic() string {
	return fmt.Sprintf(`
		data "ibm_pi_instance_vpmem_volume" "instance_vpmem_volume_instance" {
			pi_cloud_instance_id = "%s"
			pi_pvm_instance_id = "%s"
			pi_vpmem_volume_id = "%s"
		}
	`, acc.Pi_cloud_instance_id, acc.Pi_instance_name, acc.Pi_volume_id)
}
