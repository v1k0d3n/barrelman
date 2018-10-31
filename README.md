# barrelman
*A project to deploy extremely atomic Helm charts as more complex application release groups.*

Barrelman is a [Helm plugin](https://github.com/helm/helm/blob/master/docs/plugins.md) that strives for document compatability with [Armada](https://github.com/att-comdev/armada) and follows Aramada YAML conventions.

The two main concepts are the ability to process a single YAML file that consists of multiple charts and target state commanding.

The YAML configuration document may contain multiple sub-documents or charts denoted by the YAML directive seperator "---". Each section within the YAML file will be sent as a chart to kubernetes, routed to a kubernetes namespace specified in the section.

Barrelman does diff analysis on each commit and only executes those changes necassary to acheive the configured state. Barrelman can be configured to rollback all changes within the current or last transaction on a detected failure, or when commanded by the command line interface. A failure as indicated by kubernetes when commiting one chart will result in the rolling back to the previously commited state on all configured charts.

As a Helm plugin, Barrelman is largely configured by the Helm environment including Kubernetes server settings and authorization keys. Likewise Barrelman will automatically update the local Helm state when there is a relevant change commited or observed.