
Install
<p class="fragment fade-up">Self Heal</p>
<p class="fragment fade-up">Scale</p>
<p class="fragment fade-up">Update</p>
<p class="fragment fade-up">Backup</p>
<p class="fragment fade-up">Clean Up</p>


<aside class="notes">
As we have heard from Manna and Sid,
Operators encapsulate operational tasks in code so we can manage application lifecycle actions using Kubernetes APIs.
We automate 
Install -
Self heal -
scale -
update -
backup -
clean up - 

So far so good, but what if during the upgrade the pod keeps cycling itself because it's got a bad image. And let's say that the upgrade happens automatically thru CI/CD pipeline and this pipeline doesn't manage image's health. So it would be nice for the operator to have obserbality into the pods. This can be as simple as the operator detecting upgrade failure thus sends an alert to fix the image.

</aside>
---

#### Level 4 - Deep Insights

<p class="fragment fade-in"> Monitoring</p>
<p class="fragment fade-in"> Alert</p>



<aside class="notes">


This brings us to level 4 
deep insights.

Monitoring the operator and operand's health as well as
monitoring operand's performance
There are several levels of alerts, such as info, warning, and critical. For each of the levels, you can map to a different protocal such as email, slack, etc. 
Even emit custom events

</aside>

---

### roadmap

---

### Demo

<aside class="notes">
- Export metrics to Prometheus Operator
- Showcase metrics in Grafana Operator
</aside>

---

### Wow! I want this, but how do I get it?

<p class="fragment fade-in"> Controller-runtime builds a global prometheus registry and publishes a collection of performance metrics for each controller.</p>

<aside class="notes">
As you recall, Manna told us that Metrics set up automatically in any generated Go-based Operator for use on clusters where the Prometheus Operator is deployed
</aside>

---
### Demo controller metrics
---

### grant permission
- Metrics are protect by kube-rbac-proxy

<aside class="notes">
- Grant permission to your Prometheus server so that it can scrape the protected metrics.
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
      - nodes
      - secrets
    verbs:
      - get
      - list
      - watch

```
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
---

### Export Metrics to Promethus

<p class="fragment fade-up">Scrape metrics with ServiceMonitor resource</p>

<aside class="notes">
- Operator
  - Operator-sdk create the ServiceMonitor to export metrics
- Operand
  - Create in the operand namespace

  Explain what scrape metric means
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
---

```
All wonderful, 
    but I want to create CUSTOM METRICS
```

---

Publishing Custom Metrics to Prometheus

<aside class="notes">
Declare collectors as global variables

Register them using init() in the controller's package

Apply using methods provided by the prometheus client

</aside>

---

- Declare collectors as global variables

```
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
```
---

- Register

```
func init() {
  metrics.Registry.MustRegister(ApplicationUpgradeCounter, ApplicationUpgradeFailure)
}

```
<aside class="notes">
Register them using init() in the controller's package
</aside>
---

```
bestie_metrics.ApplicationUpgradeCounter.Inc()

```
<aside class="notes">
Everytime the operand is upgraded, the bestie_upgrade_counter increases by 1
</aside>
---

```
rc := getPodstatusReason(nonTerminatedPodList)
bestie_metrics.ApplicationUpgradeFailure.Set(rc)

```

```
func getPodstatusReason(pods []corev1.Pod) float64 {
...  
if string(pod.Status.Phase) == "Pending" {
    if string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ErrImagePull" ||
        string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ImagePullBackOff" {
        return 1
    }
}
return 0

```
<aside class="notes">
- Set the state for the operand upgrade
</aside>

---

Alert

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
Alert tells us when things goes wrong and get operator to take care of it.

Best-practice

Avoid Over-alerting
Select use case-specific alerts
</aside>

---

Emit custom events

```
if string(pod.Status.Phase) == "Pending" {
    if string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ErrImagePull" ||
        string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ImagePullBackOff" {
        //do something here with event
        return 1
    }
}
return 0

```
<aside class="notes">
When you introduce a change that breaks production, you should have a plan to roll back that change.

</aside>

---

Exposing Operand's metrics

- health
- performance

... by adding /metrics route to the application

<aside class="notes">
( bestie-route-bestie.apps.demo.opdev.io/metrics )

bestie_http_request_total counter

before you can scrape operand metrics, you need to add /metrics to the operand

</aside>

---

bestie_http_request_total{method="GET",status="200"} 375945.0

<p class="fragment fade-up">bestie_http_request_total{method="POST",status="405"} 13.0</p>

<p class="fragment fade-up">bestie_http_request_total{method="POST",status="200"} 1.0</p>

<p class="fragment fade-up">bestie_http_request_total{method="HEAD",status="200"} 1.0</p>

<aside class="notes">
( bestie-route-bestie.apps.demo.opdev.io/metrics )

bestie_http_request_total counter

before you can scrape operand metrics, you need to add /metrics to the operand

</aside>

---

### End


<aside class="notes">
introduce Yuri
</aside>

