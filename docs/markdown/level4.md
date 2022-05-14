
Life is full of surprises
<img src="images/life-issurprise.png" width="85%" alt="Prometheus and Grafana monitors and displace metrics">

<aside class="notes">

Thank you Sid.

- Life is full of surprises so for everything else we use prometheus and grafana.
- Working level 3 operator
- seamless upgrades
- full lifecycle
- something unexpected
- upgrade bestie application version
- change manifest and apply
- a while goes by
- manager, upgrade didn't work
- pod status, in the Waiting state because the image cannot be pulled because I had the wrong version
- after I applied the cr with the correct version number the upgrade succeded.

so this brings us to Level 4 - Deep Insights.

- metrics, alerts, log processing and workload analysis.

- implement level 4 cabability
- detect errors and anomality early and create alert support if needed.
- full monitoring of the operator and the operand. 
- Create metrics and alerts accordingly.

- Prometheus to stores both the metrics and the alerting rules
- grafana produces meaninful and customizable dashboards to display metrics and alerts by using prometheus data.

1. expose metrics.

2. Create Alert based on metrics 

3. Demo

</aside>
---

Exposing operator metrics

<img src="images/ServiceMonitor.jpg" width="85%" alt="servicemonitor flowchart">

<aside class="notes">

- Servicemonitor
- ServiceMonitor describes the set of targets to be monitored by Prometheus 
- scrape metrics from the service.
- operator-service target
- operand-service target
- can anyone tell me why we don't need to setup a servicemonitor to scrape operator's metrics?
- in what namespace apply servermonitor yaml manifest
- same namespace as the service

</aside>

---

```
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: bestie-servicemonitor
  labels:
    name: bestie-servicemonitor
spec:
  endpoints:
    - path: /metrics
      port: metrics
      scheme: http
  selector:
    matchLabels:
      app: bestie
```
<aside class="notes">

- servicemonitor to scrape bestie application metrics

- expose application's metrics with /metrics path to the application's URL.

- Spec.endpoint.port is the service's port name
- Spec.Selector.matchlabels needs to match the service's label


</aside>

---

```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus-k8s-role 
rules:
  - apiGroups:
      - ""
    resources:
      - endpoints
      - pods
      - services
    verbs:
      - get
      - list
      - watch
```

<aside class="notes">

- operator-framework use kubebuilder to scalfold operator
- metrics are protected by kube-rbac-proxy by default
- grant permissions to the Prometheus server so that it can scrape the protected metrics.
- To achieve this
- clusterrole

</aside>
---
```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus-k8s-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-k8s-role
subjects:
  - kind: ServiceAccount
    name: prometheus-k8s
    namespace: openshift-monitoring
```

<aside class="notes">

- and a clusterrole binding to bind
- clusterrole to prometheus-k8s serviceaccount in the openshift-monitoring namespace

</aside>
---

```
kubectl label namespace <operator_namespace> \
openshift.io/cluster-monitoring="true"

```


<aside class="notes">
also create a label for each namespace where the servicemonitor is scraping the metrics.

we created a label in the operator namespace and another lable in the operand's namespace

</aside>
---

```
var (
	ApplicationUpgradeCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "bestie_upgrade_counter",
			Help: "Number of successful bestie application upgrades processed",
		},
	)
	ApplicationUpgradeFailure = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bestie_upgrade_failure",
			Help: "1 if ImagePullBackOff, otherwise 0",
		},
	)
)
```

<aside class="notes">
To publish the metrics, we first need to
Declare collectors as global variables.

Best practice, prefix metric with name of your operator
Easy search for your metric

4 types of metrics, 
Counter, Gauge, histogram, and summary

A counter is a cumulative metric that represents a single monotonically increasing counter whose value can only increase or be reset to zero on restart. For example, you can use a counter to represent the number of requests served, tasks completed, or errors.

A gauge is a metric that represents a single numerical value that can arbitrarily go up and down.

Gauges are typically used for measured values like temperatures or current memory usage, but also "counts" that can go up and down, like the number of concurrent requests.

bestie_upgrade_counter : type : counter
bestie_upgrade_failure : type : gauge



register the collectors with metrics.registry.mustregister before using it.


</aside>
---

```
bestie_metrics.ApplicationUpgradeCounter.Inc()
```

```
bestie_metrics.ApplicationUpgradeFailure.Set(rc)
```
<aside class="notes">

bestie_upgrade_counter metric is a counter that tracks the number oof successful operand upgrades processed.

To increase the metric by increasing the collector, ApplicationUpgradeCounter.Inc().

bestie_upgrade_failure metric tracks the status of the pod's image status.
When the pod status keeps cycling and never reach the complete status. We look at the waiting.Reason. If it's ErrImagePull or ImagePullBackoff, then we set
the ApplicationUpgradeFailure.Set(rc)

and we wanted to create a metric for the upgrade's pod status.

 created bestie_upgrade_failure, a gauge, set a number. 0 is Upgrade not found image issue and 1 is bad image found.
 

- Everytime the operand is upgraded, the bestie_upgrade_counter increases by 1.

- Set the state for the operand upgrade
if string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ErrImagePull" ||
        string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ImagePullBackOff" {
        return 1
    }
</aside>
---

```
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: bestie-alert
spec:
  groups:
  - name: example
    rules:
    - alert: BestieImageFailureAlert
      expr: bestie_upgrade_failure{job="l5-operator-controller-manager-metrics-service"} == 1
      labels:
        severity: critical
```
<aside class="notes">
Use metrics to create Alert rules to produce alerts
</aside>
---

### Demo

<iframe width="560" height="315" src="https://www.youtube.com/embed/9JYMSE4TZj8" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

---

### What's next for metrics and alerts

<aside class="notes">
Add metrics -

Number of visitors to the site

What pages were viewed

What operations did the user do - search pet/insert add/updated add/etc. And did it succeed or failed

How many pets were adopted and their details as labels (pet type:dog/cat/etc, time the add was listed/ age of the pet/ gender, has picture:yes/no)

Integrate with AlertManager. Routes alert to email, slack, etc.

</aside>

