// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/service/power"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccIBMPIInstanceNetworkAttachBasic(t *testing.T) {
	name := fmt.Sprintf("tf-pi-instance-network-attach-%d", acctest.RandIntRange(10, 100))
	attachRes := "ibm_pi_instance_network_attach.network_attach"
	instanceRes := "ibm_pi_instance.power_instance"

	attachNetwork := acc.Pi_network_id
	if attachNetwork == "" || attachNetwork == acc.Pi_network_name || attachNetwork == "terraform-test-power" {
		attachNetwork = "3242c9a0-8ee9-4c1e-902d-de05993d6aa8"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMPIInstanceNetworkAttachDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceNetworkAttachConfig(name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPIInstanceNetworkAttachExists(attachRes),
					resource.TestCheckResourceAttrSet(attachRes, "id"),
					resource.TestCheckResourceAttrSet(attachRes, "ip_address"),
					resource.TestCheckResourceAttrSet(attachRes, "mac_address"),
					resource.TestCheckResourceAttrSet(attachRes, "network_name"),
					resource.TestCheckResourceAttrSet(attachRes, "type"),
				),
			},
			{
				// Detach the network by removing the attach resource, the instance stays running.
				Config: testAccCheckIBMPIInstanceNetworkAttachConfig(name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPIInstanceNetworkDetached(instanceRes, attachNetwork),
					testAccCheckIBMPIInstanceActive(instanceRes),
				),
			},
			{
				// Re-attach so the final destroy runs the attachment's delete path while the instance still exists.
				Config: testAccCheckIBMPIInstanceNetworkAttachConfig(name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPIInstanceNetworkAttachExists(attachRes),
				),
			},
			{
				ResourceName:      attachRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIBMPIInstanceNetworkAttachConfig(name string, attach bool) string {
	imageID := acc.Pi_image
	if imageID == "" || imageID == "c93dc4c6-e85a-4da2-9ea6-f24576256122" {
		if acc.Pi_image_id != "" && acc.Pi_image_id != "IBMi-72-09-2924-11" {
			imageID = acc.Pi_image_id
		}
	}

	primaryNetwork := acc.Pi_network_name
	if primaryNetwork == "" || primaryNetwork == "terraform-test-power" {
		primaryNetwork = "9412c41e-d80f-4258-b686-5c41b000b4c0"
	}

	attachNetwork := acc.Pi_network_id
	if attachNetwork == "" || attachNetwork == acc.Pi_network_name || attachNetwork == "terraform-test-power" {
		attachNetwork = "3242c9a0-8ee9-4c1e-902d-de05993d6aa8"
	}

	config := fmt.Sprintf(`
	resource "ibm_pi_instance" "power_instance" {
		pi_cloud_instance_id = "%[1]s"
		pi_image_id          = "%[3]s"
		pi_instance_name     = "%[2]s"
		pi_memory            = "2"
		pi_proc_type         = "shared"
		pi_processors        = "0.25"
		pi_storage_type      = "%[6]s"
		pi_sys_type          = "s922"
		pi_network {
			network_id = "%[4]s"
		}
	}`, acc.Pi_cloud_instance_id, name, imageID, primaryNetwork, attachNetwork, acc.PiStorageType)

	if attach {
		config += fmt.Sprintf(`
	resource "ibm_pi_instance_network_attach" "network_attach" {
		pi_cloud_instance_id = ibm_pi_instance.power_instance.pi_cloud_instance_id
		pi_instance_id       = ibm_pi_instance.power_instance.instance_id
		pi_network_id        = "%s"
	}`, attachNetwork)
	}
	return config
}

func testAccCheckIBMPIInstanceNetworkAttachExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
		if err != nil {
			return err
		}
		parts, err := flex.IdParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		networkClient := instance.NewIBMPIInstanceNetworksClient(context.Background(), sess, parts[0])

		networks, err := networkClient.GetAll(parts[1])
		if err != nil {
			return err
		}
		for _, net := range networks.Networks {
			if net != nil && net.NetworkID == parts[2] {
				return nil
			}
		}
		return fmt.Errorf("network %s is not attached to instance %s", parts[2], parts[1])
	}
}

// testAccCheckIBMPIInstanceNetworkDetached verifies with direct API calls that the network is no longer attached to the instance.
func testAccCheckIBMPIInstanceNetworkDetached(instanceName, networkID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceRs, ok := s.RootModule().Resources[instanceName]
		if !ok {
			return fmt.Errorf("Not found: %s", instanceName)
		}

		sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
		if err != nil {
			return err
		}
		cloudInstanceID := instanceRs.Primary.Attributes["pi_cloud_instance_id"]
		instanceID := instanceRs.Primary.Attributes["instance_id"]

		return testAccCheckIBMPIInstanceNetworkGone(sess, cloudInstanceID, instanceID, networkID)
	}
}

func testAccCheckIBMPIInstanceActive(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
		if err != nil {
			return err
		}
		client := instance.NewIBMPIInstanceClient(context.Background(), sess, rs.Primary.Attributes["pi_cloud_instance_id"])
		pvm, err := client.Get(rs.Primary.Attributes["instance_id"])
		if err != nil {
			return err
		}
		if pvm.Status == nil || strings.ToLower(*pvm.Status) != power.State_Active {
			return fmt.Errorf("instance %s is not active after the detach", rs.Primary.Attributes["instance_id"])
		}
		return nil
	}
}

func testAccCheckIBMPIInstanceNetworkAttachDestroy(s *terraform.State) error {
	sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ibm_pi_instance_network_attach" {
			continue
		}
		parts, err := flex.IdParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		if err := testAccCheckIBMPIInstanceNetworkGone(sess, parts[0], parts[1], parts[2]); err != nil {
			return err
		}
	}
	return nil
}

// testAccCheckIBMPIInstanceNetworkGone returns an error if the instance still has an interface on the network.
// The API returns not found when the instance, or the network on the instance, no longer exists.
func testAccCheckIBMPIInstanceNetworkGone(sess *ibmpisession.IBMPISession, cloudInstanceID, instanceID, networkID string) error {
	networkClient := instance.NewIBMPIInstanceNetworksClient(context.Background(), sess, cloudInstanceID)
	networks, err := networkClient.GetAll(instanceID)
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, power.NotFound) || strings.Contains(msg, "notfound") || strings.Contains(msg, "[404]") {
			return nil
		}
		return err
	}
	for _, n := range networks.Networks {
		if n != nil && n.NetworkID == networkID {
			return fmt.Errorf("network %s is still attached to instance %s", networkID, instanceID)
		}
	}
	return nil
}
