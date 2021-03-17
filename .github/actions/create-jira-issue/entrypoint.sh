#!/bin/sh

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
  }
}
EOF
)"
