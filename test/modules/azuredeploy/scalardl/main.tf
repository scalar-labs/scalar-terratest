module "scalardl" {
  source = "git::https://github.com/scalar-labs/scalar-terraform.git//modules/azure/scalardl?ref=master"

  # Required Variables (Use remote state)
  network   = local.network
  cassandra = local.cassandra

  # Optional Variables
  base     = var.base
  scalardl = var.scalardl
  envoy    = var.envoy
}
