// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package vpc_test

import (
	"fmt"
	"strings"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMISBMSDataSource_basic(t *testing.T) {
	var server string
	resName := "data.ibm_is_bare_metal_server.test1"
	vpcname := fmt.Sprintf("tf-vpc-%d", acctest.RandIntRange(10, 100))
	name := fmt.Sprintf("tf-server-%d", acctest.RandIntRange(10, 100))
	subnetname := fmt.Sprintf("tfip-subnet-%d", acctest.RandIntRange(10, 100))
	publicKey := strings.TrimSpace(`
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCKVmnMOlHKcZK8tpt3MP1lqOLAcqcJzhsvJcjscgVERRN7/9484SOBJ3HSKxxNG5JN8owAjy5f9yYwcUg+JaUVuytn5Pv3aeYROHGGg+5G346xaq3DAwX6Y5ykr2fvjObgncQBnuU5KHWCECO/4h8uWuwh/kfniXPVjFToc+gnkqA+3RKpAecZhFXwfalQ9mMuYGFxn+fwn8cYEApsJbsEmb0iJwPiZ5hjFC8wREuiTlhPHDgkBLOiycd20op2nXzDbHfCHInquEe/gYxEitALONxm0swBOwJZwlTDOB7C6y2dzlrtxr1L59m7pCkWI4EtTRLvleehBoj3u7jB4usR
`)
	sshname := fmt.Sprintf("tf-sshname-%d", acctest.RandIntRange(10, 100))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckIBMISBMSDataSourceConfig(vpcname, subnetname, sshname, publicKey, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMISBareMetalServerExists("ibm_is_bare_metal_server.testacc_bms", server),
					resource.TestCheckResourceAttr(
						resName, "name", name),
					resource.TestCheckResourceAttr(
						"data.ibm_is_bare_metal_server.test1", "zone", acc.ISZoneName),
					resource.TestCheckResourceAttr(
						"data.ibm_is_bare_metal_server.test1", "profile", acc.IsBareMetalServerProfileName),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.address"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.name"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.href"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.reserved_ip"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.resource_type"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.port_speed"),
				),
			},
		},
	})
}

func TestAccIBMISBMSDataSource_firmwareUpdate(t *testing.T) {
	var server string
	resName := "data.ibm_is_bare_metal_server.test1"
	vpcname := fmt.Sprintf("tf-vpc-%d", acctest.RandIntRange(10, 100))
	name := fmt.Sprintf("tf-server-%d", acctest.RandIntRange(10, 100))
	subnetname := fmt.Sprintf("tfip-subnet-%d", acctest.RandIntRange(10, 100))
	publicKey := strings.TrimSpace(`
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCKVmnMOlHKcZK8tpt3MP1lqOLAcqcJzhsvJcjscgVERRN7/9484SOBJ3HSKxxNG5JN8owAjy5f9yYwcUg+JaUVuytn5Pv3aeYROHGGg+5G346xaq3DAwX6Y5ykr2fvjObgncQBnuU5KHWCECO/4h8uWuwh/kfniXPVjFToc+gnkqA+3RKpAecZhFXwfalQ9mMuYGFxn+fwn8cYEApsJbsEmb0iJwPiZ5hjFC8wREuiTlhPHDgkBLOiycd20op2nXzDbHfCHInquEe/gYxEitALONxm0swBOwJZwlTDOB7C6y2dzlrtxr1L59m7pCkWI4EtTRLvleehBoj3u7jB4usR
`)
	sshname := fmt.Sprintf("tf-sshname-%d", acctest.RandIntRange(10, 100))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckIBMISBMSDataSourceConfig(vpcname, subnetname, sshname, publicKey, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMISBareMetalServerExists("ibm_is_bare_metal_server.testacc_bms", server),
					resource.TestCheckResourceAttr(
						resName, "name", name),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "firmware_update_type_available"),
				),
			},
		},
	})
}
func TestAccIBMISBMSDataSourceVNI_basic(t *testing.T) {
	var server string
	resName := "data.ibm_is_bare_metal_server.test1"
	vpcname := fmt.Sprintf("tf-vpc-%d", acctest.RandIntRange(10, 100))
	name := fmt.Sprintf("tf-server-%d", acctest.RandIntRange(10, 100))
	subnetname := fmt.Sprintf("tfip-subnet-%d", acctest.RandIntRange(10, 100))
	publicKey := strings.TrimSpace(`
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCKVmnMOlHKcZK8tpt3MP1lqOLAcqcJzhsvJcjscgVERRN7/9484SOBJ3HSKxxNG5JN8owAjy5f9yYwcUg+JaUVuytn5Pv3aeYROHGGg+5G346xaq3DAwX6Y5ykr2fvjObgncQBnuU5KHWCECO/4h8uWuwh/kfniXPVjFToc+gnkqA+3RKpAecZhFXwfalQ9mMuYGFxn+fwn8cYEApsJbsEmb0iJwPiZ5hjFC8wREuiTlhPHDgkBLOiycd20op2nXzDbHfCHInquEe/gYxEitALONxm0swBOwJZwlTDOB7C6y2dzlrtxr1L59m7pCkWI4EtTRLvleehBoj3u7jB4usR
`)
	sshname := fmt.Sprintf("tf-sshname-%d", acctest.RandIntRange(10, 100))
	vniname := fmt.Sprintf("tf-vni-%d", acctest.RandIntRange(10, 100))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckIBMISBMSDataSourceVniConfig(vpcname, subnetname, sshname, publicKey, vniname, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMISBareMetalServerExists("ibm_is_bare_metal_server.testacc_bms", server),
					resource.TestCheckResourceAttr(
						resName, "name", name),
					resource.TestCheckResourceAttr(
						"data.ibm_is_bare_metal_server.test1", "zone", acc.ISZoneName),
					resource.TestCheckResourceAttr(
						"data.ibm_is_bare_metal_server.test1", "profile", acc.IsBareMetalServerProfileName),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.address"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.name"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.href"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.reserved_ip"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.primary_ip.0.resource_type"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_interface.0.port_speed"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_attachment.#"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_attachment.0.href"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_attachment.0.id"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_attachment.0.name"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_attachment.0.primary_ip.#"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_attachment.0.resource_type"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_attachment.0.subnet.#"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "primary_network_attachment.0.virtual_network_interface.#"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "network_attachments.#"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "network_attachments.0.href"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "network_attachments.0.id"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "network_attachments.0.name"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "network_attachments.0.primary_ip.#"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "network_attachments.0.resource_type"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "network_attachments.0.subnet.#"),
					resource.TestCheckResourceAttrSet("data.ibm_is_bare_metal_server.test1", "network_attachments.0.virtual_network_interface.#"),
				),
			},
		},
	})
}

func testAccCheckIBMISBMSDataSourceConfig(vpcname, subnetname, sshname, publicKey, name string) string {
	// status filter defaults to empty
	return testAccCheckIBMISBareMetalServerConfig(vpcname, subnetname, sshname, publicKey, name) + fmt.Sprintf(`
      data "ibm_is_bare_metal_server" "test1" {
		  name =	"%s"
      }`, name)
}
func testAccCheckIBMISBMSDataSourceVniConfig(vpcname, subnetname, sshname, publicKey, vniname, name string) string {
	// status filter defaults to empty
	return testAccCheckIBMISBareMetalServerVNIConfig(vpcname, subnetname, sshname, publicKey, vniname, name) + fmt.Sprintf(`
      data "ibm_is_bare_metal_server" "test1" {
		  depends_on	= [ ibm_is_bare_metal_server.testacc_bms ]
		  name 			=	"%s"
      }`, name)
}
