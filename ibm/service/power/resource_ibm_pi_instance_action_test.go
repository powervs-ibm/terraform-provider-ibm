// Copyright IBM Corp. 2022 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/service/power"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPIInstanceAction(t *testing.T) {
	name := fmt.Sprintf("tf-pi-instance-%d", acctest.RandIntRange(10, 100))
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceActionConfig(name, power.Action_Stop),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "status", strings.ToUpper(power.State_Shutoff)),
				),
			},
			{
				// Try to stop already stopped instance
				Config: testAccCheckIBMPIInstanceActionConfig(name, power.Action_Stop),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "status", strings.ToUpper(power.State_Shutoff)),
				),
			},
			{
				Config: testAccCheckIBMPIInstanceActionConfig(name, power.Action_Start),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "status", strings.ToUpper(power.State_Active)),
				),
			},
			{
				// Try to start already started instance
				Config: testAccCheckIBMPIInstanceActionConfig(name, power.Action_Start),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "status", strings.ToUpper(power.State_Active)),
				),
			},
		},
	})
}

func TestAccIBMPIInstanceActionHardReboot(t *testing.T) {
	name := fmt.Sprintf("tf-pi-instance-%d", acctest.RandIntRange(10, 100))
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceActionWithHealthStatusConfig(name, power.Action_HardReboot, power.Warning),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "status", strings.ToUpper(power.State_Active)),
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "health_status", power.Warning),
				),
			},
		},
	})
}

func TestAccIBMPIInstanceActionResetState(t *testing.T) {
	name := fmt.Sprintf("tf-pi-instance-%d", acctest.RandIntRange(10, 100))
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceActionWithHealthStatusConfig(name, power.Action_ResetState, power.Warning),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "status", strings.ToUpper(power.State_Active)),
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "health_status", power.Critical),
				),
			},
		},
	})
}

func TestAccIBMPIInstanceActionDumprestart(t *testing.T) {
	name := fmt.Sprintf("tf-pi-instance-%d", acctest.RandIntRange(10, 100))
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceActionWithHealthStatusConfig(name, power.Action_Dumprestart, power.Warning),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "status", strings.ToUpper(power.State_Active)),
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "health_status", power.Warning),
				),
			},
		},
	})
}

// Good path for a D-mode IPL start with CHAP. It needs a SHUTOFF instance that already has a
// network boot configuration staged through the instance operations API, which Terraform can't do,
// so it skips unless the environment provides one:
//
//	PI_NETBOOT_INSTANCE_ID   instance with a staged iSCSI netboot configuration, in SHUTOFF state
//	PI_NETBOOT_MAC_ADDRESS   MAC address of the network interface to boot from
//	PI_CHAP_NAME             CHAP name (1 to 32 ASCII characters)
//	PI_CHAP_PASSWORD         CHAP password (12 to 32 ASCII characters)
func TestAccIBMPIInstanceActionChapStart(t *testing.T) {
	instanceID := os.Getenv("PI_NETBOOT_INSTANCE_ID")
	macAddress := os.Getenv("PI_NETBOOT_MAC_ADDRESS")
	chapName := os.Getenv("PI_CHAP_NAME")
	chapPassword := os.Getenv("PI_CHAP_PASSWORD")
	if instanceID == "" || macAddress == "" || chapName == "" || chapPassword == "" {
		t.Skip("PI_NETBOOT_INSTANCE_ID, PI_NETBOOT_MAC_ADDRESS, PI_CHAP_NAME and PI_CHAP_PASSWORD must be set to run this test")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceActionChapStartConfig(instanceID, macAddress, chapName, chapPassword),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ibm_pi_instance_action.example", "status", strings.ToUpper(power.State_Active)),
				),
			},
		},
	})
}

func testAccCheckIBMPIInstanceActionChapStartConfig(instanceID, macAddress, chapName, chapPassword string) string {
	return fmt.Sprintf(`
	resource "ibm_pi_instance_action" "example" {
		pi_action            = "start"
		pi_chap_name         = "%[3]s"
		pi_chap_password     = "%[4]s"
		pi_cloud_instance_id = "%[5]s"
		pi_instance_id       = "%[1]s"
		pi_mac_address       = "%[2]s"
	}
	`, instanceID, macAddress, chapName, chapPassword, acc.Pi_cloud_instance_id)
}

func TestAccIBMPIInstanceActionChapInvalidAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckIBMPIInstanceActionChapInvalidActionConfig(power.Action_Stop, "chapuser", "chappassword12345"),
				ExpectError: regexp.MustCompile("can only be set when"),
			},
		},
	})
}

