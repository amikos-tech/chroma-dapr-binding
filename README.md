# Dapr Binding for Chroma Vector Database

This binding allows you to use Chroma Vector Database as an output binding in Dapr.

## Supported Chroma Operations

- ✅ Reset
- ✅ Heartbeat
- ✅ List Collections
- ✅ Get Version
- ✅ Create Collection
- ✅ Delete Collection
- 🚫 Collection Add Embedding
- ⚠️ Collection Get (partial without additional parameters)
- ✅ Collection Count
- 🚫 Collection Query
- 🚫 Collection Modify Embeddings
- 🚫 Collection Update
- 🚫 Collection Upsert
- 🚫 Collection Delete - delete documents in collection

## Prerequisites

Optional Minikube setup:

```bash
minikube start --profile chromago
minikube profile chromago
```

Install Chroma using Helm:

```bash
helm repo add chroma https://amikos-tech.github.io/chromadb-chart/
helm repo update
helm install chroma chroma/chromadb --set chromadb.allowReset=true,chromadb.apiVersion=0.4.5-dev
```

## Installation

```bash
go get github.com/amikos-tech/chroma-dapr-binding
```
