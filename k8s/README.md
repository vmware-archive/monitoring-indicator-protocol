
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