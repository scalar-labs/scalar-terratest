#!/bin/sh -l

curl --request POST \
  --url "${JIRA_BASE_URL}/rest/api/2/issue" \
  --user "${JIRA_USER_EMAIL}:${JIRA_API_TOKEN}" \
  --header 'Accept: application/json' \
  --header 'Content-Type: application/json' \
  --data "$(cat <<EOF
{
  "fields": {
    "summary": "[Terratest $(date +"%Y-%m-%d")] Failed in ${CLOUD_ENV}",
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
