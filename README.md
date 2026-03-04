# Application Catalog Manager

## Overview

Application Catalog Manager is a Kubernetes controller that synchronizes `ApplicationCatalog`
custom resources to create and manage Kubermatic `ApplicationDefinition` resources. It bridges
Helm chart catalogs with Kubermatic's application management system.

The controller watches `ApplicationCatalog` resources and converts each defined Helm chart
into a corresponding `ApplicationDefinition`, enabling users to manage application catalogs
declaratively.

### Key Features

- Declarative application catalog management via custom resources
- Support for OCI and HTTP/HTTPS Helm repositories
- Per-chart and per-version repository configuration
- Secret-based credential management for private repositories
- Multi-version support per application
- Optional merging with the default Kubermatic application catalog
- Annotation-based filtering for selective default chart inclusion

## Installation

The controller is deployed as part of Kubermatic KKP. See the
[KKP documentation](https://docs.kubermatic.com/kubermatic/) for installation instructions.

### CRD and Samples

The Custom Resource Definition and sample ApplicationCatalog manifests are available in the
repository:

- CRD: `deploy/crd/applicationcatalog.k8c.io_applicationcatalogs.yaml`
- Samples: `deploy/samples/` - various example catalogs demonstrating different configurations

## More Information

For detailed information about Application Catalog Manager, see the
[Application Catalog Manager documentation](https://docs.kubermatic.com/kubermatic/main/tutorials-howtos/applications/application-catalog-manager/).

## Troubleshooting

If you encounter issues, file an issue][1] or talk to us on the [#XXX channel][12] on the [Kubermatic Slack][15].

## Contributing

Thanks for taking the time to join our community and start contributing!

Feedback and discussion are available on [the mailing list][11].

### Before you start

* Please familiarize yourself with the [Code of Conduct][4] before contributing.
* See [CONTRIBUTING.md][2] for instructions on the developer certificate of origin that we require.

### Pull requests

* We welcome pull requests. Feel free to dig through the [issues][1] and jump in.

## Changelog

See [the list of releases][3] to find out about feature changes.

[1]: https://github.com/kubermatic/application-catalog-manager/issues
[2]: https://github.com/kubermatic/application-catalog-manager/blob/main/CONTRIBUTING.md
[3]: https://github.com/kubermatic/application-catalog-manager/releases
[4]: https://github.com/kubermatic/application-catalog-manager/blob/main/CODE_OF_CONDUCT.md

[11]: https://groups.google.com/forum/#!forum/kubermatic-dev
[12]: https://kubermatic.slack.com/messages/XXX
[15]: http://slack.kubermatic.io/

[21]: https://kubermatic.github.io/XXX/
