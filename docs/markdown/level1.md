### Requirements
- Deployment
- Service
- Route/Ingress
- Postgres database
- Job


<aside class="notes">
	- resource definition files
	<br>
	Crunchy Postgres Operator
</aside>

---

### Operator SDK
- installation
- out of the box content
- controller-runtime library

<aside class="notes">
	- we used op sdk to orchestrate our op came with lots out of the box
	<br>
	- Operator SDK is a framework w/ controller-runtime library which (simplifies building,
	testing, packaging ops)
	- leverage operator sdk incorprate
	- custom resource & controller
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
	- an object that allows you to extend Kubernetes capabilities by adding any kind of API object useful for your application
	- CR we create containes the deploy, service job, needed to help make application run
	- This is the sticky note that the controller watches and tries to copy whenever an event occurs
</aside>

---

### Controller & Reconciler

<img src="https://i.imgur.com/G49iwt5.jpg" width="70%" alt="op diagram">

<aside class="notes">
	- Control loop that watches the state of the current cluster and tries to bring it
	closer to the desired state that's declared in the resource definition files.
	<br>
	- An controller is basically a human who is running commands based on a sticky note/template/image/recipe that we give it (cr)
	for example: something happens (event) the controller looks up the sticky note (cr) and does a thing (reconcile)
	<br>
	- EX: something happens(event) we install the operator (event) controller runtimes triggered controller which looks up the sticky note (cr) and does a something (reconcile),
	it calls reconciler function to check the status of our resources (deploy, job, postgrescluster) and if its not what we want it brings it closer to what we define in our resource definition file
	<br>
	- Summary: Now the l5 operator has level 1 capabilites it has the controller and the CR, so it can automatically provision and configure all the resource we need for the flask application upon installation
	<br>
	- Next Soundharya will talk more about the capabilities of a level 2 operator and how
	we brought the l5 operator from level 1: basic installation to level 2: seamless upgrades
</aside>
