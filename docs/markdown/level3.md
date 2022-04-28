#### More Parts of the lifecyle
- What about backups and restores ?

<aside class="notes">
  Speaker notes:
  So far we have seen how we can have relatively seamless upgrades but there are more parts to the application life cycle, specifically backups and restores
</aside>

---
#### Good news!
- Some things in life are free for everything else there is k8s

<aside class="notes">
  Speaker notes:
  As has been the theme so far we get a lot of things for free and this carries over to the backup and restore parts of the application lifecycle
</aside>

---
#### Handling Backups
- What things need to be backed up ?

<aside class="notes">
  Speaker Notes:
  Our app is stateless we need only backup the db. We could also have some operator state stored in our custom resource and thats something that needs to be backed up as well.
</aside>

---
#### Postgres Operator
- Gives us backups for free

<aside class="notes">
  Speaker note:
  Since our app stores its state in a postgres database which has been provisioned by the postgres operator, we can continue to leverage that operators features to have backup and reestore functionality
</aside>

---
#### Backup Options
- Write Ahead Logs
- Scheduled and On Demand Backups Offsite
- Retention Policies

<aside class="notes">
  Speaker Notes:
  The postgres operator gives us many things out of the box including WAL, offsite backups to S3 and other storage providers. It also allows us to configure differntial and full backups with specific schedules and retention policies.
</aside>

---
#### What about Restores
- Implement database backward compatibility
- Link app version and db version via the CR

<aside class="notes">
  Speaker notes:
  But backups are only part of the picture what about restores ? Restores bring about a few more complications. If we restore a database backup the app might not work correctly as the database version and the app version are not compatible for all pods. So one we need to ensure that the app and the db are compatible and two we need to minimize dataloss and distruption when switching between versions. One approach to handle this is to bake the appropriate migrations scripts into the app itself so that the app can be compatible with different database versions. However this may not always be possible.
</aside>

---
#### Advanced scenarios
- Read only mode
- Spin up a new instance
- Switch traffic over

<aside class="notes">
  Speaker notes:
  Another way to handle this kind of a scenario i.e. restore a backup of a database version that is not compatible with both the current and the target app version is to follow this sort of general orchestration workflow. Switch the app into a read only mode, spin up a new database instance and a a new deployment with an older version of the app and the database (which are compatible) and switch traffic over.
</aside>

---
#### Demo

```
kubectl annotate -n postgres-operator postgrescluster bestie-pgc \
  postgres-operator.crunchydata.com/pgbackrest-backup="$(date)"
```

<aside class="notes">
Speaker notes:
Don't yet have a demo prepared but open to questions and feedback
</aside>

---
#### Questions ?
