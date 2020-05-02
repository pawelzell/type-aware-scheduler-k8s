kHOST=`hostname`
if [[ ( $kHOST != "baati" ) && ( $kHOST != "naan" ) && ( $kHOST != "dosa" )]]; then
  echo "Hostname $kHOST not supported, please create kubeadm-config file for new host"
  exit 1
fi
CONFIG_FILE="kubeadm-${kHOST}.yaml"
echo "Will use the config file: $CONFIG_FILE"

sudo kubeadm init --config $CONFIG_FILE || exit 1


mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
kubectl taint nodes --all node-role.kubernetes.io/master-
kubectl apply -f flannel.yaml
#kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
