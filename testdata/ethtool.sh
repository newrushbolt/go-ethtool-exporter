#!/bin/sh

# Beware of `-o pipefail` is not yet avaliable in old versions of DASH shell,
# for example the one used in quay.io/prometheus/golang-builder:1.23-base
set -e

SCRIPT_DIR="$(dirname "$0")"

sleep 0.1

case "$1" in
  eth4)
    cat "$SCRIPT_DIR/eth4.generic_info.src"
    exit 0
    ;;
  -i)
    echo "driver info for $2"
    exit 0
    ;;
  -m)
    echo "module info for $2"
    exit 0
    ;;
  -S)
    echo "statistics for $2"
    exit 0
    ;;
  -*)
    echo "Invalid ethtool mode <$1>"
    exit 1
    ;;
  *)
    echo "generic info for $1"
    exit 0
    ;;
esac
