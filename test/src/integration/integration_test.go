package test

import (
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

var terraformDir = flag.String("directory", "", "Directory path of the terraform module to test")
var cloudProvider = flag.String("cloud_provider", "aws", "Cloud provider")

func TestEndToEndTerraform(t *testing.T) {
	t.Parallel()
	logger.Logf(t, "Start End To End Test")

	os.Setenv("TEST_TYPE", "terraform")

	defer test_structure.RunTestStage(t, "teardown", func() {
		scalarModules := []string{"monitor", "scalardl", "cassandra", "network"}

		for _, m := range scalarModules {
			terraformOptions := &terraform.Options{
				TerraformDir: *terraformDir + *cloudProvider + "/" + m,
				NoColor:      true,
			}

			logger.Logf(t, "Destroying <%s> Infrastructure", m)
			terraform.DestroyE(t, terraformOptions)
		}

		logger.Logf(t, "Finished End To End Test")
	})

	test_structure.RunTestStage(t, "setup", func() {
		scalarModules := []string{"network", "cassandra", "scalardl", "monitor"}

		for _, m := range scalarModules {
			terraformOptions := &terraform.Options{
				TerraformDir: *terraformDir + *cloudProvider + "/" + m,
				NoColor:      true,
			}

			logger.Logf(t, "Creating <%s> Infrastructure", m)
			terraform.InitAndApply(t, terraformOptions)
		}

		logger.Logf(t, "Finished Creating Infrastructure: Tests will continue in 2 minutes")
		time.Sleep(120 * time.Second)
	})

	test_structure.RunTestStage(t, "goss", func() {
		logger.Logf(t, "Run Ansible playbooks with Goss")
		runGoss(t, []string{"cassandra"}, "cassandra")
	})

	test_structure.RunTestStage(t, "validate", func() {
		t.Run("TestScalarDL", TestScalarDL)
		// t.Run("TestPrometheusAlerts", TestPrometheusAlerts)
	})
}

func TestEndToEndK8s(t *testing.T) {
	t.Parallel()
	logger.Logf(t, "Start k8s End To End Test")

	os.Setenv("TEST_TYPE", "k8s")

	defer test_structure.RunTestStage(t, "teardown", func() {
		runHelmDelete(t)

		scalarModules := []string{"kubernetes", "cassandra", "network"}

		for _, m := range scalarModules {
			terraformOptions := &terraform.Options{
				TerraformDir: *terraformDir + *cloudProvider + "/" + m,
				NoColor:      true,
			}
			if m == "kubernetes" {
				terraform.RunTerraformCommand(t, terraformOptions, "state", "rm", "module.kubernetes.kubernetes_config_map.aws_auth")
			}
			logger.Logf(t, "Destroying <%s> Infrastructure", m)
			terraform.DestroyE(t, terraformOptions)
		}

		logger.Logf(t, "Finished k8s End To End Test")
	})

	test_structure.RunTestStage(t, "setup", func() {
		scalarModules := []string{"network", "cassandra", "kubernetes"}

		for _, m := range scalarModules {
			terraformOptions := &terraform.Options{
				TerraformDir: *terraformDir + *cloudProvider + "/" + m,
				NoColor:      true,
			}

			logger.Logf(t, "Creating <%s> Infrastructure", m)
			terraform.InitAndApply(t, terraformOptions)
		}

		logger.Logf(t, "Finished Creating Infrastructure: Tests will continue in 2 minutes")
		time.Sleep(120 * time.Second)
	})

	test_structure.RunTestStage(t, "ansible", func() {
		logger.Logf(t, "Run Ansible playbooks")
		runAnsiblePlaybooks(t)
		time.Sleep(120 * time.Second)
	})

	test_structure.RunTestStage(t, "validate", func() {
		t.Run("TestScalarDL", TestScalarDL)
	})
}

func lookupTargetValue(t *testing.T, module string, targetValue string) string {
	terraformOptions := &terraform.Options{
		TerraformDir: *terraformDir + *cloudProvider + "/" + module,
		NoColor:      true,
	}

	return terraform.OutputRequired(t, terraformOptions, targetValue)
}

func runAnsiblePlaybooks(t *testing.T) {
	k8sModuleDir := "./scalar-kubernetes"

	// Delete existing dir
	err := os.RemoveAll(k8sModuleDir)
	if err != nil {
		t.Fatal(err)
	}

	// Git clone scalar-kubernetes
	gitClone(t, "scalar-labs/scalar-kubernetes.git", k8sModuleDir)

	// Replace k8s custom values file
	replaceCommand := shell.Command{
		Command:    "sed",
		Args:       []string{"-ie", "s/load-balancer-internal: \"true\"/load-balancer-internal: \"false\"/g", "./conf/scalardl-custom-values.yaml"},
		WorkingDir: k8sModuleDir,
	}

	shell.RunCommand(t, replaceCommand)

	err = files.CopyFile("../../modules/"+*cloudProvider+"/kubernetes/kube_config", "./kube_config")
	if err != nil {
		t.Fatal(err)
	}

	err = files.CopyFile("../../modules/"+*cloudProvider+"/network/network_inventory", "./inventories/network_inventory")
	if err != nil {
		t.Fatal(err)
	}

	// Install tools
	runAnsiblePlaybook(t, k8sModuleDir, "../inventories", []string{"./playbooks/playbook-install-tools.yml", "-e", "base_local_directory=../../../../"})

	// Deploy scalardl
	runAnsiblePlaybook(t, k8sModuleDir, "../inventories", []string{"./playbooks/playbook-deploy-scalardl.yml", "-e", "local_helm_charts_values_directory=../../../conf"})
}

func runGoss(t *testing.T, targetModules []string, targetHosts string) {
	err := files.CopyFile("../../modules/"+*cloudProvider+"/network/ssh.cfg", "./ssh.cfg")
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range targetModules {
		err = files.CopyFile("../../modules/"+*cloudProvider+"/"+m+"/"+m+"_inventory", "./inventories/"+m+"_inventory")
		if err != nil {
			t.Fatal(err)
		}
	}

	// Ansible goss role
	runAnsiblePlaybook(t, "./", "./inventories", []string{"../../modules/" + *cloudProvider + "/network/.terraform/modules/network/provision/ansible/playbooks/goss-server.yml", "-l", targetHosts})
}

func runAnsiblePlaybook(t *testing.T, workingDir string, inventory string, playbookOptions []string) {
	args := []string{"-i", inventory}

	ansibleCommand := shell.Command{
		Command:    "ansible-playbook",
		Args:       append(args, playbookOptions...),
		WorkingDir: workingDir,
	}

	shell.RunCommand(t, ansibleCommand)
}

func gitClone(t *testing.T, repo string, moduleDir string) {
	gitCommand := shell.Command{
		Command:    "git",
		Args:       []string{"clone", "-b", "master", "--depth", "1", "https://github.com/" + repo, moduleDir},
		WorkingDir: "./",
	}

	shell.RunCommand(t, gitCommand)
}

func runHelmDelete(t *testing.T) {
	bastionIP := strings.Trim(lookupTargetValue(t, "network", "bastion_ip"), "\"")

	publicHost := ssh.Host{
		Hostname:    bastionIP,
		SshAgent:    true,
		SshUserName: "centos",
	}

	helmDeleteCommand := "helm delete prod"

	ssh.CheckSshCommandE(t, publicHost, helmDeleteCommand)
}
