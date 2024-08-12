mock_provider "ibm" {
  source = "./test_template/ibm"
  alias = "fake"
}
 
run "mock_pi_key_resource" {
    providers = {
    ibm = ibm.fake
  }

  assert {
    condition     = ibm_pi_key.testacc_sshkey.id == "22d21196-6bf2-4164-81d0-0c930f827dbb"
    error_message = "incorrect ssh key id"
  }
  assert {
    condition     = ibm_pi_key.testacc_sshkey.ssh_key == "ssh-rsa-long-string"
    error_message = "incorrect ssh key"
  }
}
# Adding override
override_data {
  target = data.ibm_pi_key.ds_instance

  values = {
   creation_date = "date2"
   
  }
}
run "mock_pi_key_datasource" {
   providers = {
    ibm = ibm.fake
  }
    assert {
    condition     = data.ibm_pi_key.ds_instance.creation_date == "date2"
    error_message = "incorrect creation date"
  }
     
}
