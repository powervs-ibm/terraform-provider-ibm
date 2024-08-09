mock_resource "ibm_pi_key"{
  defaults = {
    pi_key_name          = "terraform-test-key"
    pi_ssh_key           = "ssh-rsa-long-string"
    pi_cloud_instance_id = "49fba6c9-23f8-40bc-9899-aca322ee7d5b"
    creation_date = "date1"
    id = "22d21196-6bf2-4164-81d0-0c930f827dbb"
    name = "terraform-test-key"
    ssh_key = "ssh-rsa-long-string"
  }
}

mock_data "ibm_pi_key" {
  defaults = {
    pi_key_name          = "terraform-test-key"
    pi_cloud_instance_id = "49fba6c9-23f8-40bc-9899-aca322ee7d5b"
    id = "22d21196-6bf2-4164-81d0-0c930f827dbb"
    creation_date = "date1"
    ssh_key = "ssh-rsa-long-string"
  }
}