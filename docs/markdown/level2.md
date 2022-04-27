#### Seamless upgrade of the Operator and the Operand

- An upgrade of the operator automatically ensures the instantiated resources for each CR are in the new desired state and which would upgrade the operand.
- Operator can be upgraded seamlessly and can either still manage older versions of the operand or update them.

---

#### Workflow

- Implement version upgrade in the operator controller
- Deploy the operator
- Update the CR version and apply
- Test the version upgrade

---

#### Compare the current version of the CR instance with the Container Image

- Get the desired version from CR
- Upgrade the container image if current version is less than desired version

---


#### Liveness and Readiness Probe

- Controls the health of an application running inside a Pod’s container.
- Failing liveness probe will restart the container, whereas failing readiness probe will stop our application from serving traffic.