## Overview

The Flagship configuration yaml defines the desired infrastructure and control plane for a Flagship cluster. Additional properties are not supported and validation against this schema should catch many incorrect configurations.

## Specification

### addons

Additional binaries that are copied into the Flagship bin directory and copied to the bootstrap node. Only supports helm and the binary is copied from the container defined at [images.helm](#imageshelm).


* Type = object
* Required = true

### addons.enabled

The list of binaries to copy.


* Type = array
* Required = false

### bootstrap

Configuration for bootstrap node.


* Type = object
* Required = false

### bootstrap.etcd

Configuration for etcd on bootstrap node.


* Type = object
* Required = true

### bootstrap.etcd.token

Token for etcd used by Kubernetes, Calico, and CoreDNS.


* Type = string
* Required = true
* Default = bootkube

### conf

Configuration for Kubernetes control plane components.


* Type = object
* Required = true

### conf.bootkube

**NOT USED** Configuration for Bootkube.


* Type = object
* Required = true

### conf.bootkube.developer

**NOT USED** Specifies if Bootkube is running on a local development install.


* Type = boolean
* Required = true

### conf.kubelet

Configuration for kubelet. Used to create both the initial configuration file and the systemd configuration file. For details on each setting see [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = object
* Required = true

### conf.kubelet.authentication

Configuration for kubelet authentication.


* Type = object
* Required = false

### conf.kubelet.authentication.anonymous

Configuration for kubelet anonymous authentication.


* Type = object
* Required = false

### conf.kubelet.authentication.anonymous.enabled

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = boolean
* Required = false
* Default = false

### conf.kubelet.authentication.webhook

Configuration for kubelet webhook authentication.


* Type = object
* Required = false

### conf.kubelet.authentication.webhook.enabled

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = boolean
* Required = false
* Default = true

### conf.kubelet.authentication.x509

Configuration for kubelet x509 authentication.


* Type = object
* Required = false

### conf.kubelet.authentication.x509.client_ca_file

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = /etc/kubernetes/ca.crt

### conf.kubelet.authorization

Configuration for kubelet authorization.


* Type = object
* Required = false

### conf.kubelet.authorization.mode

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = Webhook

### conf.kubelet.cgroup_driver

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = cgroupfs

### conf.kubelet.cloud_provider

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = [string null]
* Required = false
* Options = [aws ]

### conf.kubelet.cluster_dns_nameservers

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = array
* Required = false
* Default = 

### conf.kubelet.cluster_domain

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = cluster.local

### conf.kubelet.config_file

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = /var/lib/kubelet/config.yaml

### conf.kubelet.dns




* Type = object
* Required = false

### conf.kubelet.dns.nameservers




* Type = array
* Required = true

### conf.kubelet.dynamic_config_dir

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = /var/lib/kubelet/dynamic-config

### conf.kubelet.fail_swap_on

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = boolean
* Required = false
* Default = false

### conf.kubelet.masters

Configuration for kubelet masters.


* Type = object
* Required = false

### conf.kubelet.masters.node_labels

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = &#34;node-role.kubernetes.io/master&#34;

### conf.kubelet.masters.register_with_taints

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = &#34;node_role_kubernetes.io/master=:NoSchedule&#34;

### conf.kubelet.static_pod_path

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = /etc/kubernetes/manifests

### conf.kubelet.workers

Configuration for kubelet workers.


* Type = object
* Required = false

### conf.kubelet.workers.node_labels

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = &#34;node-role.kubernetes.io/master&#34;

### conf.kubelet.workers.register_with_taints

See [kubelet reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options)


* Type = string
* Required = false
* Default = &#34;node_role_kubernetes.io/master=:NoSchedule&#34;

### environment

Configuration for the Flagship environment.


* Type = object
* Required = true

### environment.additional_ca_certs

List of additional CA certificates to add to the HyperKit VM. This can be used to add corporate certs to allow use of Flagship in corporate environments where traffic is being monitored and filtered. Accepts any valid URI supported by curl.


* Type = array
* Required = false

### environment.arch

The cluster node architecture. Used when getting Kubernetes control plane component binaries. Only amd64 is supported.


* Type = string
* Required = false
* Default = amd64
* Options = [amd64]

### environment.dirs

The directories used by the Flagship components.


* Type = object
* Required = true

### environment.dirs.barrelman

Path to the Barrelman manifests used for installing the Kubernetes control plane components.


* Type = string
* Required = true

### environment.dirs.bootkube

Path to the Bootkube working directory. Used for storing Bootkube generated assets like TLS certs.


* Type = string
* Required = false
* Default = /opt/flagship/bootkube

### environment.dirs.config

Path to config directories.


* Type = object
* Required = false

### environment.dirs.config.docker

Path to config directory for Docker.


* Type = object
* Required = true

### environment.dirs.config.docker.coredns

Path to config directory for Docker CoreDNS.


* Type = string
* Required = true
* Default = /etc/docker/flagship/deployments/coredns

### environment.dirs.downloads

**NOT USED** Probably intended to be used as a temp directory.


* Type = string
* Required = false
* Default = /tmp/flagship

### environment.dirs.flagship

Path to Flagship application directory.


* Type = string
* Required = false
* Default = /opt/flagship

### environment.dirs.flagship_home

Path to Flagship home directory. Used for storing generated cluster configuration and state.


* Type = string
* Required = false

### environment.dirs.persistent_data

Path for persistent data directories.


* Type = object
* Required = false

### environment.dirs.persistent_data.etcd

Path for persistent data mounted into etcd container.


* Type = string
* Required = false
* Default = /var/lib/flagship-etcd

### environment.dirs.persistent_data.etcd-tls

Path for persistent data mounted into etcd TLS container.


* Type = string
* Required = false
* Default = /etc/etcd/tls

### environment.modules

List of additional kernel modules (for example ip_vs) to load on cluster nodes. Requires at least one valid entry.


* Type = array
* Required = true

### environment.packages

List of additional packages to install on cluster nodes. Package names are specific to the package manager used by the OS installed on the cluster nodes (for example apt-get vs. yum). Requires at least one valid entry.


* Type = array
* Required = false

### environment.repository

Additional repositories to add to the cluster node package manager.


* Type = object
* Required = false

### environment.repository.add

List of additional repositories to add. Requires at least one entry. If no additional repositories are needed, set entry to null.


* Type = array
* Required = true

### environment.services

**NOT USED** Specifies desired state of services on cluster nodes.


* Type = object
* Required = false

### environment.services.enabled




* Type = array
* Required = true

### environment.services.started




* Type = array
* Required = true

### helm

Configuration for serving and installing Helm charts for the Flagship components. The local repo is used for Barrelman chart installs too.


* Type = object
* Required = false

### helm.deployments




* Type = object
* Required = false

### helm.repo

Configuration for a local Helm chart repo.


* Type = object
* Required = false

### helm.repo.clean

Specifies if a the local Helm config directory should be removed.


* Type = boolean
* Required = false
* Default = true

### helm.repo.enabled

Specifies if a local Helm chart repo should be started.


* Type = boolean
* Required = false
* Default = true

### helm.repo.name

The name of the local Helm chart repo to add.


* Type = string
* Required = false
* Default = local

### helm.repo.url

The URL of the local Helm chart repo to add.


* Type = string
* Required = false
* Default = http://localhost:8879/charts

### helm.security

Configuration for Helm security.


* Type = object
* Required = false

### helm.security.managed_tls

Specifies whether Helm should run secured by TLS.


* Type = boolean
* Required = false
* Default = false

### images

Container images for getting the core Flagship component and utility binaries. Each image has a specific use and additional images are not supported.


* Type = object
* Required = true

### images.barrelman

The container image to get Barrelman binary from.


* Type = object
* Required = true

### images.barrelman.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.barrelman.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.bootkube

The container image to get Bootkube binary from. Used to create initial temporary control plane.


* Type = object
* Required = true

### images.bootkube.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.bootkube.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.coredns

The container image to get the CoreDNS binary from. We rely on a custom binary with unbound enabled.


* Type = object
* Required = true

### images.coredns.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.coredns.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.etcd

The container image to get the etcd binary from.


* Type = object
* Required = true

### images.etcd.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.etcd.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.flagship_utils

The container image to get the flagship_utils binary from.


* Type = object
* Required = false

### images.flagship_utils.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.flagship_utils.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.helm

The container image to get the Helm binary from.


* Type = object
* Required = true

### images.helm.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.helm.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.hostess

The container image to get the hostess binary from. Used for idempotently modifying /etc/hosts.


* Type = object
* Required = true

### images.hostess.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.hostess.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.kubernetes

The container image to get the Kubernetes control plane component binaries from.


* Type = object
* Required = true

### images.kubernetes.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.kubernetes.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.s3cryptohandler

The container image to get the s3cryptohandler binary from. Used for encrypting files before storing in S3 when using the AWS infrastructure provider. Optional and not used for other providers.


* Type = object
* Required = false

### images.s3cryptohandler.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.s3cryptohandler.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.shellcheck

The container image to get the ShellCheck binary from. Used for static analysis of Flagship Bash code.


* Type = object
* Required = true

### images.shellcheck.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.shellcheck.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.terraform

The container image to get the Terraform binary from. Used for provisioning AWS infrastructure. Optional and not used for other providers.


* Type = object
* Required = false

### images.terraform.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.terraform.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### images.tree

The container image to get the tree binary from. Used for gate testing.


* Type = object
* Required = true

### images.tree.source

Defines the Docker repo for the component.


* Type = string
* Required = false

### images.tree.version

Defines the Docker repo version for the component.


* Type = string
* Required = true

### infrastructure

Configuration for provisioning cluster infrastructure on a cloud provider.


* Type = object
* Required = true

### infrastructure.provider

Configuartion for the provider.


* Type = object
* Required = true

### infrastructure.provider.kind

The type of cloud provider to use for hosting the infrastructure.


* Type = string
* Required = true
* Default = hyperkit
* Options = [aws hyperkit none]

### infrastructure.provider.spec

The cloud provider specific configuration for the infrastructure. The provider is set by [infrastructure.provider.kind](#infrastructureproviderkind). Options for each kind are shown below as infrastructure.provider.spec[KIND].*


* Type = object
* Required = true

### infrastructure.provider.spec[aws].credential




* Type = object
* Required = true

### infrastructure.provider.spec[aws].credential.assume_role




* Type = string
* Required = true

### infrastructure.provider.spec[aws].credential.location




* Type = string
* Required = true

### infrastructure.provider.spec[aws].credential.profile




* Type = string
* Required = true

### infrastructure.provider.spec[aws].credential.use_instance_profile




* Type = object
* Required = true
* Options = [true]

### infrastructure.provider.spec[aws].ec2




* Type = object
* Required = true

### infrastructure.provider.spec[aws].ec2.key_pairs




* Type = object
* Required = false

### infrastructure.provider.spec[aws].ec2.key_pairs.name




* Type = string
* Required = true

### infrastructure.provider.spec[aws].ec2.key_pairs.public_key




* Type = string
* Required = true

### infrastructure.provider.spec[aws].ec2.nodegroups




* Type = object
* Required = false

### infrastructure.provider.spec[aws].ec2.security_groups




* Type = array
* Required = false

### infrastructure.provider.spec[aws].ec2.subnets




* Type = array
* Required = false

### infrastructure.provider.spec[aws].ec2.vpc




* Type = object
* Required = false

### infrastructure.provider.spec[aws].ec2.vpc.cidr




* Type = string
* Required = true

### infrastructure.provider.spec[aws].ec2.vpc.id




* Type = string
* Required = true

### infrastructure.provider.spec[aws].iam




* Type = object
* Required = false

### infrastructure.provider.spec[aws].iam.instance_profile_role_name




* Type = string
* Required = false

### infrastructure.provider.spec[aws].region




* Type = string
* Required = true

### infrastructure.provider.spec[aws].route53




* Type = object
* Required = false

### infrastructure.provider.spec[aws].route53.dns_suffix




* Type = string
* Required = false

### infrastructure.provider.spec[aws].route53.private_zone




* Type = boolean
* Required = false

### infrastructure.provider.spec[aws].s3




* Type = object
* Required = true

### infrastructure.provider.spec[hyperkit].clean




* Type = boolean
* Required = true

### infrastructure.provider.spec[hyperkit].compute




* Type = object
* Required = true

### infrastructure.provider.spec[hyperkit].compute.disk




* Type = string
* Required = true

### infrastructure.provider.spec[hyperkit].compute.ram




* Type = string
* Required = true

### infrastructure.provider.spec[hyperkit].compute.uuid




* Type = string
* Required = true

### infrastructure.provider.spec[hyperkit].compute.vcpu




* Type = integer
* Required = true

### infrastructure.provider.spec[hyperkit].image




* Type = object
* Required = true

### infrastructure.provider.spec[hyperkit].image.boot




* Type = object
* Required = false

### infrastructure.provider.spec[hyperkit].image.boot.initrd




* Type = string
* Required = true

### infrastructure.provider.spec[hyperkit].image.boot.vmlinuz




* Type = string
* Required = true

### infrastructure.provider.spec[hyperkit].image.name




* Type = string
* Required = true

### infrastructure.provider.spec[hyperkit].image.type




* Type = string
* Required = true

### infrastructure.provider.spec[hyperkit].image.url_boot




* Type = string
* Required = false

### infrastructure.provider.spec[hyperkit].image.url_image




* Type = string
* Required = false

### infrastructure.provider.spec[hyperkit].image.url_initrd




* Type = string
* Required = false

### infrastructure.provider.spec[hyperkit].image.url_vmlinuz




* Type = string
* Required = false

### infrastructure.provider.spec[none].bootstrap

The IP address of the bootstrap node that will be used to initialize the cluster


* Type = string
* Required = true

### infrastructure.provider.spec[none].masters

An array of IP addresses for the master nodes in the cluster


* Type = array
* Required = true

### infrastructure.provider.spec[none].user

The user that will be used to log into the cluster&#39;s nodes


* Type = string
* Required = false
* Default = flagship

### infrastructure.provider.spec[none].workers

An array of IP addresses for the worker nodes in the cluster


* Type = array
* Required = true

### infrastructure.provisioner

Configuartion for the provisioner.


* Type = object
* Required = true

### infrastructure.provisioner.kind

The type of provisioner to use for creating the infrastructure.


* Type = string
* Required = true
* Default = bash
* Options = [bash terraform]

### infrastructure.provisioner.spec

The provisioner specific configuration for creating the infrastructure. The provisioner is set by [infrastructure.provisioner.kind](#infrastructureprovisionerkind). Options for each kind are shown below as infrastructure.provisioner.spec[KIND].*


* Type = object
* Required = true

### infrastructure.provisioner.spec[bash].kubeconf




* Type = object
* Required = true

### infrastructure.provisioner.spec[bash].kubeconf.context_name




* Type = string
* Required = true

### infrastructure.provisioner.spec[bash].kubeconf.user_name




* Type = string
* Required = true

### infrastructure.provisioner.spec[terraform].state




* Type = object
* Required = true

### infrastructure.provisioner.spec[terraform].state.spec




* Type = object
* Required = true

### infrastructure.provisioner.spec[terraform].state.spec.key




* Type = string
* Required = true

### infrastructure.provisioner.spec[terraform].state.spec.name




* Type = string
* Required = true

### infrastructure.provisioner.spec[terraform].state.spec.region




* Type = string
* Required = true

### infrastructure.provisioner.spec[terraform].state.type




* Type = string
* Required = true

### network

Configuration for networking.


* Type = object
* Required = true

### network.cluster

Configuration for cluster level networking.


* Type = object
* Required = true

### network.cluster.ha

Configuration for highly available cluster networking.


* Type = object
* Required = true

### network.cluster.ha.coredns

Configuration for CoreDNS.


* Type = object
* Required = true

### network.cluster.ha.coredns.config

Configuration for CoreDNS.


* Type = object
* Required = true

### network.cluster.ha.coredns.config.domains

Configuration for DNS domains to be added to the Corefile. The first key is the


* Type = array
* Required = true

### network.cluster.ha.coredns.config.domains.origin




* Type = string
* Required = true

### network.cluster.ha.coredns.config.domains.port




* Type = integer
* Required = true

### network.cluster.ha.coredns.config.domains.records




* Type = array
* Required = true

### network.cluster.ha.coredns.config.domains.records.host




* Type = string
* Required = false

### network.cluster.ha.coredns.config.domains.records.type




* Type = string
* Required = true

### network.cluster.ha.coredns.config.domains.records.value




* Type = string
* Required = true

### network.cluster.ha.coredns.config.domains.soa




* Type = object
* Required = true

### network.cluster.ha.coredns.config.domains.soa.ns




* Type = string
* Required = true

### network.cluster.ha.coredns.config.domains.soa.user




* Type = string
* Required = true

### network.cluster.ha.coredns.config.domains.ttl




* Type = string
* Required = true

### network.cluster.ha.coredns.config.forwarders

Configuration for DNS forwarders.


* Type = object
* Required = true

### network.cluster.ha.coredns.config.forwarders.ipaddr

List of hosts to forward DNS requests to. These forwarders will be added to the CoreDNS Corefile and used instead of the entries in resolv.conf. At least one entry is required, set it to an empty string &#34;&#34; to use the existing resolv.conf for forwarding.


* Type = array
* Required = true

### network.cluster.ha.coredns.config.forwarders.port

Port to forward DNS requests to.


* Type = integer
* Required = false
* Default = 53

### network.cluster.ha.coredns.enabled

Enable CoreDNS for handling cluster DNS.


* Type = boolean
* Required = false
* Default = false

### network.endpoints

Configuration for cluster DNS endpoints.


* Type = object
* Required = true

### network.endpoints.etcd




* Type = object
* Required = true

### network.endpoints.etcd.api




* Type = object
* Required = true

### network.endpoints.etcd.api.addr




* Type = array
* Required = false

### network.endpoints.etcd.api.dns




* Type = string
* Required = true

### network.endpoints.etcd.api.method




* Type = string
* Required = true

### network.endpoints.etcd.api.peer




* Type = integer
* Required = true

### network.endpoints.etcd.api.port




* Type = integer
* Required = true

### network.endpoints.kubernetes




* Type = object
* Required = true

### network.endpoints.kubernetes.api

Configuration for the Kubernetes API endpoint. A record &lt;addr&gt;


* Type = object
* Required = true

### network.endpoints.kubernetes.api.addr

The IP address of the host running the API.


* Type = array
* Required = false

### network.endpoints.kubernetes.api.dns

The cluster domain used to construct FQDN for DNS records. Used to create CoreDNS A records for the DNS, etcd, bootstrap, master, and worker nodes. Note that our default configuration uses the master nodes to host DNS and etcd but these potentially could be hosted outside the cluster.


* Type = string
* Required = false
* Default = flagship.sh

### network.endpoints.kubernetes.api.method

The protocol the endpoint will use.


* Type = string
* Required = true
* Default = https

### network.endpoints.kubernetes.api.port

The port the endpoint will use.


* Type = integer
* Required = true
* Default = 6443

### network.endpoints.kubernetes.cidr




* Type = object
* Required = true

### network.endpoints.kubernetes.cidr.pod




* Type = string
* Required = true

### network.endpoints.kubernetes.cidr.svc




* Type = string
* Required = true

### network.endpoints.kubernetes.dashboard

Configuration for the Kubernetes Daashboard endpoint.


* Type = object
* Required = false

### network.endpoints.kubernetes.dashboard.nodeport

The node port to configure for the Kubernetes Dashboard.


* Type = integer
* Required = true
* Default = 30910

### network.endpoints.kubernetes.haproxy




* Type = object
* Required = true

### network.endpoints.kubernetes.haproxy.addr




* Type = array
* Required = false

### network.endpoints.kubernetes.haproxy.dns




* Type = array
* Required = true

### network.endpoints.kubernetes.haproxy.method




* Type = string
* Required = false

### network.endpoints.kubernetes.haproxy.port




* Type = integer
* Required = false

### network.endpoints.kubernetes.haproxy.virtualaddr




* Type = string
* Required = true

### network.endpoints.weave




* Type = object
* Required = false

### network.endpoints.weave.scope




* Type = object
* Required = true

### network.endpoints.weave.scope.nodeport




* Type = integer
* Required = true

### network.host




* Type = object
* Required = true

### network.host.interface




* Type = string
* Required = true

### network.host.ipvs




* Type = object
* Required = false

### network.host.ipvs.enabled




* Type = boolean
* Required = true

### network.proxy

Configuration for using a proxy for external requests. To use authentication include the credentials in the address.


* Type = object
* Required = false

### network.proxy.docker

Configuration for the Docker level proxy.


* Type = object
* Required = false

### network.proxy.docker.enabled

If the proxy settings should be enabled.


* Type = boolean
* Required = false
* Default = false

### network.proxy.docker.excludes

The addresses to exclude from sending to the proxy.


* Type = array
* Required = false
* Default = [localhost 192.168.0.0/16 127.0.0.1 .flagship.sh flagship.sh 10.10.10.10/16 172.31.21.150/16]

### network.proxy.docker.ftp

The proxy address for FTP connections.


* Type = string
* Required = false
* Default = http://172.31.25.222:3128

### network.proxy.docker.http

The proxy address for HTTP connections.


* Type = string
* Required = false
* Default = http://172.31.25.222:3128

### network.proxy.docker.https

The proxy address for HTTPS connections.


* Type = string
* Required = false
* Default = http://172.31.25.222:3128

### network.proxy.host

Configuration for the host level proxy.


* Type = object
* Required = false

### network.proxy.host.enabled

If the proxy settings should be enabled.


* Type = boolean
* Required = false
* Default = false

### network.proxy.host.excludes

The addresses to exclude from sending to the proxy.


* Type = array
* Required = false
* Default = [localhost 192.168.0.0/16 127.0.0.1 .flagship.sh flagship.sh 10.10.10.10/16 172.31.21.150/16]

### network.proxy.host.ftp

The proxy address for FTP connections.


* Type = string
* Required = false
* Default = http://172.31.25.222:3128

### network.proxy.host.http

The proxy address for HTTP connections.


* Type = string
* Required = false
* Default = http://172.31.25.222:3128

### network.proxy.host.https

The proxy address for HTTPS connections.


* Type = string
* Required = false
* Default = http://172.31.25.222:3128

### network.sdn

Configuration for the SDN (software defined networking) layer. **This is legacy config and most Flagship config should not include it.**


* Type = object
* Required = false

### network.sdn.deployment

Configuration for the SDN deployment.


* Type = object
* Required = true

### network.sdn.deployment.source

The source of the deployment. A chart directory for chart or a url for url deployment type.


* Type = string
* Required = true
* Default = /opt/flagship/charts/network-calico-0.1.0.tgz

### network.sdn.deployment.type

The type of deployment to use.


* Type = string
* Required = true
* Default = chart
* Options = [chart url]

### network.sdn.kind

The kind of SDN to deploy.


* Type = string
* Required = true
* Default = calico
* Options = [calico]

### network.sdn.utils

Configuration for Calico utilities to install on the system. This can not be expanded to include other utils but the version can be updated by changing the source.


* Type = object
* Required = true

### network.sdn.utils.destination

The filepath to write the source to.


* Type = string
* Required = true
* Default = /opt/flagship/bin/calicoctl

### network.sdn.utils.source

The URL for the calicoctl binary.


* Type = string
* Required = true
* Default = https://github.com/projectcalico/calicoctl/releases/download/v3.4.2/calicoctl

### network.sdn.utils.sudo

Specifies if sudo is required to write to the destination.


* Type = boolean
* Required = true
* Default = false

### node

Configuration for cluster nodes.


* Type = object
* Required = true

### node.bootstrap

Configuration specific to the bootstrap node.


* Type = object
* Required = false

### node.bootstrap.addr

Hostname or IP address of the bootstrap node.


* Type = string
* Required = false

### node.cluster_name

The name of the Kubernetes cluster being created. This is used as a unique identifier. Change this value to install multiple clusters from the same bootstrap or development node.


* Type = string
* Required = true
* Default = cluster.flagship.sh

### node.dns

Configuration specific to the CoreDNS nodes.


* Type = object
* Required = false

### node.dns.addr

Hostname or IP addresses for each node running CoreDNS. Used for joining nodes to CoreDNS cluster, updating nameservers at the host level, and adding hosts to known_hosts.


* Type = array
* Required = false

### node.etcd

Configuration specific to the etcd nodes.


* Type = object
* Required = false

### node.etcd.addr

Hostname or IP addresses for each node running etcd. Used for joining nodes to etcd cluster.


* Type = array
* Required = false

### node.master

Configuration specific to the master nodes.


* Type = object
* Required = false

### node.master.addr

Hostname or IP addresses for each node running master Kubernetes control plane components.


* Type = array
* Required = false

### node.master.schedulable

Specifies if master nodes should configured to schedule workloads.


* Type = boolean
* Required = false
* Default = true

### node.worker

Configuration specific to the worker nodes.


* Type = object
* Required = false

### node.worker.addr

Hostname or IP addresses for each node running worker Kubernetes control plane components.


* Type = array
* Required = false

### node.worker.schedulable

Specifies if worker nodes should configured to schedule workloads.


* Type = boolean
* Required = false
* Default = true

### runtime

The container runtime used by Kubernetes. Only Docker is supported.


* Type = object
* Required = false

### runtime.install

Specifies if the container runtime should be installed with the OS package manager or if it is already installed.


* Type = boolean
* Required = true
* Default = true

### runtime.method

Specifies the method for installing the container runtime. Only distro is supported.


* Type = string
* Required = true
* Default = distro
* Options = [distro]

### runtime.type

The type of container runtime. Defines the type of container runtime. Only Docker is supported.


* Type = string
* Required = true
* Default = docker
* Options = [docker]

### storage

**NOT USED** Configuration for persistent storage.


* Type = object
* Required = true

### storage.persistence_enabled

**NOT USED** Specifies if persistent storage is enabled.


* Type = boolean
* Required = false
* Default = true

### storage.persistence_type

**NOT USED** Specifies the type of persistent storage.


* Type = string
* Required = false
* Default = minio
* Options = [minio]


