.PHONY: test
test:
	@echo "\n🛠️  Running unit tests..."
	go test ./...

.PHONY: build
build:
	@echo "\n🔧  Building Go binaries..."
	GOOS=linux GOARCH=amd64 go build -o bin/validatewebhook .

.PHONY: docker-build
docker-build:
	@echo "\n📦 Building simple-kubernetes-webhook Docker image..."
	docker build -t ghcr.io/randsw/validate-webhook .

.PHONY: push
push: docker-build
	@echo "\n📦 Pushing admission-webhook image into Kind's Docker daemon..."
	docker push ghcr.io/randsw/validate-webhook

# From this point `kind` is required
.PHONY: cluster
cluster:
	@echo "\n🔧 Creating Kubernetes cluster..."
	kind create cluster --config tests/kind/kind-cluster.yaml

.PHONY: delete-cluster
delete-cluster:
	@echo "\n♻️  Deleting Kubernetes cluster..."
	kind delete cluster

.PHONY: gen-certs
gen-cert:
    @echo "\n♻️  Generate certs and TLS secret..."
	cert_gen/predeploy.sh webhook-server webhook-certs

.PHONY: deploy-config
deploy-config:
	@echo "\n⚙️  Applying cluster config..."
	kubectl apply -f manifests/webhook.yaml

.PHONY: delete-config
delete-config:
	@echo "\n♻️  Deleting Kubernetes cluster config..."
	kubectl delete -f manifests/webhook.yaml

.PHONY: deploy
deploy: push delete deploy-config
	@echo "\n🚀 Deploying simple-kubernetes-webhook..."
	kubectl apply -f manifests/webhook-deployment-service.yaml

.PHONY: delete
delete:
	@echo "\n♻️  Deleting simple-kubernetes-webhook deployment if existing..."
	kubectl delete -f manifests/webhook-deployment-service.yaml

.PHONY: deployment
pod:
	@echo "\n🚀 Deploying test deployment..."
	kubectl apply -f tests/test-deployments/good-deployment.yaml

.PHONY: delete-deployment
delete-pod:
	@echo "\n♻️ Deleting test pod..."
	kubectl delete -f tests/test-deployments/good-deployment.yaml

.PHONY: bad-deployment
bad-pod:
	@echo "\n🚀 Deploying \"bad\" pod..."
	kubectl apply -f tests/test-deployments/bad-deployment.yaml

.PHONY: delete-bad-deployment
delete-bad-pod:
	@echo "\n🚀 Deleting \"bad\" pod..."
	kubectl delete -f tests/test-deployments/bad-deployment.yaml

.PHONY: logs
logs:
	@echo "\n🔍 Streaming simple-kubernetes-webhook logs..."
	kubectl logs -l app=webhook-server -f

.PHONY: delete-all
delete-all: delete delete-config delete-deployment delete-bad-deployment