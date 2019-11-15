#TODO revert clusterrolebinding hack
#TODO make sure that proxy runs only once
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0-beta4/aio/deploy/recommended.yaml &&
	kubectl apply -f dashboard-adminuser.yaml &&
  kubectl create clusterrolebinding kubernetes-dashboard-admin2 --clusterrole=cluster-admin --serviceaccount=kubernetes-dashboard:kubernetes-dashboard &&
  kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}') &&
	echo "Dashboard access: http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/"
kubectl proxy &
