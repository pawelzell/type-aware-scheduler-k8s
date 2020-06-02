go build -o app . &&
  docker build . --tag pawelzell/metrics-collector || exit 1

docker push pawelzell/metrics-collector
#kind load docker-image pawelzell/metrics-collector
