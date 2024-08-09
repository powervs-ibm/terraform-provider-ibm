resource "ibm_pi_key" "testacc_sshkey" {
  pi_key_name          = "terraform-test-key"
  pi_ssh_key           = "ssh-rsa-long-string"
  pi_cloud_instance_id = "49fba6c9-23f8-40bc-9899-aca322ee7d5b"
}

data "ibm_pi_key" "ds_instance" {
  pi_key_name          = "terraform-test-key"
  pi_cloud_instance_id = "49fba6c9-23f8-40bc-9899-aca322ee7d5b"
}
#-------------------------------------------
output "create_key" {
  value = ibm_pi_key.testacc_sshkey.id
}

output "data_key" {
  value = data.ibm_pi_key.ds_instance.id
}