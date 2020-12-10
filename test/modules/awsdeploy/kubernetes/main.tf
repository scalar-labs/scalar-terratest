module "kubernetes" {
  source = "git::https://github.com/scalar-labs/scalar-terraform.git//modules/aws/kubernetes?ref=master"

  # Required Variables (Use network remote state)
  network = local.network

  kubernetes_cluster     = var.kubernetes_cluster
  kubernetes_node_groups = var.kubernetes_node_groups

  custom_tags = var.custom_tags
}
