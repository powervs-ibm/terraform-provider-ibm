// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Description:  "The GUID of the service instance associated with an account.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_NetworkName: {
				Description:  "Network Name - This is the subnet name  in the Cloud instance.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_NetworkPortDescription: {
				Default:     "Port Created via Terraform",
				Description: "A human readable description for this network Port.",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_NetworkPortIPAddress: {
				Description: "The requested ip address of this port.",
				Computed:    true,
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_PVMInstanceId: {
				Description:  "Instance id to attach the network port to.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_MacAddress: {
				Description: "The MAC address of the port.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			Attr_NetworkPortID: {
				Description: "The ID of the port.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			Attr_PublicIP: {
				Description: " The public IP associated with the port.",
				Computed:    true,
				Type:        schema.TypeString,
			},
			Attr_Status: {
				Description: "The status of the port.",
				Computed:    true,
				Type:        schema.TypeString,
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
	networkname := d.Get(Arg_NetworkName).(string)
	instanceID := d.Get(Arg_PVMInstanceId).(string)
	description := d.Get(Arg_NetworkPortDescription).(string)
	nwportBody := &models.NetworkPortCreate{Description: description}

	if v, ok := d.GetOk(Arg_NetworkPortIPAddress); ok {
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

	d.Set(Arg_NetworkPortIPAddress, networkdata.IPAddress)
	d.Set(Arg_NetworkPortDescription, networkdata.Description)
	d.Set(Arg_PVMInstanceId, networkdata.PvmInstance.PvmInstanceID)
	d.Set(Attr_MacAddress, networkdata.MacAddress)
	d.Set(Attr_Status, networkdata.Status)
	d.Set(Attr_NetworkPortID, networkdata.PortID)
	d.Set(Attr_PublicIP, networkdata.ExternalIP)

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

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Retry, State_Build},
		Target:     []string{State_Down},
		Refresh:    isIBMPINetworkportRefreshFunc(client, id, networkname),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Minute,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isIBMPINetworkportRefreshFunc(client *instance.IBMPINetworkClient, id, networkname string) retry.StateRefreshFunc {

	log.Printf("Calling the IsIBMPINetwork Refresh Function....with the following id (%s) for network port and following id (%s) for network name and waiting for network to be READY", id, networkname)
	return func() (interface{}, string, error) {
		network, err := client.GetPort(networkname, id)
		if err != nil {
			return nil, "", err
		}

		if strings.ToLower(*network.Status) == State_Down {
			log.Printf(" The port has been created with the following ip address and attached to an instance ")
			return network, State_Down, nil
		}

		return network, State_Build, nil
	}
}
func isWaitForIBMPINetworkPortAttachAvailable(ctx context.Context, client *instance.IBMPINetworkClient, id, networkname, instanceid string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for Power Network (%s) that was created for Network Zone (%s) to be available.", id, networkname)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Retry, State_Build},
		Target:     []string{State_Active},
		Refresh:    isIBMPINetworkPortAttachRefreshFunc(client, id, networkname, instanceid),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Minute,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isIBMPINetworkPortAttachRefreshFunc(client *instance.IBMPINetworkClient, id, networkname, instanceid string) retry.StateRefreshFunc {

	log.Printf("Calling the IsIBMPINetwork Refresh Function....with the following id (%s) for network port and following id (%s) for network name and waiting for network to be READY", id, networkname)
	return func() (interface{}, string, error) {
		network, err := client.GetPort(networkname, id)
		if err != nil {
			return nil, "", err
		}

		if strings.ToLower(*network.Status) == State_Active && network.PvmInstance.PvmInstanceID == instanceid {
			log.Printf(" The port has been created with the following ip address and attached to an instance ")
			return network, State_Active, nil
		}

		return network, State_Build, nil
	}
}
