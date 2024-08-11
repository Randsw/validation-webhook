# Validation webhook for kubernetes cluster

This is basic implementation of kubernetes validation webhooks

## Description

Validation webhhok used for validation resoures applied (or updated - depend on configuration) to cluster against some rules.
For example:
    - Presence of mandatory labels or annotations
    - Presense of limit/request
    - Allowed image repository in pod spec

In this example we want that created or updated deployment in namespace what marked with annotation `validate: "true"` has label `team`.

## Instalation

### Requirements

- Docker
- kubectl
- Kind
- Go >=1.16 (optional)

## Usage

- Create kubernetes cluster using `kind`

`kind create cluster --config tests/kind/kind-cluster`

- Build webhook server
  `docker build -t <your-repository>:<your-tag> .`
  Push builded image to your favorite image repository. Or load to your kind cluster -  https://kind.sigs.k8s.io/docs/user/quick-start/#loading-an-image-into-your-cluster

- Generate sefl-signed certificate for webhook. Kubernetes control plane works only over secured connection via https, so we need  certificate and key for our webhook. And also we need to provide CA certificate bundle to kubernetes conrol plane so it can trust our server certificate
All this operation automated by bash script:
`cert_gen/predeloy.sh <serviceName> <namespace> <secretName>`
where:
  - `<serviceName>` - name of service for access to validation webhook server. Need to generate proper SAN field in certificate. Will be needed further
  - `<namespace>` - namespace where we deploy our webhook server. Created in this step. Will be needed further
  - `<secretName>` - Secret where our TLS certificate and key are placed

- Deploy webhook server to kind cluster
  Open `manifests/webhook-deployment-service.yaml` and change namespace field in `deployment`, `serviceAccount`, `service` resources and name of `service` resource to the one you specified when generating certificates.
  Also change `service` namespace and name fields in `ValidatingWebhookConfiguration` in `manifests/webhook.yaml`
  Deploy manifests:
  `kubectl apply -f manifests/`
  Verify that webhook server pod is healthy and running


## Test

- Aplly "good" deployment
`kubectl apply -f tests/test-deployments/good-deployment.yaml`
Deployment created without any problem

- Aplly "bad" deployment
`kubectl apply -f tests/test-deployments/bad-deployment.yaml`
Deployment is rejected by validation controller

```

Error from server: error when creating "tests/test-deployments/bad-deployment.yaml": admission webhook "webhook-server.webhook.svc" denied the request: Denied because the Deployment is missing label team

```
