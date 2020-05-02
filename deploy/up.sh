kHOST=`hostname`
if [[ (($kHOST != "baati") && ($kHOST != "naan") && ($kHOST != "dosa")) ]]; then
  echo "Host $kHOST not supported, please update this script"
  exit 1
fi
# Script for kubernetes v1.15
kubectl apply -f serviceAccounts/default\:type-aware-scheduler &&
  kubectl apply -f serviceAccounts/kube-system\:type-aware-scheduler &&
  kubectl create clusterrolebinding type-aware-scheduler-admin --clusterrole=cluster-admin --serviceaccount=default:type-aware-scheduler &&
  kubectl create clusterrolebinding type-aware-scheduler-admin2 --clusterrole=cluster-admin --serviceaccount=kube-system:type-aware-scheduler &&
  kubectl create clusterrolebinding default-view --clusterrole=view --serviceaccount=default:default &&
  # Setup grafana and influxdb:
  kubectl create secret generic influxdb-creds  --from-literal=INFLUXDB_DATABASE=type_aware_scheduler --from-literal=INFLUXDB_USERNAME=root --from-literal=INFLUXDB_PASSWORD=root --from-literal=INFLUXDB_HOST=influxdb &&
  kubectl apply -f "pv-${kHOST}.yaml" &&
  kubectl create -f influxdb-pvc.yaml &&
  kubectl apply -f influxdb-1-15.yaml &&
  kubectl expose deployment influxdb --port=8086 --target-port=8086 --protocol=TCP --type=ClusterIP &&
  kubectl create secret generic grafana-creds --from-literal=GF_SECURITY_ADMIN_USER=admin --from-literal=GF_SECURITY_ADMIN_PASSWORD=graphsRcool &&
  kubectl create configmap grafana-config --from-file=influxdb-datasource.yml=influxdb-datasource.yml --from-file=grafana-dashboard-provider.yml=grafana-dashboard-provider.yml &&
  kubectl apply -f grafana-1-15.yaml &&
  kubectl expose deployment grafana --type=NodePort --port=3000 # &&
#kubectl expose deployment grafana --type=LoadBalancer --port=80 --target-port=3000 --protocol=TCP
# TODO influxdb and grafana manage password
# TODO bring up metrics server using:
# kubectl create -f deploy/1.8+/
