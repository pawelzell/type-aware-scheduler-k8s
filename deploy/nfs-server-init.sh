DIR="/mnt/k8s-type-aware-scheduler"
CLIENT_IP="10.9.99.1"

sudo apt update
sudo apt install nfs-kernel-server
sudo mkdir -p ${DIR}
sudo chown nobody:nogroup ${DIR}
sudo chmod 777 ${DIR}
echo "${DIR} ${CLIENT_IP}(rw,sync,no_subtree_check)" | sudo tee -a /etc/exports
sudo exportfs -a
sudo systemctl restart nfs-kernel-server
#sudo ufw allow from ${CLIENT_IP} to any port nfs
#sudo ufw status | grep ${CLIENT_IP}
