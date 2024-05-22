// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
)

func ResourceIBMPINetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPINetworkCreate,
		ReadContext:   resourceIBMPINetworkRead,
		UpdateContext: resourceIBMPINetworkUpdate,
		DeleteContext: resourceIBMPINetworkDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_Cidr: {
				Computed:    true,
				Description: "PI network CIDR",
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_CloudInstanceID: {
				Description:  "PI cloud instance ID",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_DNS: {
				Computed:    true,
				Description: "List of PI network DNS name",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Type:        schema.TypeSet,
			},
			Arg_Gateway: {
				Computed:    true,
				Description: "PI network gateway",
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_IPAddressRange: {
				Computed:    true,
				Description: "List of one or more ip address range(s)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						Arg_EndingIPAddress: {
							Description:  "Ending ip address",
							Required:     true,
							Type:         schema.TypeString,
							ValidateFunc: validation.NoZeroValues,
						},
						Arg_StartingIPAddress: {
							Description:  "Starting ip address",
							Required:     true,
							Type:         schema.TypeString,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
				Optional: true,
				Type:     schema.TypeList,
			},
			Arg_NetworkAccessConfig: {
				Computed:     true,
				Description:  "PI network communication configuration",
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{Internal_Only, Outbound_Only, Bidirectional_Static_Route, Bidirectional_BGP, Bidirectional_L2Out}),
			},
			Arg_NetworkJumbo: {
				Computed:      true,
				ConflictsWith: []string{Arg_NetworkMTU},
				Deprecated:    "This field is deprecated, use pi_network_mtu instead.",
				Description:   "PI network enable MTU Jumbo option",
				Optional:      true,
				Type:          schema.TypeBool,
			},
			Arg_NetworkMTU: {
				Computed:      true,
				ConflictsWith: []string{Arg_NetworkJumbo},
				Description:   "PI Maximum Transmission Unit",
				Optional:      true,
				Type:          schema.TypeInt,
			},
			Arg_NetworkName: {
				Description:  "PI network name",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			Arg_NetworkType: {
				Description:  "PI network type",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{VLAN, Pub_VLAN}),
			},

			// Attributes
			Attr_NetworkID: {
				Computed:    true,
				Description: "PI network ID",
				Type:        schema.TypeString,
			},
			Attr_VLanID: {
				Computed:    true,
				Description: "VLAN Id value",
				Type:        schema.TypeFloat,
			},
		},
	}
}

func resourceIBMPINetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}
	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	networkname := d.Get(Arg_NetworkName).(string)
	networktype := d.Get(Arg_NetworkType).(string)

	client := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)
	var body = &models.NetworkCreate{
		Type: &networktype,
		Name: networkname,
	}
	if v, ok := d.GetOk(Arg_DNS); ok {
		networkdns := flex.ExpandStringList((v.(*schema.Set)).List())
		if len(networkdns) > 0 {
			body.DNSServers = networkdns
		}
	}

	if v, ok := d.GetOk(Arg_NetworkJumbo); ok {
		body.Jumbo = v.(bool)
	}
	if v, ok := d.GetOk(Arg_NetworkMTU); ok {
		var mtu int64 = int64(v.(int))
		body.Mtu = &mtu
	}
	if v, ok := d.GetOk(Arg_NetworkAccessConfig); ok {
		body.AccessConfig = models.AccessConfig(v.(string))
	}

	if networktype == VLAN {
		var networkcidr string
		var ipBodyRanges []*models.IPAddressRange
		if v, ok := d.GetOk(Arg_Cidr); ok {
			networkcidr = v.(string)
		} else {
			return diag.Errorf("%s is required when %s is vlan", Arg_Cidr, Arg_NetworkType)
		}

		gateway, firstip, lastip, err := generateIPData(networkcidr)
		if err != nil {
			return diag.FromErr(err)
		}

		ipBodyRanges = []*models.IPAddressRange{{EndingIPAddress: &lastip, StartingIPAddress: &firstip}}

		if g, ok := d.GetOk(Arg_Gateway); ok {
			gateway = g.(string)
		}

		if ips, ok := d.GetOk(Arg_IPAddressRange); ok {
			ipBodyRanges = getIPAddressRanges(ips.([]interface{}))
		}

		body.IPAddressRanges = ipBodyRanges
		body.Gateway = gateway
		body.Cidr = networkcidr
	}

	networkResponse, err := client.Create(body)
	if err != nil {
		return diag.FromErr(err)
	}

	networkID := *networkResponse.NetworkID

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, networkID))

	_, err = isWaitForIBMPINetworkAvailable(ctx, client, networkID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIBMPINetworkRead(ctx, d, meta)
}

func resourceIBMPINetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, networkID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	networkC := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)
	networkdata, err := networkC.Get(networkID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set(Attr_NetworkID, networkdata.NetworkID)
	d.Set(Arg_Cidr, networkdata.Cidr)
	d.Set(Arg_DNS, networkdata.DNSServers)
	d.Set(Attr_VLanID, networkdata.VlanID)
	d.Set(Arg_NetworkName, networkdata.Name)
	d.Set(Arg_NetworkType, networkdata.Type)
	d.Set(Arg_NetworkJumbo, networkdata.Jumbo)
	d.Set(Arg_NetworkMTU, networkdata.Mtu)
	d.Set(Arg_NetworkAccessConfig, networkdata.AccessConfig)
	d.Set(Arg_Gateway, networkdata.Gateway)
	ipRangesMap := []map[string]interface{}{}
	if networkdata.IPAddressRanges != nil {
		for _, n := range networkdata.IPAddressRanges {
			if n != nil {
				v := map[string]interface{}{
					Arg_EndingIPAddress:   n.EndingIPAddress,
					Arg_StartingIPAddress: n.StartingIPAddress,
				}
				ipRangesMap = append(ipRangesMap, v)
			}
		}
	}
	d.Set(Arg_IPAddressRange, ipRangesMap)

	return nil

}

func resourceIBMPINetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, networkID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges(Arg_NetworkName, Arg_DNS, Arg_Gateway, Arg_IPAddressRange) {
		networkC := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)
		body := &models.NetworkUpdate{
			DNSServers: flex.ExpandStringList((d.Get(Arg_DNS).(*schema.Set)).List()),
		}
		if d.Get(Arg_NetworkType).(string) == VLAN {
			body.Gateway = flex.PtrToString(d.Get(Arg_Gateway).(string))
			body.IPAddressRanges = getIPAddressRanges(d.Get(Arg_IPAddressRange).([]interface{}))
		}

		if d.HasChange(Arg_NetworkName) {
			body.Name = flex.PtrToString(d.Get(Arg_NetworkName).(string))
		}

		_, err = networkC.Update(networkID, body)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIBMPINetworkRead(ctx, d, meta)
}

func resourceIBMPINetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	log.Printf("Calling the network delete functions. ")
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, networkID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	networkC := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)
	err = networkC.Delete(networkID)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func isWaitForIBMPINetworkAvailable(ctx context.Context, client *instance.IBMPINetworkClient, id string, timeout time.Duration) (interface{}, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Retry, State_Build},
		Target:     []string{State_Available},
		Refresh:    isIBMPINetworkRefreshFunc(client, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isIBMPINetworkRefreshFunc(client *instance.IBMPINetworkClient, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		network, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}

		if network.VlanID != nil {
			return network, State_Available, nil
		}

		return network, State_Build, nil
	}
}

func generateIPData(cdir string) (gway, firstip, lastip string, err error) {
	_, ipv4Net, err := net.ParseCIDR(cdir)

	if err != nil {
		return "", "", "", err
	}

	var subnetToSize = map[string]int{
		"21": 2048,
		"22": 1024,
		"23": 512,
		"24": 256,
		"25": 128,
		"26": 64,
		"27": 32,
		"28": 16,
		"29": 8,
		"30": 4,
		"31": 2,
	}

	gateway, err := cidr.Host(ipv4Net, 1)
	if err != nil {
		log.Printf("Failed to get the gateway for this cidr passed in %s", cdir)
		return "", "", "", err
	}
	ad := cidr.AddressCount(ipv4Net)

	convertedad := strconv.FormatUint(ad, 10)
	// Powervc in wdc04 has to reserve 3 ip address hence we start from the 4th. This will be the default behaviour
	firstusable, err := cidr.Host(ipv4Net, 4)
	if err != nil {
		log.Print(err)
		return "", "", "", err
	}
	lastusable, err := cidr.Host(ipv4Net, subnetToSize[convertedad]-2)
	if err != nil {
		log.Print(err)
		return "", "", "", err
	}
	return gateway.String(), firstusable.String(), lastusable.String(), nil

}

func getIPAddressRanges(ipAddressRanges []interface{}) []*models.IPAddressRange {
	ipRanges := make([]*models.IPAddressRange, 0, len(ipAddressRanges))
	for _, v := range ipAddressRanges {
		if v != nil {
			ipAddressRange := v.(map[string]interface{})
			ipRange := &models.IPAddressRange{
				EndingIPAddress:   flex.PtrToString(ipAddressRange[Arg_EndingIPAddress].(string)),
				StartingIPAddress: flex.PtrToString(ipAddressRange[Arg_StartingIPAddress].(string)),
			}
			ipRanges = append(ipRanges, ipRange)
		}
	}
	return ipRanges
}
