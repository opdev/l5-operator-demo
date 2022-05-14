
Life is full of surprises
<img src="images/life-issurprise.png" width="85%" alt="Prometheus and Grafana monitors and displace metrics">

<aside class="notes">

Thank you Sid.

We begin with the theme, Operator is full of surprise for everything else there is prometheus and grafana.
Here we are, very happy with how far we have come with the Bestie operator. It's capable of seamless upgrades and full lifecycle, but just when we thought this is all the automation we need, something caught our attention. I was asked to upgrade Bestie operator to the bestie app image version 1.4.0. Easy right!.  I edited the cr to change the spec.version to from 1.3.0 to 1.4.0 and I thought nothing of it.
Seemless upgrade. A few days go by, my manager comes to me and says, Rose, did you do the upgrade like you promised ?
I was like yeah and then he said, well, it doesn't look like it worked. Hmmm, i said, let me do some digging. I'm sure it's something stupid I did on my part. so I checked the pods status and there it is, in plain sight, the pod stays in the Waiting state because the image cannot be pulled because I had the wrong version number.  So, after I applied the cr with the correct version number the upgrade succeded.

so this brings us to Level 4 - Deep Insights.

The proper definition is metrics, alerts, log processing and workload analysis.

We decided to implement level 4 cabability so we can catch errors early and alert support if needed. We wanted full monitoring of the operator and the operand. Create metrics and alerts accordingly.

We use Prometheus to stores both the metrics and the alerting rules, while grafana produces meaninful and customizable dashboards to display metrics and alerts by using prometheus data.

1. we're going to talk about how we expose metrics.

2. Alert is created when the condition of the alerting rule is meet, so we're going to see an example of the alerting rule

3. I will share with you a Demo that I have prepared about metrics and alerts.
</aside>
---

### Expose Operator Metrics

<img src="images/ServiceMonitor.jpg" width="85%" alt="servicemonitor flowchart">

<aside class="notes">
Prometheus's servicemonitor allows prometheus to scrape metrics from the service.

so we need to setup a servicemonitor to scrape operand's metrics.

can anyone tell me why we don't need to setup a servicemonitor to scrape operator's metrics?

in what namespace is the servicemonitor in?

the answer is with the service you want to scrape the metrics from.

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
besides implementing the servicemonitor, we also had to

updated the bestie app to expose application's metrics and create /metrics path to the application's URL.

Spec.endpoint.port is the service's port name and
Spec.Selector.matchlabels needs to match the service's label


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
For security, rbac is on by default.

the metrics are protected by rbac rules.

so we need to create a clusterrole 
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
and a clusterrole binding to bind
the clusterrole to the prometheus-k8s serviceaccount in the openshift-monitoring namespace
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

