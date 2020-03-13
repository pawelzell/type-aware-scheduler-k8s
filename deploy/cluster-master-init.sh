sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --apiserver-advertise-address 10.9.99.2 || exit 1


mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
