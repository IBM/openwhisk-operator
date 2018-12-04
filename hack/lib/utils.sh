
#!/bin/bash
#
# Copyright 2017-2018 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

COLOR_RESET="\e[00m"
COLOR_GREEN="\e[1;32m"
COLOR_RED="\e[00;31m"

BOLD=$(tput bold)
NORMAL=$(tput sgr0)

CHECKMARK="${COLOR_GREEN}✔${COLOR_RESET}"
CROSSMARK="${COLOR_RED}✗${COLOR_RESET}"

# print header in bold
function u::header() {
    echo ""
    echo ${BOLD}${1}${NORMAL}
}

# print test suite name
function u::testsuite() {
    u::header "$1"
    u::header "${BOLD}====${NORMAL}"
}
