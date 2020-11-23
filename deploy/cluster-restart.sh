#!/bin/bash
./cluster-reset.sh
./cluster-master-init.sh && 
  ./up.sh && 
  kubectl apply -f metrics-server.yaml && 
  kubectl apply -f ../metrics-collector/deployment.yaml &&
  kubectl apply -f ../main/deployment.yaml
  # ./up-os-metrics-collector.sh

