#!/bin/sh

sleep 0.1

case "$1" in
  eth0)
    echo "ethtool output for eth0"
    exit 0
    ;;
  eth1)
    cat "../testdata/eth1.generic_info"
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
  *)
    exit 1
    ;;
esac
