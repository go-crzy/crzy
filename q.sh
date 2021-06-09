#!/bin/bash

red="$(tput setaf 1)$(tput setab 0)"
green="$(tput setaf 2)$(tput setab 0)"
yellow="$(tput setaf 3)$(tput setab 0)"
blue="$(tput setaf 4)$(tput setab 0)"
magenta="$(tput setaf 5)$(tput setab 0)"
cyan="$(tput setaf 6)$(tput setab 0)"
black="$(tput setaf 0)$(tput setab 1)"
nc="$(tput sgr0)"

while true; do
  content=$(curl -s localhost:8081)
  color=$(echo $content | jq -r .color)
  echo "${!color}${content}${nc}"
  sleep 1
done

