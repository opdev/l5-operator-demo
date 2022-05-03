#### Seamless upgrade of the Operator and the Operand
<aside class="notes">
An upgrade of the operator automatically ensures the instantiated resources for each CR are in the new desired state and which would upgrade the operand.
Operator can be upgraded seamlessly and can either still manage older versions of the operand or update them.
</aside>

---

#### Workflow

- Implement version upgrade in the operator controller
- Deploy the operator
- Update the CR version and apply
- Test the version upgrade

---

#### Compare the current version of the CR instance with the Container Image
<aside class="notes"> 
 Get the desired version from CR
 Upgrade the container image if current version is less than desired version </aside>
---

####

```
bestieImageDifferent := !reflect.DeepEqual(dp.Spec.Template.Spec.Containers[0].Image, getBestieContainerImage(bestie))

	if bestieImageDifferent {
		if bestieImageDifferent {
			log.Info("Upgrade Operand")
			dp.Spec.Template.Spec.Containers[0].Image = getBestieContainerImage(bestie)
		}Spec.Template.Spec.Containers[0].Image = getBestieContainerImage(bestie)
        }
 ```

---

#### How do you ensure seamless upgrade ?

---

#### Liveness and Readiness Probe

- Controls the health of an application running inside a Pod’s container.

<aside class="notes"> Failing liveness probe will restart the container, whereas failing readiness probe will stop our application from serving traffic.</aside>

---

#### Demo
