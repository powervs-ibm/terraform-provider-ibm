// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/service/power"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccIBMPIInstanceSnapshotV2Basic(t *testing.T) {
	instanceName := fmt.Sprintf("tf-pi-instance-snapshot-v2-%d", acctest.RandIntRange(10, 100))
	snapshotName := fmt.Sprintf("tf-pi-snapshot-v2-%d", acctest.RandIntRange(10, 100))
	snapshotNameUpdated := fmt.Sprintf("%s-upd", instanceName)
	snapshotRes := "ibm_pi_instance_snapshot_v2.testacc_snapshot_v2"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.TestAccPreCheck(t) },
		Providers:    acc.TestAccProviders,
		CheckDestroy: testAccCheckIBMPIInstanceSnapshotV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIInstanceSnapshotV2Config(instanceName, snapshotName, power.OK),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPIInstanceSnapshotV2Exists(snapshotRes),
					resource.TestCheckResourceAttr(snapshotRes, "pi_snapshot_name", snapshotName),
					resource.TestCheckResourceAttr(snapshotRes, "status", power.State_Available),
					resource.TestCheckResourceAttrSet(snapshotRes, "id"),
					resource.TestCheckResourceAttrSet(snapshotRes, "snapshot_id"),
					resource.TestCheckResourceAttrSet(snapshotRes, "type"),
				),
			},
			{
				Config: testAccCheckIBMPIInstanceSnapshotV2Config(instanceName, snapshotNameUpdated, power.OK),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIBMPIInstanceSnapshotV2Exists(snapshotRes),
					resource.TestCheckResourceAttr(snapshotRes, "pi_snapshot_name", snapshotNameUpdated),
					resource.TestCheckResourceAttr(snapshotRes, "status", power.State_Available),
				),
			},
		},
	})
}

func testAccCheckIBMPIInstanceSnapshotV2Destroy(s *terraform.State) error {
	sess, err := acc.TestAccProvider.Meta().(conns.ClientSession).IBMPISession()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ibm_pi_instance_snapshot_v2" {
			continue
		}
		cloudInstanceID, snapshotID, err := splitID(rs.Primary.ID)
		if err != nil {
			return err
		}
		client := instance.NewIBMPISnapshotClient(context.Background(), sess, cloudInstanceID)
		_, err = client.V2Get(snapshotID)
		if err == nil {
			return fmt.Errorf("PI Instance Snapshot V2 still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckIBMPIInstanceSnapshotV2Exists(n string) resource.TestCheckFunc {
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
		cloudInstanceID, snapshotID, err := splitID(rs.Primary.ID)
		if err != nil {
			return err
		}
		client := instance.NewIBMPISnapshotClient(context.Background(), sess, cloudInstanceID)
		_, err = client.V2Get(snapshotID)
		return err
	}
}

func testAccCheckIBMPIInstanceSnapshotV2Config(instanceName, snapshotName, healthStatus string) string {
	return testAccCheckIBMPIInstanceConfig(instanceName, healthStatus) + fmt.Sprintf(`
		resource "ibm_pi_instance_snapshot_v2" "testacc_snapshot_v2" {
			depends_on           = [ibm_pi_instance.power_instance]
			pi_cloud_instance_id = "%s"
			pi_pvm_instance_id   = ibm_pi_instance.power_instance.instance_id
			pi_snapshot_name     = "%s"
		}`, acc.Pi_cloud_instance_id, snapshotName)
}
