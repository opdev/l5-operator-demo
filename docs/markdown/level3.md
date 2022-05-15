#### More Parts of the lifecyle
- backups
- restores

<aside class="notes">
  Speaker notes:
  So far we have seen how we can have relatively seamless upgrades but there are more parts to the application life cycle, specifically backups and restores
</aside>

---
#### 2nd Free Lunch
The postgres operator gives us level 3 for free

<aside class="notes">
  Speaker note:
  However, Since our app stores its state in a postgres database which has been provisioned by the postgres operator, we can continue to leverage that operators features to have backup and restore functionality. 
</aside>

---
#### Postgres Operator
A Database-as-a-Service but within our cluster

<aside class="notes">
  Speaker note:
  Our app stores its state in a postgres database, which has been provisioned by the postgres operator. The postgres operator essentially allows us to have a "database-as-a-service" but one that is completely in our control.
</aside>

---
#### How to consume the postgres operator
- OLM installs it as a dependency
- Our operator manipulates the PostgresCluster custom resource

<aside class="notes">
  Speaker note:
  When we say we consume the postgres operator we mean two things
</aside>

---
#### The PostgresCluster custom resource
```
apiVersion: postgres-operator.crunchydata.com/v1beta1
kind: PostgresCluster
metadata:
  name: bestie-pgo
  spec:
    image: registry.developers.crunchydata.com/crunchydata/crunchy-postgres:ubi8-14.2-1
    postgresVersion: 14
    instances:
      - name: instance1
        dataVolumeClaimSpec:
          accessModes:
          - "ReadWriteOnce"
          resources:
            requests:
              storage: 1Gi
  
```

<aside class="notes">
  Speaker note:
  Earlier we noted that the custom resource can be thought of as a user or consumer interface to operators or kubernetes native applications. An operator can consume another operator by creating / interacting with another operators custom resource which is intern reconciled by that operator.
</aside>

---
#### Backup Options
- Write Ahead Logs
- Scheduled and On Demand Backups Offsite
- Retention Policies

<aside class="notes">
  Speaker Notes:
  The postgres operator is an example of one of the more mature operators in the ecosystem (The top %5 as we alluded to earlier). The postgres operator gives us many things out of the box including WAL which enable point in time recovery, offsite backups to S3 and other storage providers. It also allows us to configure differntial and full backups with specific schedules and retention policies.
</aside>

---
#### Simple Example
One-off backup
```
oc annotate -n postgres-operator postgrescluster 
bestie-pgc \
postgres-operator.crunchydata.com/pgbackrest-backup="$(date)"
```

<aside class="notes">
Speaker notes:
The operator can achieve this step by retrieving the latest version of the PostgresCluster and editing its annotations. The Postgres Operator will be notified of this change via an event since it is watching for changes and will go ahead and run the appropriate actions to complete the backup and once done will remove the annotation.
</aside>

---
#### What about restores ?

<aside class="notes">
  Speaker notes:
  But backups are only part of the picture.. Restores bring about a few more complications..
</aside>

---
#### Service disruptions
Coming back to the problem of application and database compatibility

<aside class="notes">
  Speaker notes:
  If we restore a database backup the app might not work correctly as the database version and the app version are not compatible for all pods. 
</aside>

---
#### Ensure Backward compatibility
Effectively always roll forward

![Incompatible Changes](images/incompatible_upgrade.png)

<aside class="notes">
  Speaker notes:
  Another way to avoid disruption is to make small changes and always ensure that the app and the db are compatible. One approach to handle this is to bake the appropriate migrations scripts into the app itself so that the app can be compatible with different database versions. However this may not always be possible.
</aside>

---
#### The "easy way"
Allow for some service disruption

---
#### In place point in time recovery
```
spec:
  backups:
    pgbackrest:
      restore:
        enabled: true
        repoName: repo1
        options:
        - --type=time
        - --target="2021-06-09 14:15:11-04"
```

```
kubectl annotate -n postgres-operator postgrescluster bestie-pgc --overwrite \
  postgres-operator.crunchydata.com/pgbackrest-restore=id1
```
  
<aside class="notes">
  Speaker notes:
  One way to approach this is to take advantage of the write ahead logs and perform an inplace point in time recovery. This can be achieved in a similar way to the simple backup example we saw earlier.
</aside>

---
#### A more generalized approach
- Clone the existing database
- Spin up a new instance with a different app version
- Perform disruptve operations
- Switch traffic over

<aside class="notes">
  Speaker notes:
  A more general way to handle this kind of a scenario i.e. restore a backup of a database version that is not compatible with both the current and the target app version is to follow this sort of general orchestration workflow. 
  - switch the app into a read only mode in order to prevent dataloss during the upgrade process..
  - spin up a new database instance by cloning the existing db
  - perform disruptive actions ex: rolling back or pushing an update with incompatible database changes..
  - switch traffic over. This is something that can be automated by software operators.
</aside>

---
#### Cloning the db (from Bestie version A)
Set the old db as the datasource

```
apiVersion: postgres-operator.crunchydata.com/v1beta1
kind: PostgresCluster
metadata:
  name: bestie2-pgo
spec:
  dataSource:
    postgresCluster:
      clusterName: bestie1-pgc
      repoName: repo1
```

<aside class="notes">
Speaker notes:
This can be done by modifying by creating a new PostgresCluster custom resource and referring to the existing database as the datasource in the spec
</aside>

---
#### Create Bestie version B
Point to the cloned database (Version A)
```
apiVersion: pets.bestie.com/v1
kind: Bestie
metadata:
  name: bestie2
spec:
  size: 3
  image: quay.io/opdev/bestie
  maxReplicas: 10
  version: "1.4"
```

<aside class="notes">
Speaker notes:
The operator refers to the database via the postgrescluster custom resource with a pre-defined naming convention so it will automatically use the cloned database. Lastly we update the existing ingress to point to this new stack. At this point in compatible migrations can be done without any user impact.
</aside>


---
#### All this can be orchrestrated by the operator!

![General Approach](images/general_approach.png)

<aside class="notes">
Speaker notes:
There are other tools that can do achieve this sort of blue/green red/black deployment strategy but the advantage of using an operator is that you can customize, package and distribute this with your application and provide your users with a one click app store like experience via the operator hub. So all the building blocks to acheieve this workflow are in place and can be orchrestated by our operator. I don't yet have a working demo for this part but you should be able to try in the upcoming weeks using the demo l5 operator that has been published to the community operator hub.
</aside>
