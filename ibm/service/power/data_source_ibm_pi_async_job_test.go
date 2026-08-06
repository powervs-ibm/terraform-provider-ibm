// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPIAsyncJobDataSource_basic(t *testing.T) {
	asyncJobResData := "data.ibm_pi_async_job.testacc_ds_async_job"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIAsyncJobDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(asyncJobResData, "id"),
					resource.TestCheckResourceAttrSet(asyncJobResData, "action"),
					resource.TestCheckResourceAttrSet(asyncJobResData, "status"),
					resource.TestCheckResourceAttrSet(asyncJobResData, "resource_id"),
					resource.TestCheckResourceAttrSet(asyncJobResData, "resource_type"),
				),
			},
		},
	})
}

func testAccCheckIBMPIAsyncJobDataSourceConfig() string {
	return fmt.Sprintf(`
		data "ibm_pi_async_job" "testacc_ds_async_job" {
			pi_cloud_instance_id = "%s"
			pi_async_job_id      = "%s"
		}`, acc.Pi_cloud_instance_id, acc.Pi_async_job_id)
}
