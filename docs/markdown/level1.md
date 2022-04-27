### Operand & Operator

<div style="text-align: left">
Operand: <a href="https://bestie-rescue.herokuapp.com/">Bestie</a> application

Operator: Automate the management of the Bestie application
</div>

<aside class="notes">
	Operator and Operand
	App with a React frontend and Flask backend utilizing a Postgres database.
	What we are trying to solve
	automate management of this applciation
</aside>

---

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

### Operator SDK Go-opertator
- Installation
- Out of the box content

<aside class="notes">
	orchestrated in some way we use an operator to build we use sdk
	kubernetes and opeator framework
	leverage oeprator sdk incorprate
	conroller and customer resourse

	Operator SDK is a framework controller-runtime library (simplies building, testing, packaging ops), OLM(streamline packaging, install, manage, ugrade ops on a cluster), catalog (publish/share)
</aside>

---

### Controllers & Reconiler

```
log.Info("reconcile postgres if it does not exist")
pgo := &pgov1.PostgresCluster{}

err = r.Get(ctx, types.NamespacedName{Name: BestieName + "-pgo", Namespace: bestie.Namespace}, pgo)
if err != nil {
	if errors.IsNotFound(err) {
		log.Info("Creating a new PGC for bestie")
		fileName := "config/resources/postgrescluster.yaml"
		err := r.applyManifests(ctx, bestie, pgo, fileName)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
		}
	} else {
		return ctrl.Result{Requeue: true}, err
	}
}
```

<aside class="notes">
	- Control loop that watches the state of the current cluster and tries to bring it
	closer to the desired state that's declared in the resource definition files.
	- In Kubernetes, controllers are control loops that watch the state of your cluster,
	making or requesting changes. Each controller tries to move the current cluster state
	closer to the desired state.
	- Contains the controllers and is triggered every time an event occurs
	- Crunchy Postgres Operator uses the YAML we provide and configures a postgres cluster
	for us to consume
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
	- Next Soundharya will talk more about the capabilities of a level 2 operator and how we brought the l5 operator from level 1 to level 2
</aside>

---

<!-- ### Demo

![]()

<aside class="notes">
	Show video and voice over.


---
</aside> -->

<!--

create deployment for bestie
create service for bestie
make sure postgresql db is up and running before bestie deployment
create routes for bestie
only seed if there's no data.
document prerequisite for https i.e. certificate manager -->
