#### Creating Resources
- Deployment
- Service
- Route

<aside class="notes">
	YAML --> refactored --> Go
</aside>

---

#### Custom Resource

[Bestie](https://bestie-rescue.herokuapp.com/) is a pet adoption web application with a React frontend and Flask backend utilizing a Postgres database.

<aside class="notes">
	Speaker note
</aside>

---

#### Controller

Control loop that watches the state of the current cluster and tries to bring it closer to the desired state that's declared in the YAML files.


<aside class="notes">
	In Kubernetes, controllers are control loops that watch the state of your cluster, making or requesting changes. Each controller tries to move the current cluster state closer to the desired state.
</aside>

---

#### Reconciler


<aside class="notes">

</aside>

---

#### Demo


<aside class="notes">

</aside>

<!--
implement basic operator
report app version in status
use bestie as an operand

field for s3 bucket / object storage in cr
create deployment for bestie
create service for bestie
make sure postgresql db is up and running before bestie deployment
create routes for bestie
only seed if there's no data.
document prerequisite for https i.e. certificate manager -->
