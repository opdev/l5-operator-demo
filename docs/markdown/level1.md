### Operand

<a href="https://bestie-rescue.herokuapp.com/" rel="noopener noreferrer" target="_blank"><img src="images/bestie.png" width="70%"></a>

---

### Operator SDK
- leverage the kubernetes controller runtime
- integrate with Operator Lifecycle Manager(OLM)
- metrics with Go-based operator utilizing the Prometheus operator


<aside class="notes">
	- we used op sdk to orchestrate our op came with lots out of the box like
	<br>
	- Operator SDK is a framework that uses controller-runtime library which (simplifies building,
	testing, packaging ops)
	<br>
	- Integration with Operator Lifecycle Manager (OLM) to streamline packaging, installing, and running Operators on a cluster
	<br>
	- Metrics set up automatically in any generated Go-based Operator for use on clusters where the Prometheus Operator is deployed
</aside>


---

<div class="r-stack">
  <span class="fragment fade-out" data-fragment-index="0">
	<h3>Desired State</h3>
	<ul>
		<li>Deployment</li>
		<li>Service</li>
		<li>Route/Ingress</li>
		<li>Postgres database</li>
		<li>Job</li>
	</ul>
  </span>
  <img class="fragment current-visible" data-fragment-index="0" src="images/bestie_k8s_black.png" >
</div>

<aside class="notes">
	- define the required resources our application needs to run, the desired state of our application
	- requirements ~ resource definition files application needs
	<br>
	- Crunchy Postgres Operator
	- out of the box
	<br>
	- seed database before app deployment
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
  image: quay.io/opdev/bestie
  maxReplicas: 10
  version: "1.3"
```

<aside class="notes">
	- an object that allows you to extend Kubernetes capabilities by adding any kind of API object useful for your application
	- custom kind that specifically pertains to our operand bestie application
	<br>
	- This is the sticky note that the controller watches and tries to copy whenever an event occurs
	<br>
	- here we specify size, maxReplicas, and the application image we want anas well as the version
</aside>

---

### Controller

<img src="images/operator.jpg" alt="operator diagram" width="70%">

<aside class="notes">
	- Control loop that watches the state of the current cluster and tries to bring it
	closer to the desired state that's declared in the resource definition files.
	<br>
	- An controller can be thought of as a person who is doing things by looking up sticky note/template/image/recipe that we provide(cr)
	<br>
	- example: the CR is created(event) it triggers the reconciler function in the controller which prompts the creation of resources (deploy, service, job) which are needed to help make application run
	<br>
	- EX: something happens(event) we install the operator (event) controller runtimes triggered controller which looks up the sticky note (cr) and does a something (reconcile),
	it calls reconciler function to check the status of our resource (deploy, job, postgrescluster) and if its not what we want it brings it closer to what we define in our resource definition file
</aside>


---

### Reconcilers

```
subReconcilerList := []srv1.Reconciler{
	srv1.NewPostgresClusterCRReconciler(r.Client, log, r.Scheme),
	srv1.NewDatabaseSeedJobReconciler(r.Client, log, r.Scheme),
	srv1.NewDeploymentReconciler(r.Client, log, r.Scheme),
	srv1.NewDeploymentSizeReconciler(r.Client, log, r.Scheme),
	srv1.NewDeploymentImageReconciler(r.Client, log, r.Scheme),
	srv1.NewServiceReconciler(r.Client, log, r.Scheme),
	srv1.NewHPAReconciler(r.Client, log, r.Scheme),
	srv1.NewRouteReconciler(r.Client, log, r.Scheme),
}

```

<aside class="notes">
	- here is a snippert of the reconciler function being called by our controller
	- sub-reconiler === more organized, broken into sub-reconciles for each resource we need
	<br>
	- goes through each reconciler function and checks whether the resources are what we defined them to be in their resource definition files
	<br>
	- Reconciliation is level-based, meaning action isn't driven off changes in individual Events, but instead is driven by actual cluster state read from the apiserver or a local cache. It's not taking note if a Pod Delete Event, the Request won't contain that, instead the reconcile function observes this when reading the cluster state and see that the Pod as missing.
</aside>

---

### Demo

<iframe width="560" height="315" src="https://www.youtube.com/embed/0UF2Khwc0AU" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

<aside class="notes">
	- so to demonstrate, here is our operator with a level one capability of basic installation.
	<br>
	- in the cluster all you need to do is find our operator and install it
	- apply the CR
	- and your done, now you can view the application
	<br>
	- you can see the pgo cluster be created and set before the job runs and the deployment only creating once the job has finished
	- we specify all this in our reconcilers
</aside>

---


### Overview

<img src="images/operator.jpg" alt="operator diagram" width="70%">

<aside class="notes">
	- Summary: Now the l5 operator has level 1 capabilites it has the controller and the CR, so it can automatically provision and configure all the resource we need for the flask application upon installation
	<br>
	- Next Sid will talk more about the capabilities of a level 2 operator and how
	we brought the l5 operator from level 1: basic installation to level 2: seamless upgrades
</aside>
