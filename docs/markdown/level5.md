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
#### Demo

---
### Demo

<iframe width="560" height="315" src="https://www.youtube.com/embed/U3yelj0avfY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen>
</iframe>

---
#### Custom Metrics
- `apiVersion: autoscaling/v2`

- Any ideas?

	1. Requests per second?
	2. HTTP error rate?
	3. Number of restarts?

---
#### Take Aways 
- Operator is able to Deploy the operand application.
- It's possible to perform seamless upgrades to the operator and operands.
- Backup and Restore are in place.
- The operator as well as the operands expose metrics, which are aggregarated using Prometheus.
- The operator is able to autoscale based on the application loop.

---
#### For the near Future:
- Scaffolding a L5 Operator with all capabilities in place.
- Operator will take action in autopilot mode based on what it learned from Performance Baseline.
