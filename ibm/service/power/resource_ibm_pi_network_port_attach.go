// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIBMPINetworkPortAttach() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceIBMPINetworkPortAttachCreate,
		ReadContext:   resourceIBMPINetworkPortAttachRead,
		DeleteContext: resourceIBMPINetworkPortAttachDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			Arg_CloudInstanceID: {
				ForceNew: true,
				Required: true,
				Type:     schema.TypeString,
			},
			PIInstanceId: {
				Description: "Instance id to attach the network port to",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			PINetworkName: {
				Description: "Network Name - This is the subnet name  in the Cloud instance",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			PINetworkPortDescription: {
				Description: "A human readable description for this network Port",
				Default:     "Port Created via Terraform",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			PINetworkPortIPAddress: {
				Computed: true,
				ForceNew: true,
				Optional: true,
				Type:     schema.TypeString,
			},

			//Computed Attributes
			Attr_MacAddress: {
				Computed: true,
				Type:     schema.TypeString,
			},
			Attr_NetworkPortID: {
				Computed: true,
				Type:     schema.TypeString,
			},
			Attr_Status: {
				Computed: true,
				Type:     schema.TypeString,
			},
			Attr_PublicIP: {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}

}

func resourceIBMPINetworkPortAttachCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	networkname := d.Get(PINetworkName).(string)
	instanceID := d.Get(PIInstanceId).(string)
	description := d.Get(PINetworkPortDescription).(string)
	nwportBody := &models.NetworkPortCreate{Description: description}

	if v, ok := d.GetOk(PINetworkPortIPAddress); ok {
		ipaddress := v.(string)
		nwportBody.IPAddress = ipaddress
	}

	nwportattachBody := &models.NetworkPortUpdate{
		Description:   &description,
		PvmInstanceID: &instanceID,
	}

	client := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)

	networkPortResponse, err := client.CreatePort(networkname, nwportBody)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Printing the networkresponse %+v", &networkPortResponse)

	networkPortID := *networkPortResponse.PortID

	_, err = isWaitForIBMPINetworkportAvailable(ctx, client, networkPortID, networkname, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	networkPortResponse, err = client.UpdatePort(networkname, networkPortID, nwportattachBody)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = isWaitForIBMPINetworkPortAttachAvailable(ctx, client, networkPortID, networkname, instanceID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", cloudInstanceID, networkname, networkPortID))

	return resourceIBMPINetworkPortAttachRead(ctx, d, meta)
}

func resourceIBMPINetworkPortAttachRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := parts[0]
	networkname := parts[1]
	portID := parts[2]

	networkC := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)
	networkdata, err := networkC.GetPort(networkname, portID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set(PINetworkPortIPAddress, networkdata.IPAddress)
	d.Set(PINetworkPortDescription, networkdata.Description)
	d.Set(PIInstanceId, networkdata.PvmInstance.PvmInstanceID)
	d.Set(Attr_MacAddress, networkdata.MacAddress)
	d.Set(Attr_Status, networkdata.Status)
	d.Set("network_port_id", networkdata.PortID)
	d.Set("public_ip", networkdata.ExternalIP)

	return nil
}

func resourceIBMPINetworkPortAttachDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	log.Printf("Calling the network delete functions. ")
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := parts[0]
	networkname := parts[1]
	portID := parts[2]

	client := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)

	log.Printf("Calling the delete with the following params delete with cloud instance (%s) and networkid (%s) and portid (%s) ", cloudInstanceID, networkname, portID)
	err = client.DeletePort(networkname, portID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func isWaitForIBMPINetworkportAvailable(ctx context.Context, client *instance.IBMPINetworkClient, id string, networkname string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for Power Network (%s) that was created for Network Zone (%s) to be available.", id, networkname)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", PINetworkProvisioning},
		Target:     []string{"DOWN"},
		Refresh:    isIBMPINetworkportRefreshFunc(client, id, networkname),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Minute,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isIBMPINetworkportRefreshFunc(client *instance.IBMPINetworkClient, id, networkname string) resource.StateRefreshFunc {

	log.Printf("Calling the IsIBMPINetwork Refresh Function....with the following id (%s) for network port and following id (%s) for network name and waiting for network to be READY", id, networkname)
	return func() (interface{}, string, error) {
		network, err := client.GetPort(networkname, id)
		if err != nil {
			return nil, "", err
		}

		if *network.Status == "DOWN" {
			log.Printf(" The port has been created with the following ip address and attached to an instance ")
			return network, "DOWN", nil
		}

		return network, PINetworkProvisioning, nil
	}
}
func isWaitForIBMPINetworkPortAttachAvailable(ctx context.Context, client *instance.IBMPINetworkClient, id, networkname, instanceid string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for Power Network (%s) that was created for Network Zone (%s) to be available.", id, networkname)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", PINetworkProvisioning},
		Target:     []string{"ACTIVE"},
		Refresh:    isIBMPINetworkPortAttachRefreshFunc(client, id, networkname, instanceid),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Minute,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isIBMPINetworkPortAttachRefreshFunc(client *instance.IBMPINetworkClient, id, networkname, instanceid string) resource.StateRefreshFunc {

	log.Printf("Calling the IsIBMPINetwork Refresh Function....with the following id (%s) for network port and following id (%s) for network name and waiting for network to be READY", id, networkname)
	return func() (interface{}, string, error) {
		network, err := client.GetPort(networkname, id)
		if err != nil {
			return nil, "", err
		}

		if *network.Status == "ACTIVE" && network.PvmInstance.PvmInstanceID == instanceid {
			log.Printf(" The port has been created with the following ip address and attached to an instance ")
			return network, "ACTIVE", nil
		}

		return network, PINetworkProvisioning, nil
	}
}
