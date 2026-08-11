// Copyright IBM Corp. 2024 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power_test

import (
	"fmt"
	"testing"

	acc "github.com/IBM-Cloud/terraform-provider-ibm/ibm/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIBMPIAsyncJobsDataSource_basic(t *testing.T) {
	asyncJobsResData := "data.ibm_pi_async_jobs.testacc_ds_async_jobs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIBMPIAsyncJobsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(asyncJobsResData, "id"),
				),
			},
		},
	})
}

func testAccCheckIBMPIAsyncJobsDataSourceConfig() string {
	return fmt.Sprintf(`
		data "ibm_pi_async_jobs" "testacc_ds_async_jobs" {
			pi_cloud_instance_id = "%s"
		}`, acc.Pi_cloud_instance_id)
}
