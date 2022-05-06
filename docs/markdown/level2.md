#### Upgrades
- Operator
- Operand (Our Application)

<aside class="notes">
Operator upgrades can be configured to be done automatically via the Operator Lifecycle Manager
The version of the Operand is controlled by a field in our operators Custom Resource
</aside>

---
#### Minor versions
- As easy as updating the image in our deployment

<aside class="notes"> 
Since the "user interface" of our operator is the CR we can perform application updates by updating the our custom resource.
We get the desired version from custom resource and update the pod template in the deployment image if current version is different than the desired version 
</aside>

---
#### What if you have a bad version ?

---
#### K8s has your back
- Liveness and Readiness probes
- Ensure that your image is rolled out only if it is healthy

<aside class="notes"> Failing liveness probe will restart the container, whereas failing readiness probe will stop our application from serving traffic.</aside>

---
#### What if there are incompatible changes ?
- Rollouts will be seamless as long as there are no breaking database changes

<aside class="notes"> 
A bit more sophistication is needed requiring level 3 capabilities
</aside>
