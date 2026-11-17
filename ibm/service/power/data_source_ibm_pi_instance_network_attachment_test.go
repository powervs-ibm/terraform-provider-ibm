// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPIInstanceNetworkAttachmentDataSource_basic(t *testing.T) {
	ds := "data.ibm_pi_instance_network_attachment.testacc_ds_instance_network_attachment"
	name := fmt.Sprintf("tf-pi-instnetattachmentds-%d", acctest.RandIntRange(10, 100))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMPIInstanceNetworkAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceNetworkAttachmentDataSourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(ds, "id"),
					resource.TestCheckResourceAttr(ds, "vpc_virtual_network_interface.0.bandwidth", "10"),
					resource.TestCheckResourceAttrSet(ds, "vpc_virtual_network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(ds, "vpc_virtual_network_interface.0.mac_address"),
				),
			},
		},
	})
}

func testAccCheckIBMPIInstanceNetworkAttachmentDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "ibm_pi_instance" "power_instance" {
		pi_cloud_instance_id = "%[1]s"
		pi_memory             = "2"
		pi_processors         = "0.25"
		pi_instance_name      = "%[2]s"
		pi_proc_type          = "shared"
		pi_image_id           = "%[3]s"
		pi_sys_type           = "s1022"
		pi_network {
			network_id = "%[4]s"
		}
	}
	resource "ibm_pi_instance_network_attachment" "example" {
		pi_cloud_instance_id = "%[1]s"
		pi_instance_id       = ibm_pi_instance.power_instance.instance_id
		vpc_virtual_network_interface {
			crn       = "%[5]s"
			bandwidth = 10
		}
	}
	data "ibm_pi_instance_network_attachment" "testacc_ds_instance_network_attachment" {
		pi_cloud_instance_id    = "%[1]s"
		pi_instance_id          = ibm_pi_instance.power_instance.instance_id
		pi_network_attachment_id = ibm_pi_instance_network_attachment.example.network_attachment_id
	}
	`, acc.Pi_cloud_instance_id, name, acc.Pi_image, acc.Pi_network_id, acc.Pi_external_network_interface_crn)
}
