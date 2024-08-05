#!/bin/bash

# Copyright © 2023-present, Meta Platforms, Inc. and affiliates
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

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
