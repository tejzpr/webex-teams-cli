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

echo $OSTYPE
if [[ "$OSTYPE" == "linux-gnu" ]]; then
  if [[ "$ARCH_TYPE"  == "arm"* ]]; then
    echo "ARM"
    ./webex-teams-cli_linux_arm5 -a $WEBEX_ACCESS_TOKEN room --pe $EMAIL --i true msg
  elif [[ "$ARCH_TYPE"  == "x86_64" ]]; then
    echo "Linux x64"
    ./webex-teams-cli_linux -a $WEBEX_ACCESS_TOKEN room --pe $EMAIL --i true msg
  fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
  echo "Darwin"
  ./webex-teams-cli_darwin -a $WEBEX_ACCESS_TOKEN room --pe $EMAIL --i true msg
fi
