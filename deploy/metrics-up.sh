# TODO you need to install helm
helm init &&
  kubectl create serviceaccount --namespace kube-system tiller &&
	kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller &&
	kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}' &&
  kubectl apply -f crb-kubelet-api-admin.yaml &&
  until helm install --name metrics-server stable/metrics-server --namespace metrics --set args={"--kubelet-insecure-tls=true,--kubelet-preferred-address-types=InternalIP\,Hostname\,ExternalIP"}
	do
		echo "Waiting for tiller to be available. It may take a few minutes."
		sleep 10s
	done
