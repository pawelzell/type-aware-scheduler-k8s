#!/bin/bash
kMETRICS_SERVER=../../metrics-server
./cluster-reset.sh
./cluster-master-init.sh && 
  ./up.sh && 
  kubectl apply -f $kMETRICS_SERVER/deploy/1.8+/ && 
  kubectl apply -f ../main/deployment.yaml #&&
  #./up-os-metrics-collector.sh

