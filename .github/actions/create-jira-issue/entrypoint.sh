#!/bin/sh -l

date=$(date +"%Y-%m-%d")
jira_api_url="${JIRA_BASE_URL}/rest/api/2/issue"
jira_auth="${JIRA_USER_EMAIL}:${JIRA_API_TOKEN}"

curl --request POST \
  --url ${jira_api_url} \
  --user ${jira_auth} \
  --header 'Accept: application/json' \
  --header 'Content-Type: application/json' \
  --data "$(cat <<EOF
{
  "fields": {
    "summary": "[Terratest ${date}] Failed in ${CLOUD_ENV}",
    "issuetype": {
      "name": "Bug"
    },
    "project": {
      "key": "DLT"
    },
    "description": "${JIRA_ISSUE_DESCRIPTION}",
    "assignee": {
      "id": "${JIRA_ASSIGNEE_ID}"
    }
  }
}
EOF
)"
