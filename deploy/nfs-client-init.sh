DIR="/mnt/k8s-type-aware-scheduler"
SERVER_IP="10.9.99.2"

sudo apt update
sudo apt install nfs-common
sudo mkdir -p ${DIR}
sudo mount ${SERVER_IP}:${DIR} ${DIR}

