#!/bin/bash
# This runs the commands from the README.md
# Will shortly add `ttpforge test` invocation
# to run all the examples
set -ex

# ensure we have a valid ttpforge binary path
ttpforge_binary="$1"
if [ ! -e "${ttpforge_binary}" ]
then
  echo "Provided TTPForge Binary Path Does Not Exist: ${ttpforge_binary}"
  exit 1
fi
ttpforge_binary=$(realpath "${ttpforge_binary}")

# run all commands from the README.md
"${ttpforge_binary}" init
"${ttpforge_binary}" list repos
"${ttpforge_binary}" list ttps
"${ttpforge_binary}" show ttp examples//args/basic.yaml
"${ttpforge_binary}" run examples//args/basic.yaml \
  --arg str_to_print=hello \
  --arg run_second_step=true

# run all the TTP test cases
./run-all-ttp-tests.sh "${ttpforge_binary}" "example-ttps"
