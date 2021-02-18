region = "us-east-1"

base = "bai"

name = "Terratest"

locations = [
  "us-east-1a",
  "us-east-1c",
]

public_key_path = "../../test_key.pub"

private_key_path = "../../test_key"

additional_public_keys_path = "../../../config/terraform/additional_public_keys"

internal_domain = "internal.scalar-labs.com"

network = {
  bastion_resource_count = "1"
  bastion_resource_type  = "t3.large"
}
