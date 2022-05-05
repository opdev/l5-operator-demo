#### recap

Operators encapsulate operational tasks ( Sticky notes ) in code, so we can manage application lifecycle actions using Kubernetes APIs.  

Sticky notes :
- how to build/install ( level 1 )
- how to scale it ( level 1 ) 
- how to upgrade it ( level 2 ) 
- how to recover from failure scenarios ( level 3 )

---

<aside class="notes">

So what additional skills a Kubernetes Operator performing manual task need to have.
Well, there's the hidden stuff, not as flashy or fun stuff that sometimes called the plumbing.
It's the networking parts ( level 1 ), it's the traffic management stuff ( level 1 ), It's also the visibility and monitoring ( Level 4 )

one thing I want to focus in (pause) on a little bit is the visibility and observability tools that you need to have for Kubernetes because it is a different paradigm. As an example, you were giving of a Pod keeps recycling itself because the pod is starting, crashing, starting again, and then crashing again or it’s got running out of memory. You need a way to pick up and realize that’s bad. Undesirable behavior.

( sticky note)

 So a lot of it is you have to get the visibility, but then you have to learn and understand enough that you know what to look for. You know the actual signs to look for, to understand what the symptoms are.
</aside>

#### Level 4 - Deep Insights

- Monitoring
    - Operator exposing metrics about its health
    - Operator exposes health and performance metrics about the Operand

- Alert
    - Operand send alerts
    - Emit custom events

---

### Monitoring

### Operator exposing metrics about its health

- Configure custom metric
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

- Register custom metrics with the global prometheus registry
```
metrics.Registry.MustRegister(ApplicationUpgradeCounter, ApplicationUpgradeFailure)
```

- Increment custom metrics counter
```
bestie_metrics.ApplicationUpgradeCounter.Inc()

```

- Set pod's state
```
bestie_metrics.ApplicationUpgradeFailure.Set(rc)
...

if string(pod.Status.Phase) == "Pending" {
    if string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ErrImagePull" ||
        string(pod.Status.ContainerStatuses[0].State.Waiting.Reason) == "ImagePullBackOff" {
        return 1
    }
}
```

### Operator exposes health and performance metrics about the Operand

- Application must add /metrics to the path
- Add servicemetrics to the same namespace and the operand


### Prometheus to store the metrics, and Grafana's dashboard to see the metrics


### Alert

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

### Emit custom events

```
bestie_metrics.ApplicationUpgradeFailure.Set(rc)

```
<aside class="notes">
When you introduce a change that breaks production, you should have a plan to roll back that change.

</aside>