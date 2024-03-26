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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM-Cloud/power-go-client/clients/instance"

	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
)

const (
	piEndingIPAaddress   = "pi_ending_ip_address"
	piStartingIPAaddress = "pi_starting_ip_address"
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
			Arg_NetworkType: {
				Description:  "PI network type",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{"vlan", "pub-vlan"}),
			},
			Arg_NetworkName: {
				Description: "PI network name",
				Required:    true,
				Type:        schema.TypeString,
			},
			Arg_NetworkDNS: {
				Computed:    true,
				Description: "List of PI network DNS name",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Type:        schema.TypeSet,
			},
			Arg_NetworkCidr: {
				Computed:    true,
				Description: "PI network CIDR",
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_NetworkGateway: {
				Computed:    true,
				Description: "PI network gateway",
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_NetworkJumbo: {
				Computed:      true,
				ConflictsWith: []string{Arg_NetworkMtu},
				Deprecated:    "This field is deprecated, use pi_network_mtu instead.",
				Description:   "PI network enable MTU Jumbo option",
				Optional:      true,
				Type:          schema.TypeBool,
			},
			Arg_NetworkMtu: {
				Computed:      true,
				ConflictsWith: []string{Arg_NetworkJumbo},
				Description:   "PI Maximum Transmission Unit",
				Optional:      true,
				Type:          schema.TypeInt,
			},
			Arg_NetworkAccessConfig: {
				Computed:     true,
				Description:  "PI network communication configuration",
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validate.ValidateAllowedStringValues([]string{"internal-only", "outbound-only", "bidirectional-static-route", "bidirectional-bgp", "bidirectional-l2out"}),
			},
			Arg_CloudInstanceID: {
				Description: "PI cloud instance ID",
				Required:    true,
				Type:        schema.TypeString,
			},
			Arg_NetworkIPAddressRange: {
				Computed:    true,
				Description: "List of one or more ip address range(s)",
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						piEndingIPAaddress: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Ending ip address",
						},
						piStartingIPAaddress: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Starting ip address",
						},
					},
				},
			},

			//Computed Attributes
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
	if v, ok := d.GetOk(Arg_NetworkDNS); ok {
		networkdns := flex.ExpandStringList((v.(*schema.Set)).List())
		if len(networkdns) > 0 {
			body.DNSServers = networkdns
		}
	}

	if v, ok := d.GetOk(Arg_NetworkJumbo); ok {
		body.Jumbo = v.(bool)
	}
	if v, ok := d.GetOk(Arg_NetworkMtu); ok {
		var mtu int64 = int64(v.(int))
		body.Mtu = &mtu
	}
	if v, ok := d.GetOk(Arg_NetworkAccessConfig); ok {
		body.AccessConfig = models.AccessConfig(v.(string))
	}

	if networktype == "vlan" {
		var networkcidr string
		var ipBodyRanges []*models.IPAddressRange
		if v, ok := d.GetOk(Arg_NetworkCidr); ok {
			networkcidr = v.(string)
		} else {
			return diag.Errorf("%s is required when %s is vlan", Arg_NetworkCidr, Arg_NetworkType)
		}

		gateway, firstip, lastip, err := generateIPData(networkcidr)
		if err != nil {
			return diag.FromErr(err)
		}

		ipBodyRanges = []*models.IPAddressRange{{EndingIPAddress: &lastip, StartingIPAddress: &firstip}}

		if g, ok := d.GetOk(Arg_NetworkGateway); ok {
			gateway = g.(string)
		}

		if ips, ok := d.GetOk(Arg_NetworkIPAddressRange); ok {
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
	d.Set(Arg_NetworkCidr, networkdata.Cidr)
	d.Set(Arg_NetworkDNS, networkdata.DNSServers)
	d.Set(Attr_VLanID, networkdata.VlanID)
	d.Set(Arg_NetworkName, networkdata.Name)
	d.Set(Arg_NetworkType, networkdata.Type)
	d.Set(Arg_NetworkJumbo, networkdata.Jumbo)
	d.Set(Arg_NetworkMtu, networkdata.Mtu)
	d.Set(Arg_NetworkAccessConfig, networkdata.AccessConfig)
	d.Set(Arg_NetworkGateway, networkdata.Gateway)
	ipRangesMap := []map[string]interface{}{}
	if networkdata.IPAddressRanges != nil {
		for _, n := range networkdata.IPAddressRanges {
			if n != nil {
				v := map[string]interface{}{
					piEndingIPAaddress:   n.EndingIPAddress,
					piStartingIPAaddress: n.StartingIPAddress,
				}
				ipRangesMap = append(ipRangesMap, v)
			}
		}
	}
	d.Set(Arg_NetworkIPAddressRange, ipRangesMap)

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

	if d.HasChanges(Arg_NetworkName, Arg_NetworkDNS, Arg_NetworkGateway, Arg_NetworkIPAddressRange) {
		networkC := instance.NewIBMPINetworkClient(ctx, sess, cloudInstanceID)
		body := &models.NetworkUpdate{
			DNSServers: flex.ExpandStringList((d.Get(Arg_NetworkDNS).(*schema.Set)).List()),
		}
		if d.Get(Arg_NetworkType).(string) == "vlan" {
			body.Gateway = flex.PtrToString(d.Get(Arg_NetworkGateway).(string))
			body.IPAddressRanges = getIPAddressRanges(d.Get(Arg_NetworkIPAddressRange).([]interface{}))
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
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", Arg_NetworkProvisioning},
		Target:     []string{"NETWORK_READY"},
		Refresh:    isIBMPINetworkRefreshFunc(client, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isIBMPINetworkRefreshFunc(client *instance.IBMPINetworkClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		network, err := client.Get(id)
		if err != nil {
			return nil, "", err
		}

		if network.VlanID != nil {
			return network, "NETWORK_READY", nil
		}

		return network, Arg_NetworkProvisioning, nil
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

	//subnetsize, _ := ipv4Net.Mask.Size()

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
				EndingIPAddress:   flex.PtrToString(ipAddressRange[piEndingIPAaddress].(string)),
				StartingIPAddress: flex.PtrToString(ipAddressRange[piStartingIPAaddress].(string)),
			}
			ipRanges = append(ipRanges, ipRange)
		}
	}
	return ipRanges
}
