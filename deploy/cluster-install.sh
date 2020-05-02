#!/bin/bash
# Steps needed once to configure a node for hosting kubernetes
# https://bigstep.com/blog/kubernetes-on-bare-metal-cloud
sudo apt install kubelet=1.15.6-00 kubeadm=1.15.6-00 kubectl=1.15.6-00
sudo apt install docker.io=18.09.7-0ubuntu1~18.04.4
sudo systemctl enable docker.service


#Make sure that the br_netfilter module is loaded before this step. This can be done by running:
# lsmod | grep br_netfilter
# In order to load it, explicitly call:
# modprobe br_netfilter
sudo swapoff -a

# When adding support for new node update ip addresses in
# cluster-master-init.sh flannel.yaml pv-`hostname`.yaml
