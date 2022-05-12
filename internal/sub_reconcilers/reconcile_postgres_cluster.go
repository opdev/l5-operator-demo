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

	pgov1 "github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
	"github.com/opdev/l5-operator-demo/internal/util"
)

type PostgresClusterCRReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewPostgresClusterCRReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *PostgresClusterCRReconciler {
	return &PostgresClusterCRReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *PostgresClusterCRReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "PostgresReconciler",
	}
	log := r.Log.WithValues("postgres", logInfo)
	log.Info("reconcile postgres if it does not exist")
	pgo := &pgov1.PostgresCluster{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-pgo", Namespace: bestie.Namespace}, pgo)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Postgres Cluster for Bestie")
		fileName := "config/resources/postgrescluster.yaml"
		err = r.applyManifests(ctx, bestie, pgo, fileName)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error during Manifests apply - %w", err)
		}
		err = util.RefreshCustomResource(ctx, r.client, bestie)
		if err != nil {
			log.Error(err, "Unable to refresh custom resource")
			return ctrl.Result{}, err
		}
		meta.SetStatusCondition(&bestie.Status.Conditions, NewDatabaseCreatedCondition())
		err = r.client.Status().Update(ctx, bestie)
		if err != nil {
			log.Error(err, "Unable to refresh custom resource")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *PostgresClusterCRReconciler) applyManifests(ctx context.Context, bestie *petsv1.Bestie, obj client.Object, fileName string) error {
	err := util.ApplyManifests(r.client, r.Scheme, ctx, bestie, obj, fileName)
	if err != nil {
		return err
	}
	return err
}
