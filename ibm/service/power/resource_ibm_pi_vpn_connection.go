// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/errors"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_v_p_n_connections"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
)

func ResourceIBMPIVPNConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIVPNConnectionCreate,
		ReadContext:   resourceIBMPIVPNConnectionRead,
		UpdateContext: resourceIBMPIVPNConnectionUpdate,
		DeleteContext: resourceIBMPIVPNConnectionDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description: "PI cloud instance ID",
				Required:    true,
				Type:        schema.TypeString,
			},
			Attr_VPNConnectionName: {
				Description: "Name of the VPN Connection",
				Required:    true,
				Type:        schema.TypeString,
			},
			Attr_VPNIKEPolicyId: {
				Description: "Unique identifier of IKE Policy selected for this VPN Connection",
				Required:    true,
				Type:        schema.TypeString,
			},
			Attr_VPNIPSecPolicyId: {
				Description: "Unique identifier of IPSec Policy selected for this VPN Connection",
				Required:    true,
				Type:        schema.TypeString,
			},
			Attr_VPNConnectionMode: {
				Description:      "Mode used by this VPN Connection, either 'policy' or 'route'",
				DiffSuppressFunc: flex.ApplyOnce,
				Required:         true,
				Type:             schema.TypeString,
				ValidateFunc:     validate.ValidateAllowedStringValues([]string{"policy", "route"}),
			},
			Attr_VPNConnectionNetworks: {
				Description: "Set of network IDs to attach to this VPN connection",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Type:        schema.TypeSet,
			},
			Attr_VPNConnectionPeerGatewayAddress: {
				Description: "Peer Gateway address",
				Required:    true,
				Type:        schema.TypeString,
			},
			Attr_VPNConnectionPeerSubnets: {
				Description: "Set of CIDR of peer subnets",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Type:        schema.TypeSet,
			},

			//Computed Attributes
			Attr_VPNConnectionId: {
				Computed:    true,
				Description: "VPN connection ID",
				Type:        schema.TypeString,
			},
			PIVPNConnectionLocalGatewayAddress: {
				Computed:    true,
				Description: "Local Gateway address, only in 'route' mode",
				Type:        schema.TypeString,
			},
			PIVPNConnectionStatus: {
				Computed:    true,
				Description: "Status of the VPN connection",
				Type:        schema.TypeString,
			},
			PIVPNConnectionVpnGatewayAddress: {
				Computed:    true,
				Description: "Public IP address of the VPN Gateway (vSRX) attached to this VPN Connection",
				Type:        schema.TypeString,
			},
			PIVPNConnectionDeadPeerDetection: {
				Computed:    true,
				Description: "Dead Peer Detection",
				Type:        schema.TypeMap,
			},
		},
	}
}

func resourceIBMPIVPNConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	name := d.Get(Attr_VPNConnectionName).(string)
	ikePolicyId := d.Get(Attr_VPNIKEPolicyId).(string)
	ipsecPolicyId := d.Get(Attr_VPNIPSecPolicyId).(string)
	mode := d.Get(Attr_VPNConnectionMode).(string)
	networks := d.Get(Attr_VPNConnectionNetworks).(*schema.Set)
	peerSubnets := d.Get(Attr_VPNConnectionPeerSubnets).(*schema.Set)
	peerGatewayAddress := d.Get(Attr_VPNConnectionPeerGatewayAddress).(string)
	pga := models.PeerGatewayAddress(peerGatewayAddress)

	body := &models.VPNConnectionCreate{
		IkePolicy:          &ikePolicyId,
		IPSecPolicy:        &ipsecPolicyId,
		Mode:               &mode,
		Name:               &name,
		PeerGatewayAddress: &pga,
	}
	// networks
	if networks.Len() > 0 {
		body.Networks = flex.ExpandStringList(networks.List())
	} else {
		return diag.Errorf("%s is a required field", Attr_VPNConnectionNetworks)
	}
	// peer subnets
	if peerSubnets.Len() > 0 {
		body.PeerSubnets = flex.ExpandStringList(peerSubnets.List())
	} else {
		return diag.Errorf("%s is a required field", Attr_VPNConnectionPeerSubnets)
	}

	client := instance.NewIBMPIVpnConnectionClient(ctx, sess, cloudInstanceID)
	vpnConnection, err := client.Create(body)
	if err != nil {
		log.Printf("[DEBUG] create VPN connection failed %v", err)
		return diag.FromErr(err)
	}

	vpnConnectionId := *vpnConnection.ID
	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, vpnConnectionId))

	if vpnConnection.JobRef != nil {
		jobID := *vpnConnection.JobRef.ID
		jobClient := instance.NewIBMPIJobClient(ctx, sess, cloudInstanceID)

		_, err = waitForIBMPIJobCompleted(ctx, jobClient, jobID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIBMPIVPNConnectionRead(ctx, d, meta)
}

func resourceIBMPIVPNConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, vpnConnectionID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVpnConnectionClient(ctx, sess, cloudInstanceID)
	jobClient := instance.NewIBMPIJobClient(ctx, sess, cloudInstanceID)

	if d.HasChangesExcept(Attr_VPNConnectionNetworks, Attr_VPNConnectionPeerSubnets) {
		body := &models.VPNConnectionUpdate{}

		if d.HasChanges(Attr_VPNConnectionName) {
			name := d.Get(Attr_VPNConnectionName).(string)
			body.Name = name
		}
		if d.HasChanges(Attr_VPNIKEPolicyId) {
			ikePolicyId := d.Get(Attr_VPNIKEPolicyId).(string)
			body.IkePolicy = ikePolicyId
		}
		if d.HasChanges(Attr_VPNIPSecPolicyId) {
			ipsecPolicyId := d.Get(Attr_VPNIPSecPolicyId).(string)
			body.IPSecPolicy = ipsecPolicyId
		}
		if d.HasChanges(Attr_VPNConnectionPeerGatewayAddress) {
			peerGatewayAddress := d.Get(Attr_VPNConnectionPeerGatewayAddress).(string)
			body.PeerGatewayAddress = models.PeerGatewayAddress(peerGatewayAddress)
		}

		_, err = client.Update(vpnConnectionID, body)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChanges(Attr_VPNConnectionNetworks) {
		oldRaw, newRaw := d.GetChange(Attr_VPNConnectionNetworks)
		old := oldRaw.(*schema.Set)
		new := newRaw.(*schema.Set)

		toAdd := new.Difference(old)
		toRemove := old.Difference(new)

		for _, n := range flex.ExpandStringList(toAdd.List()) {
			jobReference, err := client.AddNetwork(vpnConnectionID, n)
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
		for _, n := range flex.ExpandStringList(toRemove.List()) {
			jobReference, err := client.DeleteNetwork(vpnConnectionID, n)
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
	if d.HasChanges(Attr_VPNConnectionPeerSubnets) {
		oldRaw, newRaw := d.GetChange(Attr_VPNConnectionPeerSubnets)
		old := oldRaw.(*schema.Set)
		new := newRaw.(*schema.Set)

		toAdd := new.Difference(old)
		toRemove := old.Difference(new)

		for _, s := range flex.ExpandStringList(toAdd.List()) {
			_, err := client.AddSubnet(vpnConnectionID, s)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		for _, s := range flex.ExpandStringList(toRemove.List()) {
			_, err := client.DeleteSubnet(vpnConnectionID, s)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return resourceIBMPIVPNConnectionRead(ctx, d, meta)
}

func resourceIBMPIVPNConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, vpnConnectionID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVpnConnectionClient(ctx, sess, cloudInstanceID)
	vpnConnection, err := client.Get(vpnConnectionID)
	if err != nil {
		uErr := errors.Unwrap(err)
		switch uErr.(type) {
		case *p_cloud_v_p_n_connections.PcloudVpnconnectionsGetNotFound:
			log.Printf("[DEBUG] VPN connection does not exist %v", err)
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] get VPN connection failed %v", err)
		return diag.FromErr(err)
	}

	d.Set(Attr_VPNConnectionId, vpnConnection.ID)
	d.Set(Attr_VPNConnectionName, vpnConnection.Name)
	if vpnConnection.IkePolicy != nil {
		d.Set(Attr_VPNIKEPolicyId, vpnConnection.IkePolicy.ID)
	}
	if vpnConnection.IPSecPolicy != nil {
		d.Set(Attr_VPNIPSecPolicyId, vpnConnection.IPSecPolicy.ID)
	}
	d.Set(PIVPNConnectionLocalGatewayAddress, vpnConnection.LocalGatewayAddress)
	d.Set(Attr_VPNConnectionMode, vpnConnection.Mode)
	d.Set(Attr_VPNConnectionPeerGatewayAddress, vpnConnection.PeerGatewayAddress)
	d.Set(PIVPNConnectionStatus, vpnConnection.Status)
	d.Set(PIVPNConnectionVpnGatewayAddress, vpnConnection.VpnGatewayAddress)

	d.Set(Attr_VPNConnectionNetworks, vpnConnection.NetworkIDs)
	d.Set(Attr_VPNConnectionPeerSubnets, vpnConnection.PeerSubnets)

	if vpnConnection.DeadPeerDetection != nil {
		dpc := vpnConnection.DeadPeerDetection
		dpcMap := map[string]interface{}{
			PIVPNConnectionDeadPeerDetectionAction:    *dpc.Action,
			PIVPNConnectionDeadPeerDetectionInterval:  strconv.FormatInt(*dpc.Interval, 10),
			PIVPNConnectionDeadPeerDetectionThreshold: strconv.FormatInt(*dpc.Threshold, 10),
		}
		d.Set(PIVPNConnectionDeadPeerDetection, dpcMap)
	}

	return nil
}

func resourceIBMPIVPNConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID, vpnConnectionID, err := splitID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := instance.NewIBMPIVpnConnectionClient(ctx, sess, cloudInstanceID)
	jobClient := instance.NewIBMPIJobClient(ctx, sess, cloudInstanceID)

	jobRef, err := client.Delete(vpnConnectionID)
	if err != nil {
		uErr := errors.Unwrap(err)
		switch uErr.(type) {
		case *p_cloud_v_p_n_connections.PcloudVpnconnectionsDeleteNotFound:
			log.Printf("[DEBUG] VPN connection does not exist %v", err)
			d.SetId("")
			return nil
		}
		log.Printf("[DEBUG] delete VPN connection failed %v", err)
		return diag.FromErr(err)
	}
	if jobRef != nil {
		jobID := *jobRef.ID
		_, err = waitForIBMPIJobCompleted(ctx, jobClient, jobID, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}
