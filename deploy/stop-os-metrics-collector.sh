#!/bin/bash
kNODE_EXPORTER_PIDFILE="run/node_exporter.pid"
kKUBECTL_PROXY_PIDFILE="run/kubectl_proxy.pid"
kOS_METRICS_COLLECTOR_PIDFILE="run/os_metrics_collector.pid"
if [[ -e $kNODE_EXPORTER_PIDFILE ]]; then
  kill "$(cat $kNODE_EXPORTER_PIDFILE)" &&
  rm $kNODE_EXPORTER_PIDFILE
fi
if [[ -e $kKUBECTL_PROXY_PIDFILE ]]; then
  kill "$(cat $kKUBECTL_PROXY_PIDFILE)" &&
  rm $kKUBECTL_PROXY_PIDFILE
fi
if [[ -e $kOS_METRICS_COLLECTOR_PIDFILE ]]; then
  kill "$(cat $kOS_METRICS_COLLECTOR_PIDFILE)" &&
  rm $kOS_METRICS_COLLECTOR_PIDFILE
fi
