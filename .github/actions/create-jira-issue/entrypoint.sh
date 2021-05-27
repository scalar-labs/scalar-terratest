#!/bin/sh

sprint_id=$(curl -u ${JIRA_AUTH} -X GET -H "Content-Type: application/json" https://scalar-labs.atlassian.net/rest/agile/1.0/board/1/sprint?state=active | jq '.values[].id')

if [ -n "${sprint_id}" ]; then
  sprint=',"customfield_10008": '${sprint_id}
fi

curl --request POST \
  --url https://scalar-labs.atlassian.net/rest/api/2/issue \
  --user ${JIRA_AUTH} \
  --header 'Accept: application/json' \
  --header 'Content-Type: application/json' \
  --data "$(cat <<EOF
{
  "fields": {
    "summary": "[Terratest $(date +"%Y-%m-%d")] Failed in ${TEST_ENVIRONMENT}",
    "issuetype": {
      "name": "Bug"
    },
    "project": {
      "key": "DLT"
    },
    "description": "${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}: ${JIRA_ISSUE_DESCRIPTION}",
    "assignee": {
      "id": "${JIRA_ASSIGNEE_ID}"
    }
    ${sprint}
  }
}
EOF
)"
