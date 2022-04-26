#### Skaffolding go project

implement basic operator
report app version in status
use bestie as an operand

field for s3 bucket / object storage in cr
create deployment for bestie
create service for bestie
make sure postgresql db is up and running before bestie deployment
create routes for bestie
only seed if there's no data.
document prerequisite for https i.e. certificate manager
---

#### Custom Resource
A paragraph with some text and a [link](http://hakim.se).
---
#### Controller

```
isOpenShiftCluster, err := verifyOpenShiftCluster(routev1.GroupName, routev1.SchemeGroupVersion.Version)
	if err != nil {
		return ctrl.Result{}, err
	}
```

---
#### Operator
