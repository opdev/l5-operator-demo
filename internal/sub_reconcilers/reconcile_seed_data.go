/*
Copyright The L5 Operator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sub_reconcilers

import (
	"context"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/opdev/l5-operator-demo/internal/util"

	pgov1 "github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
)

type DatabaseSeedJobReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewDatabaseSeedJobReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *DatabaseSeedJobReconciler {
	return &DatabaseSeedJobReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *DatabaseSeedJobReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "DatabaseSeedJobReconciler",
	}
	log := r.Log.WithValues("database_seed_job", logInfo)

	// wait for postgres to come up
	pgo := &pgov1.PostgresCluster{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-pgo", Namespace: bestie.Namespace}, pgo)
	if err != nil && errors.IsNotFound(err) {
		log.Info("No postgres instance found, has the postgres cr been created ?")
		delay := time.Second * time.Duration(5)
		log.Info(fmt.Sprintf("will retry after waiting for %s", delay))
		return ctrl.Result{RequeueAfter: delay}, nil
	}
	postgresReady := false
	if pgo.Status.InstanceSets != nil && len(pgo.Status.InstanceSets) > 0 && pgo.Status.InstanceSets[0].ReadyReplicas >= 1 {
		postgresReady = true
	}
	if !postgresReady {
		// If postgres is not ready yet, requeue after delay seconds.
		delay := time.Second * time.Duration(15)
		log.Info(fmt.Sprintf("postgres is instantiating, waiting for %s", delay))
		// implement requeue after delay in subreconciler
		return ctrl.Result{RequeueAfter: delay}, nil
	}
	// seed the database - as long as the postgres app is up and running this can run.
	log.Info("create a job to seed the database")
	job := &batchv1.Job{}
	err = r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-job", Namespace: bestie.Namespace}, job)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new job for bestie")
		fileName := "config/resources/bestie-job.yaml"
		err := r.applyManifests(ctx, bestie, job, fileName)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Error during Manifests apply - %w", err)
		}
	}
	return ctrl.Result{}, nil
}

func (r *DatabaseSeedJobReconciler) applyManifests(ctx context.Context, bestie *petsv1.Bestie, obj client.Object, fileName string) error {
	err := util.ApplyManifests(r.client, r.Scheme, ctx, bestie, obj, fileName)
	if err != nil {
		return err
	}
	return err
}
