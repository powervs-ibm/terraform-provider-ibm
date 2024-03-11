// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/errors"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_cloud_connections"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	vpcUnavailable = regexp.MustCompile("pcloudCloudconnectionsPostServiceUnavailable|pcloudCloudconnectionsPutServiceUnavailable")
)

const (
	vpcRetryCount    = 2
	vpcRetryDuration = time.Minute
)

func ResourceIBMPICloudConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPICloudConnectionCreate,
		ReadContext:   resourceIBMPICloudConnectionRead,
		UpdateContext: resourceIBMPICloudConnectionUpdate,
		DeleteContext: resourceIBMPICloudConnectionDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Required Attributes
			Arg_CloudInstanceID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PI cloud instance ID",
			},
			Arg_CloudConnectionName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the cloud connection",
			},
			Arg_CloudConnectionSpeed: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validate.ValidateAllowedIntValues([]int{50, 100, 200, 500, 1000, 2000, 5000, 10000}),
				Description:  "Speed of the cloud connection (speed in megabits per second)",
			},

			// Optional Attributes
			Arg_CloudConnectionGlobalRouting: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable global routing for this cloud connection",
			},
			Arg_CloudConnectionMetered: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable metered for this cloud connection",
			},
			Arg_CloudConnectionNetworks: {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of Networks to attach to this cloud connection",
			},
			Arg_CloudConnectionClassicEnabled: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable classic endpoint destination",
			},
			Arg_CloudConnectionClassicGreCidr: {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{Arg_CloudConnectionClassicEnabled, Arg_CloudConnectionClassicGreDest},
				Description:  "GRE network in CIDR notation",
			},
			Arg_CloudConnectionClassicGreDest: {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{Arg_CloudConnectionClassicEnabled, Arg_CloudConnectionClassicGreCidr},
				Description:  "GRE destination IP address",
			},
			Arg_CloudConnectionVPCEnabled: {
				Type:         schema.TypeBool,
				Optional:     true,
				Default:      false,
				RequiredWith: []string{Arg_CloudConnectionVPCCRNs},
				Description:  "Enable VPC for this cloud connection",
			},
			Arg_CloudConnectionVPCCRNs: {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				RequiredWith: []string{Arg_CloudConnectionVPCEnabled},
				Description:  "Set of VPCs to attach to this cloud connection",
			},
			Arg_CloudConnectionTransitEnabled: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable transit gateway for this cloud connection",
			},

			//Computed Attributes
			Attr_CloudConnectionId: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cloud connection ID",
			},
			Attr_CloudConnectionStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Link status",
			},
			Attr_CloudConnectionIBMIPAddress: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IBM IP address",
			},
			Attr_CloudConnectionUserIPAddress: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User IP address",
			},
			Attr_CloudConnectionPort: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Port",
			},
			Attr_CloudConnectionClassicGreSource: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "GRE auto-assigned source IP address",
			},
			Attr_CloudConnectionMode: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of service the gateway is attached to",
			},
		},
	}
}

func resourceIBMPICloudConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	name := d.Get(Arg_CloudConnectionName).(string)
	speed := int64(d.Get(Arg_CloudConnectionSpeed).(int))

	body := &models.CloudConnectionCreate{
		Name:  &name,
		Speed: &speed,
	}
	if v, ok := d.GetOk(Arg_CloudConnectionGlobalRouting); ok {
		body.GlobalRouting = v.(bool)
	}
	if v, ok := d.GetOk(Arg_CloudConnectionMetered); ok {
		body.Metered = v.(bool)
	}
	// networks
	if v, ok := d.GetOk(Arg_CloudConnectionNetworks); ok && v.(*schema.Set).Len() > 0 {
		body.Subnets = flex.ExpandStringList(v.(*schema.Set).List())
	}
	// classic
	if v, ok := d.GetOk(Arg_CloudConnectionClassicEnabled); ok {
		classicEnabled := v.(bool)
		classic := &models.CloudConnectionEndpointClassicUpdate{
			Enabled: classicEnabled,
		}
		gre := &models.CloudConnectionGRETunnelCreate{}
		if v, ok := d.GetOk(Arg_CloudConnectionClassicGreCidr); ok {
			greCIDR := v.(string)
			gre.Cidr = &greCIDR
			classic.Gre = gre
		}
		if v, ok := d.GetOk(Arg_CloudConnectionClassicGreDest); ok {
			greDest := v.(string)
			gre.DestIPAddress = &greDest
			classic.Gre = gre
		}
		body.Classic = classic
	}

	// VPC
	if v, ok := d.GetOk(Arg_CloudConnectionVPCEnabled); ok {
		vpcEnabled := v.(bool)
		vpc := &models.CloudConnectionEndpointVPC{
			Enabled: vpcEnabled,
		}
		if v, ok := d.GetOk(Arg_CloudConnectionVPCCRNs); ok && v.(*schema.Set).Len() > 0 {
			vpcIds := flex.ExpandStringList(v.(*schema.Set).List())
			vpcs := make([]*models.CloudConnectionVPC, len(vpcIds))
			for i, vpcId := range vpcIds {
				vpcIdCopy := vpcId[0:]
				vpcs[i] = &models.CloudConnectionVPC{
					VpcID: &vpcIdCopy,
				}
			}
			vpc.Vpcs = vpcs
		}
		body.Vpc = vpc
	}

	// Transit Gateway
	if v, ok := d.GetOk(PICloudConnectionTransitEnabled); ok {
		body.TransitEnabled = v.(bool)
	}

	client := instance.NewIBMPICloudConnectionClient(ctx, sess, cloudInstanceID)
	cloudConnection, cloudConnectionJob, err := client.Create(body)
	if err != nil {
		if vpcUnavailable.Match([]byte(err.Error())) {
			err = retryCloudConnectionsVPC(func() (err error) {
				cloudConnection, cloudConnectionJob, err = client.Create(body)
				return
			}, "create", err)
		}
		if err != nil {
			log.Printf("[DEBUG] create cloud connection failed %v", err)
			return diag.FromErr(err)
		}
	}

	if cloudConnection != nil {
		d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *cloudConnection.CloudConnectionID))
	} else if cloudConnectionJob != nil {
		d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *cloudConnectionJob.CloudConnectionID))

		jobID := *cloudConnectionJob.JobRef.ID

		client := instance.NewIBMPIJobClient(ctx, sess, cloudInstanceID)
		_, err = waitForIBMPIJobCompleted(ctx, client, jobID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIBMPICloudConnectionRead(ctx, d, meta)
}

func resourceIBMPICloudConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := parts[0]
	cloudConnectionID := parts[1]

	ccName := d.Get(Arg_CloudConnectionName).(string)
	ccSpeed := int64(d.Get(Arg_CloudConnectionSpeed).(int))

	client := instance.NewIBMPICloudConnectionClient(ctx, sess, cloudInstanceID)
	jobClient := instance.NewIBMPIJobClient(ctx, sess, cloudInstanceID)

	if d.HasChangesExcept(Arg_CloudConnectionNetworks) {

		body := &models.CloudConnectionUpdate{
			Name:  &ccName,
			Speed: &ccSpeed,
		}
		if v, ok := d.GetOk(Arg_CloudConnectionGlobalRouting); ok {
			globalRouting := v.(bool)
			body.GlobalRouting = &globalRouting
		}
		if v, ok := d.GetOk(Arg_CloudConnectionMetered); ok {
			metered := v.(bool)
			body.Metered = &metered
		}
		// classic
		if v, ok := d.GetOk(Arg_CloudConnectionClassicEnabled); ok {
			classicEnabled := v.(bool)
			classic := &models.CloudConnectionEndpointClassicUpdate{
				Enabled: classicEnabled,
			}
			gre := &models.CloudConnectionGRETunnelCreate{}
			if v, ok := d.GetOk(Arg_CloudConnectionClassicGreCidr); ok {
				greCIDR := v.(string)
				gre.Cidr = &greCIDR
				classic.Gre = gre
			}
			if v, ok := d.GetOk(Arg_CloudConnectionClassicGreDest); ok {
				greDest := v.(string)
				gre.DestIPAddress = &greDest
				classic.Gre = gre
			}
			body.Classic = classic
		} else {
			// need to disable classic if not provided
			classic := &models.CloudConnectionEndpointClassicUpdate{
				Enabled: false,
			}
			body.Classic = classic
		}
		// vpc
		if v, ok := d.GetOk(Arg_CloudConnectionVPCEnabled); ok {
			vpcEnabled := v.(bool)
			vpc := &models.CloudConnectionEndpointVPC{
				Enabled: vpcEnabled,
			}
			if v, ok := d.GetOk(Arg_CloudConnectionVPCCRNs); ok && v.(*schema.Set).Len() > 0 {
				vpcIds := flex.ExpandStringList(v.(*schema.Set).List())
				vpcs := make([]*models.CloudConnectionVPC, len(vpcIds))
				for i, vpcId := range vpcIds {
					vpcs[i] = &models.CloudConnectionVPC{
						VpcID: &vpcId,
					}
				}
				vpc.Vpcs = vpcs
			}
			body.Vpc = vpc
		} else {
			// need to disable VPC if not provided
			vpc := &models.CloudConnectionEndpointVPC{
				Enabled: false,
			}
			body.Vpc = vpc
		}

		_, cloudConnectionJob, err := client.Update(cloudConnectionID, body)
		if err != nil {
			if vpcUnavailable.Match([]byte(err.Error())) {
				err = retryCloudConnectionsVPC(func() (err error) {
					_, cloudConnectionJob, err = client.Update(cloudConnectionID, body)
					return
				}, "update", err)
			}
			if err != nil {
				log.Printf("[DEBUG] update cloud connection failed %v", err)
				return diag.FromErr(err)
			}
		}
		if cloudConnectionJob != nil {
			_, err = waitForIBMPIJobCompleted(ctx, jobClient, *cloudConnectionJob.ID, d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if d.HasChange(Arg_CloudConnectionNetworks) {
		oldRaw, newRaw := d.GetChange(Arg_CloudConnectionNetworks)
		old := oldRaw.(*schema.Set)
		new := newRaw.(*schema.Set)

		toAdd := new.Difference(old)
		toRemove := old.Difference(new)

		// call network add api for each toAdd
		for _, n := range flex.ExpandStringList(toAdd.List()) {
			_, jobReference, err := client.AddNetwork(cloudConnectionID, n)
			if err != nil {
				return diag.FromErr(err)
			}
			if jobReference != nil {
				_, err = waitForIBMPIJobCompleted(ctx, jobClient, *jobReference.ID, d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		// call network delete api for each toRemove
		for _, n := range flex.ExpandStringList(toRemove.List()) {
			_, jobReference, err := client.DeleteNetwork(cloudConnectionID, n)
			if err != nil {
				return diag.FromErr(err)
			}
			if jobReference != nil {
				_, err = waitForIBMPIJobCompleted(ctx, jobClient, *jobReference.ID, d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return resourceIBMPICloudConnectionRead(ctx, d, meta)
}

func resourceIBMPICloudConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := parts[0]
	cloudConnectionID := parts[1]

	client := instance.NewIBMPICloudConnectionClient(ctx, sess, cloudInstanceID)
	cloudConnection, err := client.Get(cloudConnectionID)
	if err != nil {
		uErr := errors.Unwrap(err)
		switch uErr.(type) {
		case *p_cloud_cloud_connections.PcloudCloudconnectionsGetNotFound:
			log.Printf("[DEBUG] cloud connection does not exist %v", err)
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] get cloud connection failed %v", err)
		return diag.FromErr(err)
	}

	d.Set(Attr_CloudConnectionId, cloudConnection.CloudConnectionID)
	d.Set(Arg_CloudConnectionName, cloudConnection.Name)
	d.Set(Arg_CloudConnectionGlobalRouting, cloudConnection.GlobalRouting)
	d.Set(Arg_CloudConnectionMetered, cloudConnection.Metered)
	d.Set(Attr_CloudConnectionIBMIPAddress, cloudConnection.IbmIPAddress)
	d.Set(Attr_CloudConnectionUserIPAddress, cloudConnection.UserIPAddress)
	d.Set(Attr_CloudConnectionStatus, cloudConnection.LinkStatus)
	d.Set(Attr_CloudConnectionPort, cloudConnection.Port)
	d.Set(Arg_CloudConnectionSpeed, cloudConnection.Speed)
	d.Set(Arg_CloudInstanceID, cloudInstanceID)
	d.Set(Attr_CloudConnectionMode, cloudConnection.ConnectionMode)
	if cloudConnection.Networks != nil {
		networks := make([]string, 0)
		for _, ccNetwork := range cloudConnection.Networks {
			if ccNetwork != nil {
				networks = append(networks, *ccNetwork.NetworkID)
			}
		}
		d.Set(Arg_CloudConnectionNetworks, networks)
	}
	if cloudConnection.Classic != nil {
		d.Set(Arg_CloudConnectionClassicEnabled, cloudConnection.Classic.Enabled)
		if cloudConnection.Classic.Gre != nil {
			d.Set(Arg_CloudConnectionClassicGreDest, cloudConnection.Classic.Gre.DestIPAddress)
			d.Set(PICloudConnectionClassicGreSource, cloudConnection.Classic.Gre.SourceIPAddress)
		}
	}
	if cloudConnection.Vpc != nil {
		d.Set(Arg_CloudConnectionVPCEnabled, cloudConnection.Vpc.Enabled)
		if cloudConnection.Vpc.Vpcs != nil && len(cloudConnection.Vpc.Vpcs) > 0 {
			vpcCRNs := make([]string, len(cloudConnection.Vpc.Vpcs))
			for i, vpc := range cloudConnection.Vpc.Vpcs {
				vpcCRNs[i] = *vpc.VpcID
			}
			d.Set(Arg_CloudConnectionVPCCRNs, vpcCRNs)
		}
	}

	return nil
}
func resourceIBMPICloudConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := parts[0]
	cloudConnectionID := parts[1]

	client := instance.NewIBMPICloudConnectionClient(ctx, sess, cloudInstanceID)
	_, err = client.Get(cloudConnectionID)
	if err != nil {
		uErr := errors.Unwrap(err)
		switch uErr.(type) {
		case *p_cloud_cloud_connections.PcloudCloudconnectionsGetNotFound:
			log.Printf("[DEBUG] cloud connection does not exist %v", err)
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] get cloud connection failed %v", err)
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Found cloud connection with id %s", cloudConnectionID)

	deleteJob, err := client.Delete(cloudConnectionID)
	if err != nil {
		log.Printf("[DEBUG] delete cloud connection failed %v", err)
		return diag.FromErr(err)
	}
	if deleteJob != nil {
		jobID := *deleteJob.ID

		client := instance.NewIBMPIJobClient(ctx, sess, cloudInstanceID)
		_, err = waitForIBMPIJobCompleted(ctx, client, jobID, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}

func retryCloudConnectionsVPC(ccVPCRetry func() error, operation string, errMsg error) error {
	for count := 0; count < vpcRetryCount && errMsg != nil; count++ {
		log.Printf("[DEBUG] unable to get vpc details for cloud connection: %v", errMsg)
		time.Sleep(vpcRetryDuration)
		log.Printf("[DEBUG] retrying cloud connection %s, retry #%v", operation, count+1)
		errMsg = ccVPCRetry()
	}
	return errMsg
}
