// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"strings"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/service/power"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPIInstanceUnshelve(t *testing.T) {
	name := fmt.Sprintf("tf-pi-instance-%d", acctest.RandIntRange(10, 100))
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceUnshelveConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_unshelve.unshelve", "status", strings.ToUpper(power.State_Active)),
				),
			},
		},
	})
}

func testAccCheckIBMPIInstanceUnshelveConfig(name string) string {
	return fmt.Sprintf(`
	data "ibm_pi_image" "power_image" {
		pi_cloud_instance_id = "%[1]s"
		pi_image_id          = "%[3]s"
	}
	data "ibm_pi_network" "power_networks" {
		pi_cloud_instance_id = "%[1]s"
		pi_network_id        = "%[4]s"
	}
	resource "ibm_pi_instance" "power_instance" {
		pi_cloud_instance_id  = "%[1]s"
		pi_image_id           = data.ibm_pi_image.power_image.id
		pi_instance_name      = "%[2]s"
		pi_memory             = "2"
		pi_proc_type          = "shared"
		pi_processors         = "0.25"
		pi_storage_pool       = data.ibm_pi_image.power_image.storage_pool
		pi_storage_type       = "%[5]s"
		pi_sys_type           = "s1022"
		pi_network {
			network_id = data.ibm_pi_network.power_networks.id
		}
	}
	resource "ibm_pi_instance_action" "shutdown" {
		pi_action				= "%[6]s"
		pi_cloud_instance_id	= "%[1]s"
		pi_instance_id			= resource.ibm_pi_instance.power_instance.instance_id 
	}
	resource "ibm_pi_instance_shelve" "shelve" {
	    depends_on  = [ibm_pi_instance_action.shutdown]
		pi_cloud_instance_id	= "%[1]s"
		pi_instance_id			= resource.ibm_pi_instance.power_instance.instance_id
	}

	resource "ibm_pi_instance_unshelve" "unshelve" {
	    depends_on				= [ibm_pi_instance_shelve.shelve]
		pi_cloud_instance_id	= "%[1]s"
		pi_instance_id			= resource.ibm_pi_instance.power_instance.instance_id
		pi_memory				= "2"
		pi_processors			= "0.25"
	}
	`, acc.Pi_cloud_instance_id, name, acc.Pi_image_id, acc.Pi_network_id, acc.PiStorageType, power.Action_ImmediateShutdown)
}
