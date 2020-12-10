module "ca" {
  source = "git::https://github.com/scalar-labs/scalar-terraform.git//modules/aws/ca?ref=master"

  # Required Variables (Use remote state)
  network = local.network

  # Optional Variables
  ca = var.ca
}
