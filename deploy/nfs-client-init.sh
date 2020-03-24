kHOST=`hostname`
if [[ $kHOST == "baati" ]]; then
  kSERVER_NAME="naan"
  kSERVER_IP=10.9.99.1
elif [[ $kHOST == "naan" ]]; then
  kSERVER_NAME="baati"
  kSERVER_IP=10.9.99.2
else
  echo "Host $kHOST not supported, please update this script"
  exit 1
fi
kDIR="/mnt/k8s-type-aware-scheduler-${kSERVER_NAME}"
echo "Assuming nfs server $kSERVER_NAME nfs local mount dir: $kDIR"

sudo apt update
sudo apt install nfs-common
sudo mkdir -p ${kDIR}
sudo mount ${kSERVER_IP}:${kDIR} ${kDIR}