// No real instance is needed: the CHAP validation fails at plan time, before any API call.
func testAccCheckIBMPIInstanceActionChapInvalidActionConfig(action, chapName, chapPassword string) string {
	return fmt.Sprintf(`
	resource "ibm_pi_instance_action" "example" {
		pi_action            = "%[1]s"
		pi_chap_name         = "%[2]s"
		pi_chap_password       = "%[3]s"
		pi_cloud_instance_id = "00000000-0000-0000-0000-000000000000"
		pi_instance_id       = "00000000-0000-0000-0000-000000000000"
		pi_mac_address       = "fa:16:3e:00:00:01"
	}
	`, action, chapName, chapPassword)
}

func TestAccIBMPIInstanceActionChapMissingMacAddress(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckIBMPIInstanceActionChapMissingMacAddressConfig("chapuser", "chappassword12345"),
				ExpectError: regexp.MustCompile("require pi_mac_address to be set"),
			},
		},
	})
}

// No real instance is needed: the validation fails at plan time, before any API call.
func testAccCheckIBMPIInstanceActionChapMissingMacAddressConfig(chapName, chapPassword string) string {
	return fmt.Sprintf(`
	resource "ibm_pi_instance_action" "example" {
		pi_action            = "start"
		pi_chap_name         = "%[1]s"
		pi_chap_password       = "%[2]s"
		pi_cloud_instance_id = "00000000-0000-0000-0000-000000000000"
		pi_instance_id       = "00000000-0000-0000-0000-000000000000"
	}
	`, chapName, chapPassword)
}

func testAccCheckIBMPIInstanceActionConfig(name, action string) string {
	return fmt.Sprintf(`
	data "ibm_pi_image" "power_image" {
		pi_cloud_instance_id = "%[1]s"
		pi_image_name        = "%[4]s"
	  }
	  data "ibm_pi_network" "power_networks" {
		pi_cloud_instance_id = "%[1]s"
		pi_network_name      = "%[5]s"
	  }
	  resource "ibm_pi_instance" "power_instance" {
		pi_cloud_instance_id  = "%[1]s"
		pi_image_id           = data.ibm_pi_image.power_image.id
		pi_instance_name      = "%[2]s"
		pi_memory             = "2"
		pi_proc_type          = "shared"
		pi_processors         = "0.25"
		pi_storage_pool       = data.ibm_pi_image.power_image.storage_pool
		pi_storage_type       = "%[6]s"
		pi_sys_type           = "s922"
		pi_network {
			network_id = data.ibm_pi_network.power_networks.id
		}
	  }

	resource "ibm_pi_instance_action" "example" {
		pi_action				= "%[3]s"
		pi_cloud_instance_id	= "%[1]s"
		pi_instance_id			= resource.ibm_pi_instance.power_instance.pi_instance_name 
	}
	`, acc.Pi_cloud_instance_id, name, action, acc.Pi_image, acc.Pi_network_name, acc.PiStorageType)
}

func testAccCheckIBMPIInstanceActionWithHealthStatusConfig(name, action, instanceHealthStatus string) string {
	return fmt.Sprintf(`
	data "ibm_pi_image" "power_image" {
		pi_cloud_instance_id = "%[1]s"
		pi_image_name        = "%[4]s"
	  }
	  data "ibm_pi_network" "power_networks" {
		pi_cloud_instance_id = "%[1]s"
		pi_network_name      = "%[5]s"
	  }
	  resource "ibm_pi_instance" "power_instance" {
		pi_cloud_instance_id  = "%[1]s"
		pi_health_status      = "%[7]s"
		pi_image_id           = data.ibm_pi_image.power_image.id
		pi_instance_name      = "%[2]s"
		pi_memory             = "2"
		pi_proc_type          = "shared"
		pi_processors         = "0.25"
		pi_storage_pool       = data.ibm_pi_image.power_image.storage_pool
		pi_storage_type       = "%[6]s"
		pi_sys_type           = "s922"
		pi_network {
			network_id = data.ibm_pi_network.power_networks.id
		}
	  }
	resource "ibm_pi_instance_action" "example" {
		pi_action				= "%[3]s"
		pi_cloud_instance_id	= "%[1]s"
		pi_health_status		= "%[7]s"
		pi_instance_id			= resource.ibm_pi_instance.power_instance.pi_instance_name 
	}
	`, acc.Pi_cloud_instance_id, name, action, acc.Pi_image, acc.Pi_network_name, acc.PiStorageType, instanceHealthStatus)
}
