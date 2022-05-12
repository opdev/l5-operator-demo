
#### Level 4 - Deep Insights

- Monitoring

- Alerting

<aside class="notes">

This brings us to Level 4 - Deep Insights.

Includes Monitor and Alert.

This includes full monitoring for the operand.

- expose a health metrics endpoint?

- expose Operand performance metrics

and alerts.

expose Operand alerts

</aside>

---

<img src="images/prometheus-grafana.png" width="45%" alt="Grafana Dashboard">

<aside class="notes">

Prometheus is an open-source for monitoring and alerting toolkit

and Grafana dashboard can display the metrics and the alert using the alerting rules.

</aside>

---

### Roadmap

1. Demo
2. Expose operator metrics
3. Alert
4. Expose operand metrics

---

### Demo

<aside class="notes">
- Export metrics to Prometheus Operator
- Showcase metrics in Grafana Operator
</aside>

---

### Expose Custom Metrics

1. Enable Prometheus
2. Create custom controller class for metrics
3. Record collectors
4. Grant permission
5. Set the labels
---

1. To enable prometheus monitor, uncomment all sections with 'PROMETHEUS' in config/default/kustomization.yaml

```
../prometheus

```
<aside class="notes">
As you recall, Manna told us that Metrics set up automatically in any generated Go-based Operator for use on clusters where the Prometheus Operator is deployed

Prerequisite : Operator-sdk

</aside>

---

2. Create a custom controller class to publish additional metrics from the Operator.

```
package bestie_metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

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

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(ApplicationUpgradeCounter, ApplicationUpgradeFailure)
}
```
<aside class="notes">
Declare collectors as global variables

Register them using init() in the controller's package

Apply using methods provided by the prometheus client

</aside>
---

3. Record to these collectors from any part of the reconcile loop in the main controller class, which determines the business logic for the metric.

```
bestie_metrics.ApplicationUpgradeCounter.Inc()

```

```
rc := getPodstatusReason(nonTerminatedPodList)
bestie_metrics.ApplicationUpgradeFailure.Set(rc)
...
func getPodstatusReason(pods []corev1.Pod) float64 {
  
if string(pod.Status.Phase) == "Pending" {
    if string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ErrImagePull" ||
        string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ImagePullBackOff" {
        return 1
    }
}
return 0

```
<aside class="notes">

- Everytime the operand is upgraded, the bestie_upgrade_counter increases by 1.

- Set the state for the operand upgrade
</aside>
---

4. Create role and role binding definitions to allow the service monitor of the Operator to be scraped by the Prometheus instance of the OpenShift Container Platform cluster.

<pre><code data-trim data-noescape>
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
</code></pre>

---

<pre><code data-trim data-noescape>
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

</code></pre>
---

5. Set the labels for the namespace that you want to scrape, which enables OpenShift cluster monitoring for that namespace:


```
oc label namespace <operator_namespace> openshift.io/cluster-monitoring="true"
```
---

### alert

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

<img src="images/zero-alert.png" width="100%" alt="no alerts">

---

<img src="images/alert-critical.png" width="100%" alt="alert">

<aside class="notes">
Alert tells us when things goes wrong and get operator to take care of it.

Best-practice

Avoid Over-alerting
Select use case-specific alerts
</aside>

---

### Expose Operand metrics

1. Add /metrics to the application

<aside class="notes">

in addition to exposing custom metrics you can also
expose operand's metrics by adding /metrics path to your application 

http://bestie-route-bestie.apps.demo.opdev.io/metrics


- Operator
  - Operator-sdk create the ServiceMonitor to export metrics
- Operand
  - Create in the operand namespace

  ServiceMonitor describes the set of targets to be monitored by Prometheus
</aside>

---

2. Create service monitor in the operand namespace.

<img src="images/servicemonitor.png" width="45%" alt="servicemonitor flowchart">

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

### End

<aside class="notes">
introduce Yuri
</aside>

