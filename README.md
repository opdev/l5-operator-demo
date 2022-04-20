# L5 Operator demo

The goal is to develop a demo operator with level 5 capabilities to serve as an example to enhance workshop [content](https://drive.google.com/drive/u/0/folders/1l6FY1QdBq1IsmwM6Ib44A8h12OSKGJbe) as well as to present at kubecon for which a [proposal](https://drive.google.com/file/d/1GjJgBcJmywP3L64m1h4vZ68UIu-XJxMZ/view?usp=sharing) was submitted. The capabilities are being developed according to our interpretation of the requirements described by the operator capability descriptions given in the operator sdk [here](https://docs.google.com/document/d/1gNa2NQzlsHDdNHBYPczCytkuokEzBCFKjlxM12X5cdk/edit?usp=sharing)

# Requirements

- The L5 Operator requires ingress controller to be installed if running in k8s cluster.
- Steps for installing ingress controller for different clusters can be followed from [here](https://kubernetes.github.io/ingress-nginx/deploy/)

# Installation

#### Make sure to install Crunchy Data Postgres Operator before installing the L5 operator. 

- Crunchy Data Postgres Operator can be directly installed from the [Openshift Container Platform](https://oauth-openshift.apps.eng.opdev.io/oauth/authorize?client_id=console&redirect_uri=https%3A%2F%2Fconsole-openshift-console.apps.eng.opdev.io%2Fauth%2Fcallback&response_type=code&scope=user%3Afull&state=d933ad0d), when running the application on an Openshift Cluster.
- While running on a vanilla kubernetes, follow the steps from [here](https://access.crunchydata.com/documentation/postgres-operator/v5/quickstart/) to install the Crunchy Data Postgres Operator.