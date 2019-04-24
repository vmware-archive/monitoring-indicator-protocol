### Purpose of this Guide

This guide is intended to direct a user on how to use Indicator Protocol to
automatically generate Grafana Dashboards and Prometheus alert configurations
for RabbitMQ on Kubernetes(k8s).
To arrive at this point,
we will also cover installing all of these utilities to k8s and configuring them
to communicate with one another.
We will use Helm charts to install Prometheus, Grafana, RabbitMQ,
and the Prometheus forwarder for RabbitMQ.
After that, we will use `kubectl apply` to install Indicator Protocol
and add an observability definition for Rabbit.
We will assume that you have a working k8s cluster running,
and that you have Helm set up to install releases to the cluster.

### Installing The Helm Charts

Helm is a tool for defining a cohesive set of functionality to be installed
onto a k8s cluster.
A group of software installed via a Helm Chart is called a 'release'.
It supports customizing installations through a `.yaml` file.
The customization options available are defined by the creator of the chart,
and may not exhaustively cover the possible configurations of the underlying
software.

#### RabbitMQ

This is the simplest of the Helm installations, so we'll do it first.
We are able to accept all of the defaults for this one.
Run:
```bash
helm install stable/rabbitmq --name rabbit-mq --namespace rabbit
```
Note that the output includes instructions for retrieving the generated admin
password after installation completes.
It should involve running a command that looks something like this:
```bash
kubectl get secret --namespace rabbit rabbit-mq -o jsonpath="{.data.rabbitmq-password}" | base64 --decode
```
You'll need this password soon, so either retrieve and retain it,
or retain the command.
The default username is likely `user`.
Additionally,
there are instructions detailing how to set up port forwarding,
which will allow you to interact with the application as if it were running on
your local host.
After setting up port forwarding,
you can log into the administrative web application by visiting
`http://localhost:15672`.
This will allow you to validate that the installation was successful.

#### Prometheus Exporter for RabbitMQ

Now that we have Rabbit itself installed, we can add its metrics exporter.
The exporter needs to be told how to communicate with Rabbit,
so installing it is a bit more involved.
If you look at `k8s/helm_config/dev_rabbitmq_exports_values.yml`,
you might see what we mean.
We use this file to override defaults defined by the Prometheus exporter helm
chart as necessary.
The only thing you need to worry about here is the password.
Replace it with the real value from the previous step.
Once you've done so, install the exporter like so:
```bash
helm install stable/prometheus-rabbitmq-exporter --name rabbitmq-exporter --namespace rabbit --values k8s/helm_config/dev_rabbitmq_exporter_values.yml
```
To validate the installation,
you can again follow the instructions from Helm to set up port forwarding.
There is no password to worry about by default.
Once this is done, visit `http://localhost:8080/metrics`.
You should see a prometheus metrics file.
Most importantly, the metric called `rabbitmq_up` should be `1`.
If the metrics page comes up, but that value is `0`,
the exporter isn't able to communicate with Rabbit, so time for some debugging.

#### Prometheus

Now that Rabbit is up and exporting metrics,
lets set something up to collect that information.
Prometheus will need to be configured to scrape the RabbitMQ metrics exporter
you set up earlier.
This is done with a configuration available on the Prometheus Helm Chart called
`extraScrapeConfigs`.
This allows us to append scrape configurations to those shipped by default.
You can see what this looks like in
`k8s/helm_config/dev_prometheus_values.yml`.
Here, we tell Prometheus to scan the `/metrics` endpoint of all k8s services
running in the `rabbit` namespace.
You shouldn't need to make any modifications to this file, just run:
```bash
helm install stable/prometheus --name prometheus --namespace prometheus --values k8s/helm_config/dev_prometheus_values.yml
```
Follow the instructions from helm to set up Port Forwarding again,
there is again no password to worry about.
Visit `http://localhhost:9090/graph` and try to find the metric called
`rabbitmq_up` in the search box.
You should get a chart with a flat line at `1` in response.

#### Grafana

Finally, lets install Grafana so that we'll be able to display a dashboard of
all the metrics defined in our Indicator Document.
Grafana needs to be told how to communicate with Prometheus to gather data.
For Indicator Protocol to create Grafana dashboards,
sidecar dashboards must be enabled as well.
You can see how we accomplish that by inspecting
`k8s/helm_config/dev_grafana_values.yml`.
You don't need to modify this file, though, so you can just run:
```bash
helm install stable/grafana --name grafana --namespace grafana --values k8s/helm_config/dev_grafana_values.yml
```
Helm will tell you how to retrieve the `admin` password and set up port
forwarding.
Log in, create a new dashboard, and try to create a graph tracking the
`rabbitmq_up` metric. You should see another chart with a flat line at `1`.

### Installing Indicator Protocol

Indicator Protocol is the final piece of this puzzle.
It will allow us to create a Grafana dashboard and set of Prometheus Alert
configurations for all key service level indicators of Rabbit by creating
and applying a `.yaml` file.
You can install the resources with one command:
```bash
kubectl apply -k k8s/config
```

### Defining Observability for Rabbit Using Indicator Protocol

Now that all the pieces are in place,
we just need to add an indicator document defining the observability plan for
RabbitMQ.
You can find one at `k8s/examples/rabbitmq-indicators.yml`.
Apply this to k8s like so:
```bash
kubectl apply -f k8s/examples/rabbitmq-indicators.yml
```
Within 2 minutes of running this command,
you should see two things.
First, a new Grafana dashboard called RabbitMQ should appear if you log into the
Grafana UI.
Second, you should see a set of alerting rules relevant to Rabbit defined if you
navigate to Prometheus and click on the 'Alerts' button. 
