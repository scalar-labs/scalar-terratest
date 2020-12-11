module "pulsar" {
  source = "git::https://github.com/scalar-labs/scalar-terraform.git//modules/aws/pulsar?ref=master"

  # Required Variables (Use network remote state)
  network = local.network

  # Optional Variables
  base      = var.base
  bookie    = var.bookie
  broker    = var.broker
  zookeeper = var.zookeeper
}
