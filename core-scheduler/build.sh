go build -o app . &&
  docker build . --tag pawelzell/type-aware-scheduler || exit 1

docker push pawelzell/type-aware-scheduler
kind load docker-image type-aware-scheduler
