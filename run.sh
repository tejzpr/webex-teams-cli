#!/bin/bash

if [[ -z "${WEBEX_ACCESS_TOKEN}" ]]; then
  echo "WEBEX_ACCESS_TOKEN not set Enter Access Token now (Get it from https://developer.webex.com/docs/api/getting-started)"
  read -s -p ": " WEBEX_ACCESS_TOKEN
  echo " "
else
  WEBEX_ACCESS_TOKEN="${WEBEX_ACCESS_TOKEN}"
fi

read -p "Email of the person to contact: " EMAIL

ARCH_TYPE=`uname -m`

./webex-teams-cli -a $WEBEX_ACCESS_TOKEN room --pe $EMAIL --i true msg
