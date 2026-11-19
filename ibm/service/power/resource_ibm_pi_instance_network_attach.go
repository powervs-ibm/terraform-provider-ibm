// Copyright IBM Corp. 2026 All Rights Reserved.
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

func ResourceIBMPIInstanceNetworkAttach() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIInstanceNetworkAttachCreate,
		ReadContext:   resourceIBMPIInstanceNetworkAttachRead,
		DeleteContext: resourceIBMPIInstanceNetworkAttachDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_InstanceID: {
				Description:  "The unique identifier of the instance.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_IPAddress: {
				Description: "The requested IP address of the network interface.",
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeString,
			},
			Arg_NetworkID: {
				Description:  "The ID of the network to attach to the instance.",
				ForceNew:     true,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_NetworkSecurityGroupIDs: {
				Description: "IDs of the network security groups that the network interface is a member of. A maximum of 1 network security group is allowed. If not specified, the default network security group is used.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				ForceNew:    true,
				MaxItems:    1,
				Optional:    true,
				Type:        schema.TypeSet,
			},

			// Attributes
			Attr_ExternalIP: {
				Computed:    true,
				Description: "The external IP address of the network interface.",
				Type:        schema.TypeString,
			},
			Attr_Href: {
				Computed:    true,
				Description: "Link to this PVM instance network.",
				Type:        schema.TypeString,
			},
			Attr_IPAddress: {
				Computed:    true,
				Description: "The IP address of the network interface.",
				Type:        schema.TypeString,
			},
			Attr_MacAddress: {
				Computed:    true,
				Description: "The MAC address of the network interface.",
				Type:        schema.TypeString,
			},
			Attr_NetworkInterfaceID: {
				Computed:    true,
				Description: "ID of the network interface.",
				Type:        schema.TypeString,
			},
			Attr_NetworkName: {
				Computed:    true,
				Description: "The name of the network.",
				Type:        schema.TypeString,
			},
			Attr_NetworkSecurityGroupIDs: {
				Computed:    true,
				Description: "IDs of the network security groups that the network interface is a member of.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Type:        schema.TypeSet,
			},
			Attr_NetworkSecurityGroupsHref: {
				Computed:    true,
				Description: "Links to the network security groups that the network interface is a member of.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Type:        schema.TypeList,
			},
			Attr_Type: {
				Computed:    true,
				Description: "The type of the network.",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceIBMPIInstanceNetworkAttachCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_network_attach", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	instanceID := d.Get(Arg_InstanceID).(string)
	networkID := d.Get(Arg_NetworkID).(string)

	body := &models.PVMInstanceAddNetwork{
		NetworkID: networkID,
	}
	if v, ok := d.GetOk(Arg_IPAddress); ok {
		body.IPAddress = v.(string)
	}
	body.NetworkSecurityGroupIDs = flex.ExpandStringList(d.Get(Arg_NetworkSecurityGroupIDs).(*schema.Set).List())

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	attached, err := client.AddNetwork(instanceID, body)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("AddNetwork failed: %s", err.Error()), "ibm_pi_instance_network_attach", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", cloudInstanceID, instanceID, networkID))
	// The MAC address identifies this interface when the instance has several on the same network.
	d.Set(Attr_MacAddress, attached.MacAddress)

	networkC := instance.NewIBMPIInstanceNetworksClient(ctx, sess, cloudInstanceID)
	_, err = isWaitForPIInstanceNetworkAttached(ctx, networkC, instanceID, networkID, attached.MacAddress, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForPIInstanceNetworkAttached failed: %s", err.Error()), "ibm_pi_instance_network_attach", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	return resourceIBMPIInstanceNetworkAttachRead(ctx, d, meta)
}

func resourceIBMPIInstanceNetworkAttachRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_network_attach", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IdParts failed: %s", err.Error()), "ibm_pi_instance_network_attach", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	if len(parts) != 3 {
		return diag.Errorf("invalid ID %q: expected <pi_cloud_instance_id>/<pi_instance_id>/<pi_network_id>", d.Id())
	}
	cloudInstanceID, instanceID, networkID := parts[0], parts[1], parts[2]

	networkC := instance.NewIBMPIInstanceNetworksClient(ctx, sess, cloudInstanceID)
	pvmNetwork, err := getPIInstanceNetwork(networkC, instanceID, networkID, d.Get(Attr_MacAddress).(string))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("Get network failed: %s", err.Error()), "ibm_pi_instance_network_attach", "read")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	if pvmNetwork == nil {
		log.Printf("[WARNING] network attachment was not found or removed outside of terraform\n")
		d.SetId("")
		return nil
	}

	flat := flattenPvmInstanceNetworksv2([]*models.PVMInstanceNetwork{pvmNetwork})[0]

	d.Set(Arg_CloudInstanceID, cloudInstanceID)
	d.Set(Arg_InstanceID, instanceID)
	d.Set(Arg_NetworkID, networkID)
	d.Set(Attr_ExternalIP, flat[Attr_ExternalIP])
	d.Set(Attr_Href, flat[Attr_Href])
	d.Set(Attr_IPAddress, flat[Attr_IPAddress])
	d.Set(Attr_MacAddress, flat[Attr_MacAddress])
	d.Set(Attr_NetworkInterfaceID, flat[Attr_NetworkInterfaceID])
	d.Set(Attr_NetworkName, flat[Attr_NetworkName])
	d.Set(Attr_NetworkSecurityGroupIDs, flat[Attr_NetworkSecurityGroupIDs])
	d.Set(Attr_NetworkSecurityGroupsHref, flat[Attr_NetworkSecurityGroupsHref])
	d.Set(Attr_Type, flat[Attr_Type])

	return nil
}

func resourceIBMPIInstanceNetworkAttachDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IBMPISession failed: %s", err.Error()), "ibm_pi_instance_network_attach", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	parts, err := flex.IdParts(d.Id())
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("IdParts failed: %s", err.Error()), "ibm_pi_instance_network_attach", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	if len(parts) != 3 {
		return diag.Errorf("invalid ID %q: expected <pi_cloud_instance_id>/<pi_instance_id>/<pi_network_id>", d.Id())
	}
	cloudInstanceID, instanceID, networkID := parts[0], parts[1], parts[2]
	macAddress := d.Get(Attr_MacAddress).(string)

	client := instance.NewIBMPIInstanceClient(ctx, sess, cloudInstanceID)
	err = client.DeleteNetwork(instanceID, networkID, &models.PVMInstanceRemoveNetwork{MacAddress: macAddress})
	if err != nil {
		if isPIInstanceNetworkNotFound(err) {
			log.Printf("[WARNING] network attachment was not found or removed outside of terraform\n")
			d.SetId("")
			return nil
		}
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("DeleteNetwork failed: %s", err.Error()), "ibm_pi_instance_network_attach", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	networkC := instance.NewIBMPIInstanceNetworksClient(ctx, sess, cloudInstanceID)
	_, err = isWaitForPIInstanceNetworkDetached(ctx, networkC, instanceID, networkID, macAddress, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForPIInstanceNetworkDetached failed: %s", err.Error()), "ibm_pi_instance_network_attach", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId("")
	return nil
}

// isPIInstanceNetworkNotFound reports whether the error returned by the instance networks API means the network is not attached.
func isPIInstanceNetworkNotFound(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, NotFound) || strings.Contains(msg, "notfound") || strings.Contains(msg, "[404]") || strings.Contains(msg, "not attached")
}

// getPIInstanceNetwork returns the interface of the instance on the network. If macAddress is set, the interface with that MAC address is returned,
// otherwise the first interface on the network is returned. A nil network and a nil error means the network is not attached.
func getPIInstanceNetwork(client *instance.IBMPIInstanceNetworksClient, instanceID, networkID, macAddress string) (*models.PVMInstanceNetwork, error) {
	networks, err := client.GetAll(instanceID)
	if err != nil {
		if isPIInstanceNetworkNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, pvmNetwork := range networks.Networks {
		if pvmNetwork != nil && pvmNetwork.NetworkID == networkID && (macAddress == "" || pvmNetwork.MacAddress == macAddress) {
			return pvmNetwork, nil
		}
	}
	return nil, nil
}

func isWaitForPIInstanceNetworkAttached(ctx context.Context, client *instance.IBMPIInstanceNetworksClient, instanceID, networkID, macAddress string, timeout time.Duration) (any, error) {
	log.Printf("Waiting for network (%s) to be attached to instance (%s)", networkID, instanceID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Pending},
		Target:     []string{State_Available},
		Refresh:    isPIInstanceNetworkAttachedRefreshFunc(client, instanceID, networkID, macAddress),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isPIInstanceNetworkAttachedRefreshFunc(client *instance.IBMPIInstanceNetworksClient, instanceID, networkID, macAddress string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		pvmNetwork, err := getPIInstanceNetwork(client, instanceID, networkID, macAddress)
		if err != nil {
			return nil, "", err
		}
		if pvmNetwork == nil {
			return &models.PVMInstanceNetwork{}, State_Pending, nil
		}
		return pvmNetwork, State_Available, nil
	}
}

func isWaitForPIInstanceNetworkDetached(ctx context.Context, client *instance.IBMPIInstanceNetworksClient, instanceID, networkID, macAddress string, timeout time.Duration) (any, error) {
	log.Printf("Waiting for network (%s) to be detached from instance (%s)", networkID, instanceID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{State_Removing},
		Target:     []string{State_NotFound},
		Refresh:    isPIInstanceNetworkDetachedRefreshFunc(client, instanceID, networkID, macAddress),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		Timeout:    timeout,
	}

	return stateConf.WaitForStateContext(ctx)
}

func isPIInstanceNetworkDetachedRefreshFunc(client *instance.IBMPIInstanceNetworksClient, instanceID, networkID, macAddress string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		pvmNetwork, err := getPIInstanceNetwork(client, instanceID, networkID, macAddress)
		if err != nil {
			return nil, "", err
		}
		if pvmNetwork == nil {
			return &models.PVMInstanceNetwork{}, State_NotFound, nil
		}
		return pvmNetwork, State_Removing, nil
	}
}
