#!/bin/bash

# This shell script spins up a new local kind cluster

kind create cluster --name indipro

kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user admin

# helm init
# helm repo update
# kubectl create serviceaccount --namespace kube-system tiller
# kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
# kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
# helm init --service-account tiller --upgrade

kubectl create namespace grafana
kubectl create namespace prometheus

helm repo update

helm install stable/grafana --values helm_config/e2e_grafana_values.yml --namespace grafana --generate-name
helm install stable/prometheus --values helm_config/e2e_prometheus_values.yml --namespace prometheus --generate-name

kubectl apply -k config

