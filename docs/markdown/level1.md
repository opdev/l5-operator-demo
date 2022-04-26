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

```
isOpenShiftCluster, err := verifyOpenShiftCluster(routev1.GroupName, routev1.SchemeGroupVersion.Version)
	if err != nil {
		return ctrl.Result{}, err
	}
```

<aside class="notes">
	Speaker note
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
