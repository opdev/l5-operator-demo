# L5 Demo Operator 🏗️

The projects goal is to develop a demo operator with level 5 capabilities to serve as an example to enhance workshop [(internal)](https://drive.google.com/drive/u/0/folders/1l6FY1QdBq1IsmwM6Ib44A8h12OSKGJbe) as well as to present at kubecon or a similar conference for which a [proposal](https://drive.google.com/file/d/1GjJgBcJmywP3L64m1h4vZ68UIu-XJxMZ/view?usp=sharing) was submitted. One version of the [(presentation)](https://opdev.github.io/l5-operator-demo/) was accepted to [(cloud native rejects)](https://cfp.cloud-native.rejekts.io/cloud-native-rejekts-eu-valencia-2022/talk/review/LESBCQ3MJQFEE3JWMRCJKKL9XT7YMDHP). The capabilities are being developed according to our interpretation of the requirements described by the operator capability descriptions given in the operator sdk which are currently being evolved [working in progress](https://docs.google.com/document/d/1gNa2NQzlsHDdNHBYPczCytkuokEzBCFKjlxM12X5cdk/edit?usp=sharing)

## Current Capability descriptions
The l5 operator is an example of a minimal implementation of the 5 capability levels described by the operator framework

### Level 1: Basic Install
The L5 operator is able to take advantage of the Operator Lifecycle Manager and is able to deployed with one click. Once deployed it is able to set up its operator upon creation of a custom resource. User can install the operand by creating and configuring the CR. When configuration changes are applied, the app version is reported in the status and the operand updates in a non-distruptive fashion when the configuration of the CR is changed.

### Level 2: Seamless Upgrades
It's possible to rollout both operator and operand updates seamlessly with some caveats. Liveness and readiness probes come free. The operator is able to update the operand via the CR field. By doing so, we are able to control the app version and report the status.

### Level 3: Full Lifecycle
Backup and restore functionality is provided via the crunchy postgres operator which we consume in order to get a database-as-a-service right within our cluster. The L5 operator contains a liveness and readiness probe, so if the connection to the database fails, it will wait for reconfiguration work to be finished. The L5 operator utilizes the rolling deployment strategy.

### Level 4: Deep Insights
The operator as well as operand expose metrics. These are aggregated using Prometheus and visualized using Grafana. The operator exposes health metrics endpoints such as app latency, request per second, and http codes. Currently we are working on implementing the RED method which defines three key metrics for any service: _*rate*_ - the number of requests per second, _*errors*_ - the number of those requests that are failing, and _*duration*_ - the amount of time those requests take.

### Level 5: Autopilot
The operator is able to autoscale by automatically provisioning a horizontal pod autoscaler that automatically changes the size of the deployment based on application load. The operator is able to scale the pods accordingly within the range of the size and maxReplicas we declare.

# Operator Installation
The operator is published as a community operator on the openshift operator hub.

## Different ways to run the Operator using the Operator SDK

### Prerequisites
- When not using OLM to install the L5 Operator, the Crunchy Data Postgres Operator needs to be installed either from Operator Hub when using Openshift or [manually](https://access.crunchydata.com/documentation/postgres-operator/v5/quickstart/)
- The L5 Operator requires an ingress controller to be installed if running on vanilla kubernetes if not already present. Steps for installing ingress controller for different clusters can be followed from [here](https://kubernetes.github.io/ingress-nginx/deploy/)

### 1. Run locally outside the cluster
```
git clone https://github.com/opdev/l5-operator-demo
```

```
cd <project>
```

```
make generate
```

```
make manifests
```

```
make install
```

```
make run
```

```
oc apply -f config/samples/pets_v1_bestie.yaml
```

### 2. Run as a Deployment inside the cluster
```
git clone https://github.com/opdev/l5-operator-demo
```

```
cd <project>
```

```
make generate
```

```
make manifests
```

- Build and push your image to the location specified by IMG:
```
make docker-build docker-push IMG=<some-registry>/<project-name>:tag
```

- Deploy the controller to the cluster with image specified by IMG:
```
make deploy IMG=<some-registry>/<project-name>:tag
```

### 3. Deploy the Operator with OLM
```
operator-sdk run bundle <operator-bundle-image>
```

# High Level Diagrams

An editable version of this diagram is on google drive [(internal)](https://drive.google.com/file/d/1zwZDZyp_OqdqhPicXgfqIDRPZB4IYjwO/view?usp=sharing)

![Deployment Diagram](docs/images/hld.svg)

## "Traditional" Architecture

![Traditional Deployment](docs/images/bestie.drawio.svg)

## What it looks like in Kubernetes

![Deployment Diagram](docs/images/bestie_k8s.drawio.svg)
