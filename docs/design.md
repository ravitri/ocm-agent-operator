# OCM Agent Operator Design

## API

The following API definitions are part of the OCM Agent Operator, using the API group/version `ocmagent.managed.openshift.io/v1alpha1`.

### OcmAgent

The `OcmAgent` Custom Resource Definition defines the deployment of the OCM Agent on a cluster.

Usage on-cluster:

```bash
$ oc get ocmagent -n openshift-ocm-agent-operator
```

### ManagedNotification

The `ManagedNotification` Custom Resource Definition defines the notification templates that are used by the OCM Agent for sending Service Log notifications.

Usage on-cluster:

```bash
$ oc get managednotification -n openshift-ocm-agent-operator
```

## Controllers

### OCMAgent Controller

The [OCMAgent Controller](https://github.com/openshift/ocm-agent-operator/tree/master/pkg/controller/ocmagent/ocmagent_controller.go) is responsible for ensuring the deployment or removal of an OCM Agent based upon the presence of an `OCMAgent` Custom Resource.

An `OcmAgent` deployment consists of:

- A `ServiceAccount` (named `ocm-agent`)
- A `Role` and `RoleBinding` (both named `ocm-agent`) that defines the OCM Agent's API  permissions.
- A `Deployment` (named `ocm-agent`) which runs the [ocm-agent](https://quay.io/openshift/ocm-agent)
- A `ConfigMap` (name defined in the `OcmAgent` CR) which contains the agent's configuration.
- A `Secret` (name defined in the `OcmAgent` CR) which contains the agent's OCM access token.
- A `Service` (named `ocm-agent`) which serves the OCM Agent API
- A `NetworkPolicy` to only grant ingress from specific cluster clients.
- A `ServiceMonitor` (named `ocm-agent-metrics`) which makes sure that the OCM Agent metrics can be exposed to Prometheus

The controller watches for changes to the above resources in its deployed namespace, in addition to changes to the cluster pull secret (`openshift-config/pull-secret`) which contains the OCM Agent's auth token.

The OCM Agent Controller is also responsible for creating/removing `ConfigMap` resource (named `ocm-agent`) in the `openshift-monitoring` namespace.

This resource is used by the [configure-alertmanager-operator](https://github.com/openshift/configure-alertmanager-operator) to appropriately configure AlertManager to communicate to OCM Agent.

The `ConfigMap` contains the following items:

| Key | Description | Example |
| --- | --- | --- |
| `serviceURL` | OCM Agent service URI | <http://ocm-agent.openshift-ocm-agent-operator.svc.cluster.local:8081/alertmanager-receiver> |

### cluster proxy support

The OCM Agent Controller will monitor the cluster proxy setting
and inject the HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment
variables to the OCM Agent deployment automatically based on the
values of the proxy/cluster object.
