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

<!-- , OLM(streamline packaging, install, manage, ugrade ops on a cluster), catalog (publish/share) -->

---

### Custom Resource
```
apiVersion: pets.bestie.com/v1
kind: Bestie
metadata:
  name: bestie
spec:
  size: 3
  image: quay.io/opdev/bestie
  maxReplicas: 10
  version: "1.3"
```

<aside class="notes">
	- This is the sticky note that the controller watches and tries to copy whenever an event occurs
</aside>

---

### Reconciler

```golang
reconcilers := []reconcilers.Reconciler{
	reconcilers.NewPipelineGitRepoReconciler(r.Client, reqLogger, r.Scheme),
	reconcilers.NewPipeDependenciesReconciler(r.Client, reqLogger, r.Scheme),
	...
}

for _, r := range reconcilers {
	requeue, err := r.Reconcile(ctx, pipeline)
	if err != nil {
		log.Error(err, "requeuing with error")
		return ctrl.Result{Requeue: true}, err
	}
	requeueResult = requeueResult || requeue
}
return ctrl.Result{Requeue: requeueResult}, nil
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
	- Next Soundharya will talk more about the capabilities of a level 2 operator and how
	we brought the l5 operator from level 1 to level 2
</aside>

<!--

create deployment for bestie
create service for bestie
make sure postgresql db is up and running before bestie deployment
create routes for bestie
only seed if there's no data.
document prerequisite for https i.e. certificate manager
infinite loop run and watches something
controller is a loop watchs cr or crd
controller(loop) received event from cr
triggers recinciler(logic)

import from controller-runtime

manager initiates

main.go
-->
