# Deploy SeaweedFS S3 cluster in Kubernetes using SeaweedFS Operator

## Overview

This repository provides a complete, production-ready blueprint for deploying a **SeaweedFS S3 cluster** on a Kubernetes environment using the **SeaweedFS Operator**.

To ensure complete observability, the project includes a fully configured **VictoriaMetrics stack** (Grafana + VictoriaMetrics Operator) tailored to monitor SeaweedFS performance metrics out of the box.

### Architecture Quick Look

* **Infrastructure**: 1 Control Plane + 3 Worker nodes (Kind) with MetalLB for LoadBalancer support.
* **SeaweedFS S3**: 3 Replicas setup with a pre-configured `data-bucket` bucket and dedicated user permissions.
* **App Layer**: A minimal Go client to generate and display mock data.

---

## Prerequisites

Ensure the following tools are installed and updated:

* **Docker** (Engine or Desktop)
* **kubectl**
* **Kind CLI**
* **Helm** (v3.x+)

---

## Getting Started & Deployment

### 1. Provision the Kubernetes Infrastructure

Initialize your local multi-node cluster. This script creates the nodes, sets up a local Docker registry integration, and configures MetalLB for external IP routing.

```bash
./cluster-setup.sh
```

### 2. Deploy the VictoriaMetrics Stack & Grafana

Install the monitoring infrastructure, including cluster-wide Kubernetes dashboards and multi-cluster viewing support.

```bash
./setup-vms.sh
```

#### Accessing Grafana

The default login username is `admin`. Retrieve your auto-generated password by running:

```bash
kubectl get secret --namespace victoria-metrics vm-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```

### 3. Deploy SeaweedFS Operator & Cluster

Deploy the operator, spin up the `Seaweed` and initialize the SeaweedFS cluster (3 master, 3 volume). This step also auto-provisions the `data-bucket` bucket and `app-user`.

```bash
./setup-seaweedfs.sh
```

### 4. Launch the Example Go Application

Deploy the application workloads (Deployment, Service, and Ingress resources) to test end-to-end S3 connectivity.

```bash
kubectl apply -f ./k8s/
```

---

## Verifying the Setup

Once all pods are running, you can interact with the Go application via your browser using the following endpoints:

* **Seed Data**: `http://app.kind.cluster/generate`  
  *Generate 100 objects and put in data-bucket*
* **View Data**: `http://app.kind.cluster/show`  
  *Fetches and prints raw rows directly data-bucket.*
* **Check buckets**: `http://app.kind.cluster/check-bucket`  
  *Check if bucket is accessible by user*

---

## Project Structure

* `/app` - Minimal Go client source code and Dockerfile.
* `app/k8s` - Kubernetes manifests for the application layer (Deployment, Service, Ingress).
* `cluster-setup.sh` - Kind cluster bootstrap and network setup automation.
* `setup-vms.sh` - Helm charts and configuration for VictoriaMetrics/Grafana.
* `setup-seaweedfs.sh` - SeaweedFS Operator and SeaweedFS Custom Resources.

---

## License

This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.
