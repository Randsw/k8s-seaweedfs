#!/usr/bin/env bash

set -e

helm upgrade --install --wait --timeout 35m --atomic --namespace seaweedfs --create-namespace  \
  --repo https://seaweedfs.github.io/seaweedfs-operator/ seaweedfs-operator seaweedfs-operator --values - <<EOF
replicaCount: 1
serviceMonitor:
  enabled: true
EOF

kubectl create namespace app || true

cat << EOF | kubectl apply -f -
apiVersion: seaweed.seaweedfs.com/v1
kind: Seaweed
metadata:
  name: seaweed-app
  namespace: app
spec:
  image: chrislusf/seaweedfs:latest
  volumeServerDiskCount: 1
  master:
    replicas: 3
  volume:
    replicas: 3
    requests:
      storage: 2Gi
  filer:
    replicas: 2
    # Enable persistence for the filer to store IAM metadata persistently
    persistence:
      enabled: true
      storageClassName: standard # Replace with your storage class
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 2Gi
    extraArgs:
      - "-iam"
      - "-s3.iam.readOnly=false"
    iam: true
    config: |
      [leveldb2]
      enabled = true
      dir = "/data/filerldb2"
    ingress:
      enabled: true
      className: nginx
      host: filer.kind.cluster
  s3:
    replicas: 1         
    port: 8333
    domainName: s3.kind.cluster   
    metricsPort: 9327 
    ingress: 
      enabled: true
      className: nginx
      host: s3.kind.cluster
EOF

cat << EOF | kubectl apply -f -
apiVersion: seaweed.seaweedfs.com/v1
kind: S3Identity
metadata:
  name: seaweed-app-user
  namespace: app
spec:
  seaweedRef:
    name: seaweed-app
  account:
    displayName: app-user
    email: user@example.com
EOF

cat << EOF | kubectl apply -f -
apiVersion: seaweed.seaweedfs.com/v1
kind: S3Credentials
metadata:
  name: example-user-creds
  namespace: app
spec:
  seaweedRef:
    name: seaweed-app
  identityRef:
    name: seaweed-app-user
  secretRef:
    name: example-user-s3-secret
EOF

cat << EOF | kubectl apply -f -
apiVersion: seaweed.seaweedfs.com/v1
kind: Bucket
metadata:
  name: data-bucket
  namespace: app
spec:
  clusterRef:
    name: seaweed-app
    namespace: app
EOF

cat << EOF | kubectl apply -f -
apiVersion: seaweed.seaweedfs.com/v1
kind: S3Policy
metadata:
  name: rw-uploads
  namespace: app
spec:
  seaweedRef:
    name: seaweed-app
  statements:
    - effect: Allow
      actions:
        - s3:ListBucket
      resources:
        - data-bucket
    - effect: Allow
      actions:
        - s3:GetObject
        - s3:PutObject
        - s3:DeleteObject
      resources:
        - data-bucket/*
EOF


cat << EOF | kubectl apply -f -
apiVersion: seaweed.seaweedfs.com/v1
kind: S3PolicyBinding
metadata:
  name: example-user-uploads-binding
  namespace: app
spec:
  seaweedRef:
    name: seaweed-app
  policyRef:
    name: rw-uploads
  subjects:
    - kind: S3Identity
      name: seaweed-app-user
EOF

kubectl label configmap seaweedfs-operator-grafana-dashboard -n seaweedfs grafana_dashboard="1" --overwrite