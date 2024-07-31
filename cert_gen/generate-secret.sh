#!/bin/bash

usage() {
    echo 'Usage: generate-secret.sh <serviceName> <namespace> <secretName>'
}

if [ "$#" -ne 3 ]; then
    usage
    exit 1
fi

service=$1
namespace=$2
secret=$3

csrName=${service}.${namespace}
tmpdir=$(mktemp -d)
            
echo "creating certs in tmpdir ${tmpdir} "

cat <<EOF >> ${tmpdir}/csr.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF

openssl req -nodes -new -x509 -keyout ${tmpdir}/ca.key -out ${tmpdir}/ca.crt -subj "/CN=${service}.${namespace}.svc"

openssl genrsa -out ${tmpdir}/key.pem 2048

openssl req -new -key ${tmpdir}/key.pem -subj "/CN=${service}.${namespace}.svc" -out ${tmpdir}/server.csr -config ${tmpdir}/csr.conf

openssl x509 -req -extfile <(printf "subjectAltName=DNS:${service}.${namespace}.svc,DNS:${service}.${namespace},DNS:${service}") \
    -days 365 -in ${tmpdir}/server.csr -CA ${tmpdir}/ca.crt -CAkey ${tmpdir}/ca.key -CAcreateserial -out ${tmpdir}/webhook-server-tls.crt

# create the secret with CA cert and server cert/key
kubectl create secret tls ${secret} \
  --key=${tmpdir}/key.pem \
  --cert=${tmpdir}/webhook-server-tls.crt \
  --dry-run=client -o yaml |
kubectl -n ${namespace} apply -f -

# Locate the cluster Certificate Authority for population in webhook YAML.
CA_BUNDLE=`cat ${tmpdir}/ca.crt | base64 | tr -d '\n'`

# Replace static string with CA_BUNDLE contents.
# Add '' after -i to run this on Mac.
# sed -i '' "s/CA_BUNDLE/$CA_BUNDLE/g" webhook.yaml
sed -i "s/CA_BUNDLE/$CA_BUNDLE/g" manifests/webhook.yaml
