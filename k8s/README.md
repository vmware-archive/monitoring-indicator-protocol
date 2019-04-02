
## SRE Resources

SRE resources are custom resources that implement the indicator
protocol for kubernetes.

### Resources

#### `IndicatorDocument`

`IndicatorDocument` is a resource that allow for indicators,
thresholds, and layouts to be configured for a given product.

### Controllers

#### Grafana

The `grafana-indicator-controller` configures dashboards for
grafana based on `IndicatorDocument`s.

#### Prometheus

The `prometheus-indicator-controller` configures alerting rules
for prometheus based on `IndicatorDocument`s.

### Deployment

You can deploy the SRE resources by:

```
kubectl apply -Rf config
```

### Cluster setup

```bash
# Create new cluster
gcloud container clusters create $NAME --zone us-central1-a

# Provide admin privileges to self
#  - Please note that the role binding name must be unique for the cluster
kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $(gcloud config get-value account)

# Ensure kubeconfig is pointing at the right cluster
kubectl config get-contexts

# [Optional] Point at the right cluster
gcloud container clusters get-credentials $NAME -z us-central1-a

# Initialize Helm
helm init
helm repo update
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
helm init --service-account tiller --upgrade

# Create Grafana namespace
kubectl create ns grafana

# Install Grafana helmchart
helm install stable/grafana --values helm_config/grafana_values.yml --name grafana --namespace grafana

# Create Prometheus namespace
kubectl create ns prometheus

# Install Prometheus helmchart
helm install stable/prometheus --name prometheus --namespace prometheus

# MIP components
kubectl apply -k config

# Apply a simple indicatordocument
kubectl apply -f test/valid/simple.yml
```

#### For test

To set up a cluster for end to end tests,
the grafana and prometheus servers must be available on the internet.
To achieve this,
modify the lines in the above script which install those services like so:

```bash
# Install Grafana helmchart
helm install stable/grafana --values helm_config/e2e_grafana_values.yml --name grafana --namespace grafana
```

```bash
# Install Prometheus helmchart
helm install stable/prometheus --values helm_config/e2e_prometheus_values.yml --name prometheus --namespace prometheus
```
