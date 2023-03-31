#!/bin/bash
#############################################################################
# (c) Meta Platforms, Inc. and affiliates. Confidential and proprietary.
#
# X509 Identity Masquerading
#
# Performs these actions:
#   * Steal an x509 cert from a sandcastle job
#   * Assume the identity of this cert
#   * Try to fetch a secret (should trigger a detection for cert being on wrong machine)
#############################################################################

set -e

echo "HELLO WORLD!"
