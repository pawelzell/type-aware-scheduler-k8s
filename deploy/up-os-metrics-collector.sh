#!/bin/bash
kNODE_EXPORTER_DIR="node_exporter-1.0.0-rc.1.linux-amd64"
if [[ ! -e run ]]; then
  mkdir run
fi
if [[ ! -e $kNODE_EXPORTER_DIR ]]; then
  wget https://github.com/prometheus/node_exporter/releases/download/v1.0.0-rc.1/node_exporter-1.0.0-rc.1.linux-amd64.tar.gz
  tar -zxf "${kNODE_EXPORTER_DIR}.tar.gz"
fi
# 1. Run node exporter
./$kNODE_EXPORTER_DIR/node_exporter &
kNODE_EXPORTER_PID="$!"
kNODE_EXPORTER_PIDFILE="run/node_exporter.pid"
echo $kNODE_EXPORTER_PID > $kNODE_EXPORTER_PIDFILE

# 3. Run kubectl  proxy
kINFLUX_POD=$(kubectl get pods | awk '/^influx/ {print $1}')
kubectl port-forward "$kINFLUX_POD" 8086:8086 &
kKUBECTL_PROXY_PID="$!"
kKUBECTL_PROXY_PIDFILE="run/kubectl_proxy.pid"
echo $kKUBECTL_PROXY_PID > $kKUBECTL_PROXY_PIDFILE

cd ../os-metrics-collector || exit 1
go build -o app . || exit 1
./app &
kOS_METRICS_COLLECTOR_PID="$!"
cd ../deploy || exit 1

kOS_METRICS_COLLECTOR_PIDFILE="run/os_metrics_collector.pid"
echo $kOS_METRICS_COLLECTOR_PID > $kOS_METRICS_COLLECTOR_PIDFILE

echo "Node exporter pid: $kNODE_EXPORTER_PID"
echo "kubectl port forward pid: $kKUBECTL_PROXY_PID"
echo "Os metrics collector pid: $kOS_METRICS_COLLECTOR_PID"
