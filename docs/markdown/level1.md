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
