go build -o app . &&
  docker build . --tag type-aware-scheduler &&
	kind load docker-image type-aware-scheduler
