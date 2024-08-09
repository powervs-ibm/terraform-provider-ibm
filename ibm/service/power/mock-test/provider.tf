terraform {

  required_version = ">= 0.13"

}

terraform {

  required_providers {

    ibm = {

      #  local 
    #   source  = "terraform.local/local/ibm"
    #   version = "1.55.0"

      source = "IBM-Cloud/ibm"
      version = "1.59.0-beta0"

    }

  }

}
#-------------------------------------------
provider "ibm" {

  ibmcloud_api_key = "key"

  region           = "dal"

  zone             = "dal12"

}