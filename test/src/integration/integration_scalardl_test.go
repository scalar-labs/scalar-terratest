package integration

import (
	"fmt"
	"io/ioutil"
	"net"
	"flag"
	"os"
	"strings"
	"testing"
	"time"
	"./grpc_helper"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

var TerraformDir = flag.String("directory", "", "Directory path of the terraform module to test")
var CloudProvider = flag.String("cloud_provider", "aws", "Cloud provider")

func TestScalarDL(t *testing.T) {
	t.Run("scalardl", func(t *testing.T) {
		t.Run("ScalarDLWithJavaClient", TestScalarDLWithJavaClientExpectStatusCodeIsValid)
		t.Run("ScalarDLWithGrpcWebClient", TestScalarDLWithGrpcWebClientExpectStatusCodeIsValid)
	})
}

func TestScalarDLWithJavaClientExpectStatusCodeIsValid(t *testing.T) {
	expectedRegisterCertStatusCode := []string{"OK", "CERTIFICATE_ALREADY_REGISTERED"}
	expectedRegisterContractStatusCode := []string{"OK", "CONTRACT_ALREADY_REGISTERED"}
	expectedExecuteContractStatusCode := "OK"
	expectedValidateLedgerStatusCode := "OK"
	expectedListContractsStatusCode := "OK"

	contractID := "test-contract1"
	contractBinaryName := "com.org1.contract.StateUpdater"
	contractClassFile := "./resources/StateUpdater.class"
	assetID := "over9000"
	assetArgument := fmt.Sprintf(`{"asset_id": "%s", "state": 9001}`, assetID)
	propertiesFile := "./resources/test.properties"
	scalarurl := ""

	if os.Getenv("TEST_TYPE") != "k8s" {
		scalarurl = strings.Trim(LookupTargetValue(t, "scalardl", "envoy_dns"), "\"")
	} else {
		scalarurl = getExternalIP(t)
	}

	logger.Logf(t, "URL: %s", scalarurl)
	writePropertiesFile(t, strings.Trim(scalarurl, "\""))

	if !isReachable(t, fmt.Sprintf(`%s:50051`, strings.Trim(scalarurl, "\""))) {
		t.Fatal("Unreachable")
	}

	code, _ := grpc_helper.GrpcJavaRegisterCert(t, propertiesFile)
	assert.Contains(t, expectedRegisterCertStatusCode, code)

	code, _ = grpc_helper.GrpcJavaRegisterContract(t, propertiesFile, contractID, contractBinaryName, contractClassFile)
	assert.Contains(t, expectedRegisterContractStatusCode, code)

	code, _ = grpc_helper.GrpcJavaExectueContract(t, propertiesFile, contractID, assetArgument)
	assert.Equal(t, expectedExecuteContractStatusCode, code)

	code, _ = grpc_helper.GrpcJavaValidateAsset(t, propertiesFile, assetID)
	assert.Equal(t, expectedValidateLedgerStatusCode, code)

	code, _ = grpc_helper.GrpcJavaListContracts(t, propertiesFile)
	assert.Equal(t, expectedListContractsStatusCode, code)
}

func TestScalarDLWithGrpcWebClientExpectStatusCodeIsValid(t *testing.T) {
	expectedStatusCode := 200
	scalarurl := ""

	if os.Getenv("TEST_TYPE") != "k8s" {
		scalarurl = strings.Trim(LookupTargetValue(t, "scalardl", "envoy_dns"), "\"")
	} else {
		scalarurl = getExternalIP(t)
	}

	logger.Logf(t, "URL: %s", scalarurl)

	//Register Certificate
	requestData := "AAAAAFIKBW1kYm94EAEiRzBFAiEAx4josbxWPEuZ7w/imnl5B/xY0FKbQLkuK/E/UFUHbmwCIGBludc3JD3pkTYqmeT186g+rDaoFkLqHg4xCQ8uCL3w"
	listContractURL := fmt.Sprintf(`http://%s:50051/rpc.Ledger/ListContracts`, scalarurl)

	statuCode, _ := grpc_helper.GrpcWebTest(t, listContractURL, requestData)

	assert.Equal(t, expectedStatusCode, statuCode)
}

func writePropertiesFile(t *testing.T, host string) {
	properties := []byte(fmt.Sprintf(`
  scalar.dl.client.server.host=%s
  scalar.dl.client.server.port=50051
  scalar.dl.client.server.privileged_port=50052
  scalar.dl.client.cert_holder_id=test
  scalar.dl.client.cert_version=1
  scalar.dl.client.cert_path=./resources/Test.pem
  scalar.dl.client.private_key_path=./resources/Test-key.pem
  `, host))

	err := ioutil.WriteFile("./resources/test.properties", properties, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func getExternalIP(t *testing.T) string {
	bastionIP := strings.Trim(LookupTargetValue(t, "network", "bastion_ip"), "\"")

	publicHost := ssh.Host{
		Hostname:    bastionIP,
		SshAgent:    true,
		SshUserName: "centos",
	}

	commandGetExternalIP := "kubectl get services prod-scalardl-envoy --output jsonpath=\"{.status.loadBalancer.ingress[0]['hostname','ip']}\""

	output, _ := ssh.CheckSshCommandE(t, publicHost, commandGetExternalIP)

	logger.Logf(t, "URL: %s", output)

	return output
}

func isReachable(t *testing.T, host string) bool {
	logger.Logf(t, "Check tcp connection: %s", host)

	numRetries := 60
	checkInterval := 10

	for i := 0; i <= numRetries; i++ {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			logger.Logf(t, "Connection check fail")
		} else {
			logger.Logf(t, "Connection check OK")
			defer conn.Close()
			return true
		}

		time.Sleep(time.Duration(checkInterval) * time.Second)
	}

	return false
}

func LookupTargetValue(t *testing.T, module string, targetValue string) string {
	terraformOptions := &terraform.Options{
		TerraformDir: *TerraformDir + *CloudProvider + "/" + module,
		NoColor:      true,
	}

	return terraform.OutputRequired(t, terraformOptions, targetValue)
}
