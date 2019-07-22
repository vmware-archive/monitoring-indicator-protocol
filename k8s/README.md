**WARNING: Indicator Protocol for Kubernetes is experimental and is subject to change!**

We welcome any feedback you might have. Feel free to open a github issue.

## Monitoring Indicator Protocol

Monitoring Indicator Protocol is a set of custom resources that implement the
indicator protocol for kubernetes.

### Resources

#### `IndicatorDocument`

`IndicatorDocument` is a resource that allow for indicators,
thresholds, and layouts to be configured for a given product.

#### `Indicator`

`Indicator` is a resource that represents a single indicator inside of an
`IndicatorDocument`.

### Controllers

#### Grafana

The `grafana-indicator-controller` configures dashboards for
grafana based on `IndicatorDocument`s.

#### Prometheus

The `prometheus-indicator-controller` configures alerting rules
for prometheus based on `IndicatorDocument`s.

#### Lifecycle

The `indicator-lifecycle-controller` creates individual Indicator
resources in k8s for each indicator defined in a document.

#### Status

The `indicator-status-controller` updates the status of each indicator 
based on thresholds and query results.

### Cluster setup

```bash
# Create new cluster with a name of your choice
gcloud container clusters create $NAME --zone us-central1-a

# Ensure kubeconfig is pointing at the right cluster
kubectl config get-contexts

# [If necessary] Point at the right cluster
gcloud container clusters get-credentials $NAME -z us-central1-a

# Provide admin privileges to self
#  - Please note that the role binding name must be unique for the cluster
kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $(gcloud config get-value account)

# Initialize Helm
helm init
helm repo update
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
helm init --service-account tiller --upgrade

# Install Grafana helmchart
helm install stable/grafana --values helm_config/dev_grafana_values.yml --name grafana --namespace grafana

# Install Prometheus helmchart
helm install stable/prometheus --name prometheus --namespace prometheus
```

### Deployment

After setting up your cluster,
you can deploy the SRE resources by:

```bash
kubectl apply -k config
```

This will install the controllers using the `latest` tag in Docker.

After deploying the SRE resources you can deploy indicatordocument resources. It may take a bit, so if you see the error message:
```
no endpoints available for service "indicator-admission"
```
give it a few minutes.

```bash
kubectl apply -f test/valid/simple.yml
```

You can verify that this worked by checking for the existence of the resources:

```bash
kubectl get indicators -A
kubectl get indicatordocuments -A
```

These commands should return output reflecting the names defined in
`simple.yml`.
To further verify,
you can check for corresponding dashboards in Grafana and alerting rules in
Prometheus.
Use `helm status grafana` and `helm status prometheus` to get guidance on
reaching the GUIs of these applications.

### E2E Tests

#### Cluster Setup

The end to end tests communicate with the Grafana and Prometheus APIs directly
to ensure the entire flow of functionality.
Thus, to set up a cluster for end to end tests,
the Grafana and Prometheus servers must be available on the internet.
To achieve this,
use the `e2e_*_values` files in `helm_config`,
instead of the standard values files:

```bash
# Install Grafana helmchart
helm install stable/grafana --values helm_config/e2e_grafana_values.yml --name grafana --namespace grafana

# Install Prometheus helmchart
helm install stable/prometheus --values helm_config/e2e_prometheus_values.yml --name prometheus --namespace prometheus
```

#### Running the tests

The k8s tests can be run with a shell script from the root directory of this
repository: `./scripts/test.sh k8s_e2e`.
The tests use the current Kubernetes auth context,
so you must be logged into a cluster to run them.
The cluster must be set up for E2E tests as detailed above.


To run both bosh and k8s e2e tests, run `./scripts/test.sh e2e`
