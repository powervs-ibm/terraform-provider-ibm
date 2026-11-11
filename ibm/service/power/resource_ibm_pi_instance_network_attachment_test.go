// Copyright IBM Corp. 2026 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccIBMPIInstanceNetworkAttachmentBasic(t *testing.T) {
	attachmentRes := "ibm_pi_instance_network_attachment.example"
	instanceRes := "ibm_pi_instance.power_instance"
	name := fmt.Sprintf("tf-pi-instnetattachment-%d", acctest.RandIntRange(10, 100))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMPIInstanceNetworkAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceNetworkAttachmentConfig(name, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPIInstanceExists(instanceRes),
					testAccCheckIBMPIInstanceNetworkAttachmentExists(attachmentRes),
					resource.TestCheckResourceAttr(attachmentRes, "vpc_virtual_network_interface.0.bandwidth", "10"),
					resource.TestCheckResourceAttrSet(attachmentRes, "vpc_virtual_network_interface.0.ip_address"),
					resource.TestCheckResourceAttrSet(attachmentRes, "vpc_virtual_network_interface.0.mac_address"),
					resource.TestCheckResourceAttrSet(attachmentRes, "network_attachment_id"),
				),
			},
			{
				// Bandwidth is updatable in place via PUT, so this step should not
				// force recreation of the network attachment.
				Config: testAccCheckIBMPIInstanceNetworkAttachmentConfig(name, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPIInstanceExists(instanceRes),
					testAccCheckIBMPIInstanceNetworkAttachmentExists(attachmentRes),
					resource.TestCheckResourceAttr(attachmentRes, "vpc_virtual_network_interface.0.bandwidth", "20"),
				),
			},
		},
	})
}

func testAccCheckIBMPIInstanceNetworkAttachmentConfig(name string, bandwidth int) string {
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
			bandwidth = %[6]d
		}
	}
	`, acc.Pi_cloud_instance_id, name, acc.Pi_image, acc.Pi_network_id, acc.Pi_external_network_interface_crn, bandwidth)
}

// testAccCheckIBMPIInstanceNetworkAttachmentExists confirms, via a direct API call,
// that the network attachment represented by the resource's own state ID exists.
func testAccCheckIBMPIInstanceNetworkAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("No Record ID is set")
		}

		parts, err := flex.IdParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		cloudInstanceID := parts[0]
		instanceID := parts[1]
		networkAttachmentID := parts[2]

		sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
		if err != nil {
			return err
		}
		client := instance.NewIBMPIInstanceClient(context.Background(), sess, cloudInstanceID)
		_, err = client.GetNetworkAttachment(instanceID, networkAttachmentID)
		if err != nil {
			return err
		}
		return nil
	}
}

// testAccCheckIBMPIInstanceNetworkAttachmentDestroy runs as part of CheckDestroy, after
// Terraform has destroyed everything, and confirms every ibm_pi_instance_network_attachment
// resource that was in state is really gone from the backend.
func testAccCheckIBMPIInstanceNetworkAttachmentDestroy(s *terraform.State) error {
	sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ibm_pi_instance_network_attachment" {
			continue
		}

		parts, err := flex.IdParts(rs.Primary.ID)
		if err != nil {
			return err
		}
		cloudInstanceID := parts[0]
		instanceID := parts[1]
		networkAttachmentID := parts[2]

		client := instance.NewIBMPIInstanceClient(context.Background(), sess, cloudInstanceID)
		_, err = client.GetNetworkAttachment(instanceID, networkAttachmentID)
		if err == nil {
			return fmt.Errorf("PI Instance Network attachment still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}
