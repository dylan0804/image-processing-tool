PHONY: run test build load-to-minikube deploy verify

run:
	cd cmd/api && go run main.go

test:
	go test -v ./internal/api/handlers

VERSION := v1.0.13

build:
	docker build -t image-processing-tool:$(VERSION) .
	@echo "Built image: image-processing-tool:$(VERSION)"
	@echo "Remember to update k8s/api.yaml with this tag and load/push the image!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"

load-to-minikube: build
	minikube image load image-processing-tool:$(VERSION)
	@echo "Loaded image image-processing-tool:$(VERSION) into Minikube."

deploy: load-to-minikube
	kubectl apply -f k8s/namespace.yaml

	kubectl apply -f k8s/redis.yaml
	kubectl apply -f k8s/elasticsearch.yaml
	kubectl apply -f k8s/kibana.yaml

	kubectl apply -f k8s/api.yaml
	kubectl apply -f k8s/ingress.yaml
	@echo "Deployment applied. Verify with 'make verify'."

verify:
	kubectl get all -n image-processing-tool