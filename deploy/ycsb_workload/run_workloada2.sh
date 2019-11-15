WORKLOADNAME="workloada2"
WORKLOAD="workloads/$WORKLOADNAME"
LOGPATTERN="^(\[UPDATE\]|\[READ\]|\[OVERALL\]).*"

cd /home/cbuser/YCSB && ./bin/ycsb load basic -P $WORKLOAD -s > /dev/null && 
	./bin/ycsb run basic -P $WORKLOAD -s | grep -E "$LOGPATTERN" > "/home/cbuser/results_$WORKLOADNAME.dat"

