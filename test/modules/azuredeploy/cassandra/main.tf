module "cassandra" {
  source = "git::https://github.com/scalar-labs/scalar-terraform.git//modules/azure/cassandra?ref=master"

  # Required Variables (Use network remote state)
  network = local.network

  # Optional Variables
  base      = var.base
  cassandra = var.cassandra
  cassy     = var.cassy
  reaper    = var.reaper
}
