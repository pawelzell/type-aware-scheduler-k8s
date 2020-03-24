kHOST=`hostname`
if [[ $kHOST == "baati" ]]; then
  kIP=10.9.99.2
elif [[ $kHOST == "naan" ]]; then
  kIP=10.9.99.1
else
  echo "Hostname $kHOST not supported, please update this script"
  exit 1
fi
echo "Detected ip $kIP"

sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --apiserver-advertise-address $kIP || exit 1


mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
kubectl taint nodes --all node-role.kubernetes.io/master-
kubectl apply -f flannel.yaml
#kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
