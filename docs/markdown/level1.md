### Requirements
- Deployment
- Service
- Route/Ingress
- Postgres database
- Job


<aside class="notes">
	YAML --> refactored --> Go
	we need database deployment, job, route etc

	Crunchy Postgres Operator
</aside>

---

### Operator SDK
- Installation
- Out of the box content

<aside class="notes">
	orchestrated in some way so we use op sdk
	Operator SDK is a framework w/ controller-runtime library (simplies building,
	testing, packaging ops)
	leverage oeprator sdk incorprate
	controller and customer resourse
</aside>


---

### Custom Resource
```
apiVersion: pets.bestie.com/v1
kind: Bestie
metadata:
  name: bestie
spec:
  size: 3
  image: quay.io/mkong/bestiev2
  maxReplicas: 10
  version: "1.3"
```

<aside class="notes">
	- This is the sticky note that the controller watches and tries to copy whenever an event occurs
</aside>

---

### Controller & Reconciler




<aside class="notes">
	- Control loop that watches the state of the current cluster and tries to bring it
	closer to the desired state that's declared in the resource definition files.
	- An controller is basically a human who is running commands based on a sticky note (cr)
	for example: something happens (event) the controller looks up the sticky note (cr) and does a thing (reconcile)
	- Next Soundharya will talk more about the capabilities of a level 2 operator and how
	we brought the l5 operator from level 1: basic installation to level 2: seamless upgrades
</aside>

