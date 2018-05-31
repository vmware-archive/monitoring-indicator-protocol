# CF Indicators

Tooling for CF component teams to define and expose performance, scaling, and service level indicators.

There are 3 main uses cases for this project: Generating documentation, validating against an actual deployment's data, 
and keeping a registry of indicators (ideally running in a bosh deployment) for use in monitoring tools.

### Documentation
The `generate_docs` command takes an indicator definition file as a
command-line argument and converts it into HTML documentation:

```
go install cmd/generate_docs
generate_docs example.yml
```

If multiple pages are required a tool like `bosh interpolate` can be used to generate a suitable input file.

### Indicator Validation (Not available)
The `validate` command does 2 things:

1. Verifies that your bosh deployment is emitting the correct metrics to loggregator based on the `metrics` block of your indicator yml. 
1. Verifies that your indicator expressions (PromQL) return values and don't trigger any warning/critical thresholds. This is based on the `indicators` block. 

It takes an indicator definition file and configuration for connecting to log-cache. Both a report and 0/1 exit status are produced 

```
go install cmd/validate
validate --indicators example.yml --deployment cf --log-cache-url http://log-cache.my-env.cf-app.com --log-cache-client my-uaa-client --log-cache-client-secret client-secret
```

The UAA client must have the `doppler.firehose` scope. The `deployment` is bosh deployment name 
that exists on this director. All validation (bosh raw metrics and indicators) will include this
deployment as a tag/label for reading metrics and executing promql.  

### Indicator Registry (Not available)
The `registry` command is a web service that holds a list of current indicators for each deployment. Monitoring
tools can use this information to draw graphs and trigger alerts on the indicator expressions and thresholds. 

## The Indicator Format
All of the packages in this repository consume a YAML formatted file. This file
should define lists of `indicators` and `metrics`, and it can also define a 
`documentation` section.

### The metrics block
The `metric` block defines a list of metrics with the following
attributes:
- **metrics** \[array, required\]
  - **title** \[string,required\]: The human-readable title of the metric
  - **name** \[string,required\]: The name of the metric emitted by the component.
  - **description**  \[markdown,required\]: A formatted description of the metric.

### The indicators block
The `indicators` block defines a list of indicators with the following
attributes:  
- **indicators** \[array, required\]
  - **title** \[string,required\]: The human-readable title of the indicator.
  - **name** \[string,required\]: A unique name used for reference in the `documentation` block.
  - **description**  \[markdown,required\]: A formatted description of the indicator.
  - **metrics** \[array,required\]: References metrics that are used in the measurement.
  - **measurement** \[markdown,required\]: The human-readable explanation of how the indicator is measured.
  - **promql** \[string,optional\]: The Prometheus Query Language (PromQL) expression for producing the measurement value.
  - **thresholds** \[array,required\]: Specifies the conditions for states defined
  by the `level` attribute.  
    - **level** \[string,required\]: The state of the object under measurement if the threshold is met. `critical` and `warning` produce "Red critical" and "Yellow warning" in documentation.
    - one of **[gt,gte,eq,neq,lte,lt]** \[number,required\]: Condition for meeting this threshold based on the measurement value. 
    - **dynamic** \[bool,optional\]: If set to `true`, documentation will omit the condition for this threshold, replacing it with "Dynamic." This signals that the threshold is variable across environments.
  - **response** \[markdown,required\]: A formatted field describing the recommended operator response to a threshold being met. This should describe in detail potential causes, diagnoses, and resolutions.
  
### The documentation block
The `documentation` block defines the composition of HTML documentation generated using the `cmd/generate_docs` package. See the [Healthwatch KPIs](https://docs.pivotal.io/pivotalcf/1-12/monitoring/kpi.html) for an example.
- **documentation** \[hash, required\]
  - **title** \[string,required\]: The top level page header. 
  - **description**  \[markdown,optional\]: A formatted text block that appears under the table of contents. 
  - **sections** \[array,required\]
    - **title** \[string,required\]: The title of the section.
    - **description** \[markdown,optional\]: A formatted text block that appears under the title. 
    - **indicators** \[array,optional\]: An array of indicator names from the `indicators` block defined above. 
    - **metrics** \[array,optional\]: An array of metric names from the `metrics` block defined above. 
