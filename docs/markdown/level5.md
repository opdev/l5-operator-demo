#### You deserve a good night's sleep

![Autopilot](images/dreamworks.png)

---

#### Can your operator:
- Read metrics?
- Auto-scale?
- Move Workloads?
- Restart internal resources?

<aside class="notes">
  Operator upgrades can be configured to be done automatically via the Operator Lifecycle Manager. The version of the Operand is controlled by a field in our operators Custom Resource
</aside>

---
#### Enabling HPA
- Set MaxReplicas to activate it.

<pre><code data-trim data-noescape>
apiVersion: pets.bestie.com/v1
kind: Bestie
metadata:
  name: bestie
spec:
  size: 3
  image: quay.io/opdev/bestie
  maxReplicas: 10
  version: "1.3"
</code></pre>

---
#### HPA Workflow

![hpa](images/HPA-Diagram.jpeg)

---
#### Load Test Demo

```
kubectl run -i --tty load-generator --rm --image=busybox:1.28 --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://bestie-service; done"
```

---
#### Custom Metrics
- `apiVersion: autoscaling/v2`
- Any ideas?
	1. Requests per second?
	2. HTTP error rate?
	3. Number of restarts?

---
#### AI in Operator?

- Learn the Performance Baseline