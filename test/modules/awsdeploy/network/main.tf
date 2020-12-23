module "network" {
  source = "git::https://github.com/scalar-labs/scalar-terraform.git//modules/aws/network?ref=master"

  # Required Variables
  name                        = var.name
  locations                   = var.locations
  public_key_path             = var.public_key_path
  private_key_path            = var.private_key_path
  additional_public_keys_path = var.additional_public_keys_path
  internal_domain             = var.internal_domain

  # Optional Variables
  network = var.network
}