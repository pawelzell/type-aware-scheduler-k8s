#!/bin/bash
kHOST=`hostname`
if [[ (($kHOST != "baati") && ($kHOST != "naan") && ($kHOST != "dosa"))  ]]; then
  echo "Host $kHOST not supported, please update this script"
  exit 1
fi

DIR="/mnt/k8s-type-aware-scheduler-$kHOST"
INFLUXDB_DIR="${DIR}/influxdb"
kIP1="10.9.99.1"
kIP2="10.9.99.2"
kIP3="10.2.1.91"
kIP4="10.2.1.93"

sudo apt update
sudo apt install nfs-kernel-server
sudo mkdir -p ${DIR}
sudo chown nobody:nogroup ${DIR}
sudo mkdir -p ${INFLUXDB_DIR}
sudo chown nobody:nogroup ${INFLUXDB_DIR}
sudo chmod 777 ${DIR}
echo "${DIR} ${kIP4}(rw,sync,no_subtree_check)" | sudo tee -a /etc/exports
echo "${DIR} ${kIP3}(rw,sync,no_subtree_check)" | sudo tee -a /etc/exports
echo "${DIR} ${kIP2}(rw,sync,no_subtree_check)" | sudo tee -a /etc/exports
echo "${DIR} ${kIP1}(rw,sync,no_subtree_check)" | sudo tee -a /etc/exports
sudo exportfs -a
sudo systemctl restart nfs-kernel-server
#sudo ufw allow from ${CLIENT_IP} to any port nfs
#sudo ufw status | grep ${CLIENT_IP}
