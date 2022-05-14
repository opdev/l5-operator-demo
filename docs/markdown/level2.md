####
- Operator
- Operand (Our Application)

<aside class="notes">
  Operator upgrades can be configured to be done automatically via the Operator Lifecycle Manager. The version of the Operand is controlled by a field in our operators Custom Resource
</aside>

---
####
![Free Lunch](images/freelunch.jpeg)

---
#### Minor versions
As easy as updating the image in our deployment

<aside class="notes"> 
  Since the "user interface" of our operator is the CR we can perform application updates by updating the our custom resource.
</aside>

---
#### Spec
```
apiVersion: pets.bestie.com/v1
kind: Bestie
metadata:
  name: bestie
spec:
  size: 3
  image: quay.io/opdev/bestie
  version: "1.3"
```

<aside class="notes">
  Both the application image as well the version are exposed in the CR  We get the desired version from custom resource and update the pod template in the deployment image if current version is different than the desired version
</aside>

---
#### Status
```

```

<aside class="notes">
  Both the application image as well the version are exposed in the CR  We get the desired version from custom resource and update the pod template in the deployment image if current version is different than the desired version
</aside>

---
#### Seamless*
Liveness and Readiness probes ensure that your image is rolled out only if it is healthy

<aside class="notes">
  Failing liveness probe will restart the container, whereas failing readiness probe will stop our application from serving traffic.</aside>


---
#### 
Seamless but conditions apply*

<aside class="notes"> 
  What if you have a bad version ?
  What if your container image does not start correctly ?
</aside>

---
#### 
What if there are incompatible changes ?
[Incompatible Changes](images/incompatible_changes.png)

<aside class="notes"> 
  A bit more sophistication is needed requiring level 3 capabilities
  Rollouts will be seamless as long as there are no breaking database changes
</aside>

