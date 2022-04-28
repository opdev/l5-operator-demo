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

package reconcilers

import (
	"context"
	"fmt"
	"os"

	pgov1 "github.com/crunchydata/postgres-operator/pkg/apis/postgres-operator.crunchydata.com/v1beta1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
)

const (
	pgoreconcilerName = "DatabaseSeedJobReconciler"
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

func (r *PostgresClusterCRReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (bool, error) {
	// reconcile Postgres
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      pgoreconcilerName,
	}
	log := r.Log.WithValues("postgres", logInfo)
	log.Info("reconcile postgres if it does not exist")
	pgo := &pgov1.PostgresCluster{}
	err := r.client.Get(ctx, types.NamespacedName{Name: bestie.Name + "-pgo", Namespace: bestie.Namespace}, pgo)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Postgres Cluster for Bestie")
		fileName := "config/resources/postgrescluster.yaml"
		err := r.ApplyManifests(ctx, bestie, pgo, fileName)
		if err != nil {
			return false, fmt.Errorf("Error during Manifests apply - %w", err)
		}
	}
	return true, err
}

func (r *PostgresClusterCRReconciler) ApplyManifests(ctx context.Context, bestie *petsv1.Bestie, obj client.Object, fileName string) error {

	Log := ctrllog.FromContext(ctx)

	b, err := os.ReadFile(fileName)
	if err != nil {
		Log.Error(err, fmt.Sprintf("Couldn't read manifest file for: %s", fileName))
		return err
	}

	if err = yamlutil.Unmarshal(b, &obj); err != nil {
		Log.Error(err, fmt.Sprintf("Couldn't unmarshall yaml file for: %s", fileName))
		return err
	}

	obj.SetNamespace(bestie.Namespace)

	err = controllerutil.SetControllerReference(bestie, obj, r.Scheme)
	if err != nil {
		return err
	}

	err = r.client.Create(ctx, obj)
	if err != nil {
		Log.Error(err, "Failed to create object", "object", obj.GetName())
		return err
	}

	return nil
}
