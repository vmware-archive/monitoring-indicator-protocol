# Indicator Protocol

This is an **observability as code** project which allows developers to define and expose performance, scaling, and service level indicators for monitoring and alerting. The indicator definition ideally lives in the same repository as the code and is automatically registered when the code is deployed.

There are 3 main uses cases for this project: Generating documentation, validating against an actual deployment's data, 
and keeping a registry of indicators for use in monitoring tools such as prometheus alert manager and grafana.

See [the wiki](https://github.com/cloudfoundry-incubator/indicators/wiki/Home) for more detailed information and documentation.
