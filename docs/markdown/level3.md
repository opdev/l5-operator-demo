#### More Parts of the lifecyle
- backups
- restores

<aside class="notes">
  Speaker notes:
  So far we have seen how we can have relatively seamless upgrades but there are more parts to the application life cycle, specifically backups and restores
</aside>

---
#### Postgres Operator
- A Database-as-a-Service but within our cluster

<aside class="notes">
  Speaker note:
  However, Since our app stores its state in a postgres database which has been provisioned by the postgres operator, we can continue to leverage that operators features to have backup and restore functionality. The postgres operator essentially allows us to have a "database-as-a-service" but one that is completely in our control.
</aside>

---
#### Backup Options
- Write Ahead Logs
- Scheduled and On Demand Backups Offsite
- Retention Policies

<aside class="notes">
  Speaker Notes:
  The postgres operator gives us many things out of the box including WAL which enable point in time recovery, offsite backups to S3 and other storage providers. It also allows us to configure differntial and full backups with specific schedules and retention policies.
</aside>

---
#### One Off backup

```
oc annotate -n postgres-operator postgrescluster bestie-pgc \
  postgres-operator.crunchydata.com/pgbackrest-backup="$(date)"
```

<aside class="notes">
Speaker notes:
Don't yet have a demo prepared but open to questions and feedback
</aside>

---
#### What about restores ?

---
#### Service disruptions
- Application and database compatibility

<aside class="notes">
  Speaker notes:
  But backups are only part of the picture what about restores ? Restores bring about a few more complications. If we restore a database backup the app might not work correctly as the database version and the app version are not compatible for all pods. 
</aside>

---
#### The "easy way"
- Allow for some service disruption

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
Don't yet have a demo prepared but open to questions and feedback
</aside>

---
#### Ensure Backward compaitibility
- Effectively always roll forward

<aside class="notes">
  Speaker notes:
  So one we need to ensure that the app and the db are compatible and two we need to minimize dataloss and distruption when switching between versions. One approach to handle this is to bake the appropriate migrations scripts into the app itself so that the app can be compatible with different database versions. However this may not always be possible. An even more simple approach is to just account for some service disruption and stop traffic to the application.
</aside>

---

#### A more generalized approach
- Clone the existing database
- Spin up a new instance with a different app version
- Switch traffic over

<aside class="notes">
  Speaker notes:
  Another way to handle this kind of a scenario i.e. restore a backup of a database version that is not compatible with both the current and the target app version is to follow this sort of general orchestration workflow. Switch the app into a read only mode in order to prevent dataloss during the upgrade process.. spin up a new database instance and a a new deployment with an older version of the app and the database (which are compatible) .. backup the read only isntance and restore it to the newly spun up instance after applying any migrations if neccessary and then switch traffic over. This is something that can be automated by software operators.
</aside>

---
#### Cloning the db (from bestie1)
- Set the db as the datasource
```
spec:
  dataSource:
    postgresCluster:
      clusterName: bestie1-pgc
      repoName: repo1
```

<aside class="notes">
Speaker notes:
we can create a new db with the old db as the source
</aside>

---
#### Cloning the application (create bestie2)
- Create a new bestie custom resource
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
we can create a new db with the old db as the source
</aside>

---
#### orchrestrated by the operator!

![General Approach](images/general_approach.png)

<aside class="notes">
Speaker notes:
There are other tools that can do this but the advantage is that you can customize, package and distribute this with your app through operator hub
</aside>

---
