// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/service/power"
)

func TestAccIBMPINetworkSecurityGroupRuleBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPINetworkSecurityGroupRuleConfigRemoveRule(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIBMPINetworkSecurityGroupRuleExists("ibm_pi_network_security_group_rule.network_security_group_rule"),
					resource.TestCheckResourceAttrSet("ibm_pi_network_security_group_rule.network_security_group_member_rule", power.Arg_NetworkSecurityGroupID),
				),
			},
		},
	})
}

func testAccCheckIBMPINetworkSecurityGroupRuleConfigRemoveRule() string {
	return fmt.Sprintf(`
		resource "ibm_pi_network_security_group_member_rule" "network_security_group_member_rule" {
			pi_cloud_instance_id = "%s"
			pi_network_security_group_id = "%s"
		}`, acc.Pi_cloud_instance_id, acc.Pi_network_security_group_id)
}

func testAccCheckIBMPINetworkSecurityGroupRuleExists(n string) resource.TestCheckFunc {
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
		cloudInstanceID, nsgID, err := splitID(rs.Primary.ID)
		if err != nil {
			return err
		}
		nsgClient := instance.NewIBMIPINetworkSecurityGroupClient(context.Background(), sess, cloudInstanceID)
		_, err = nsgClient.Get(nsgID)
		if err != nil {
			return err
		}
		return nil
	}
}
