
Life is full of surprises
<img src="images/life-issurprise.png" width="85%" alt="Prometheus and Grafana monitors and displace metrics">

<aside class="notes">

Thank you Sid.

- Life is full of surprises so for everything else we use prometheus and grafana.

At this point, life is good. Bestie operator is now a capability - level 3  operator.
- seamless upgrades
- full lifecycle

What else could we ask for ?
Well, alot more!!!
we have guardrails built around the deployment so the website is always available even if our upgrades fails.

because of this, we have to manually observe the podstatus or bring up the website to know if the upgrade worked or not.

so this brings us to Level 4 - Deep Insights.

- metrics, alerts, log processing and workload analysis.
</aside>

---

We want bestie operator to

- Setup full monitoring and alerting for the operand
- Expose metrics about its health
- Exposes health and performance metrics about the Operand
- Aggregate metrics using Prometheus and visualize using Grafana.


<aside class="notes">

1. How to expose metrics

2. Create Alerting rules 

3. Demo

</aside>
---

Expose operator metrics
- Create ServiceMonitor

<img src="images/ServiceMonitor.jpg" width="85%" alt="servicemonitor flowchart">

<aside class="notes">

What is a Servicemonitor

- ServiceMonitor describes the set of targets to be monitored by Prometheus 
- setup servicemonitor for the operand
- can anyone tell me why we don't need to setup a servicemonitor for the operator?
- the servicemonitor needs to be in the same namespace as the bestie application 

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

- servicemonitor monitor bestie application metrics
- coming from bestie website
- bestie-route-bestie.apps.demo.opdev.io/metrics

- Most of the servicemonitor yaml manifest is the same 
  except for
  - Spec.endpoint.port. It is the service's port name
  - Spec.Selector.matchlabels. It has to match the service's label

</aside>

---

Expose operator metrics
- Grant permission to Prometheus server

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

The operator-framework use kubebuilder to scalfold operator
so metrics are protected by kube-rbac-proxy by default
because of this, we need to grant permissions to the Prometheus server so that it can scrape the protected metrics.
To achieve this, create a clusterrole

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

and a clusterrolebinding to bind

clusterrole : prometheus-k8s-role to 

the serviceaccount : prometheus-k8s in the openshift-monitoring namespace

</aside>
---

```
kubectl label namespace <operator_namespace> \
openshift.io/cluster-monitoring="true"

```


<aside class="notes">

Set the labels for the namespace that you want to scrape

The label

openshift.io/cluster-monitoring="true"

enables OpenShift cluster monitoring for that namespace

For bestie, we added a label in the 

1. operator namespace
2. bestie application namespace

</aside>
---

Publish custom metrics

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

1. Declare collectors as global variables.

   - ApplicationUpgradeCounter
   - ApplicationUpgradeFailure are collectors.

4 types of metrics 
- Counter, Gauge, histogram, and summary

ApplicationUpgradeCounter is a counter, it's value can only increase or be reset to 0 on restart.
- it tracks the number of successful operand upgrades processed.

ApplicationUpgradeFailure is a gauge, is a number that can go up or down.
- it tracks the pod's image status.
  When the pod status keeps cycling and never reach the complete status. We look at the waiting.Reason. If it's ErrImagePull or ImagePullBackoff, then we set

Best practice, prefix metric with name of the operator

   - bestie_upgrade_counter
   - bestie_upgrade_failure


</aside>
---

```
import "github.com/prometheus/client_golang/prometheus"
...
ApplicationUpgradeCounter.Inc()

```

<aside class="notes">

- Increase the counter
- use Inc() method provided by the prometheus client-go package

</aside>
---

```
ApplicationUpgradeFailure.Set(rc)

```
<aside class="notes">

- Use the Set() method
- Set the state for the operand upgrade

Look at the code

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

### Metrics for thoughts

-  Number of visitors
-  What pages viewed
-  Operation performed ( success or failed )
-  How many pets were adopted

<aside class="notes">

Number of visitors to the site
What pages were viewed
What operations did the user do - search pet/insert add/updated add/etc. And did it succeed or failed
How many pets were adopted and their details as labels (pet type:dog/cat/etc, time the add was listed/ age of the pet/ gender, has picture:yes/no)


</aside>

