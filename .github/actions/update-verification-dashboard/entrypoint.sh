#!/bin/sh


#local timestamp=$(date  +"%s000")
timestamp="1627969698000"
curl --request POST \
    --url http://verification-dashboard.japaneast.cloudapp.azure.com:8080/verification/details \
    --header "Content-Type: application/json" \
    --header "Authorization: Bearer ${AUTOMATE_TOOL_TOKEN}" \
    --data "$(cat <<EOF
   {
	"testId": "${TEST_ID}",
	"testStatus": "${TEST_STATUS}",
	"summary": "[Terratest $(date +"%Y-%m-%d")] Failed in ${TEST_ENVIRONMENT}",
	"buildNumber": "",
	"testDate": ${timestamp},
	"commit":  "",
	"jiraTicket": ""
  }
EOF
)"

