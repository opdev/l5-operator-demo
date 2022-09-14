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

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespaceLabel = "openshift.io/cluster-monitoring"
)

type LabelReconciler struct {
	client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewLabelReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *LabelReconciler {
	return &LabelReconciler{
		client: client,
		Log:    log,
		Scheme: scheme,
	}
}

func (r *LabelReconciler) Reconcile(ctx context.Context, bestie *petsv1.Bestie) (ctrl.Result, error) {
	logInfo := types.NamespacedName{
		Namespace: bestie.Namespace,
		Name:      "LabelReconciler",
	}
	r.Log.WithValues("Label Reconcile", logInfo)

	ns := &corev1.Namespace{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: bestie.Namespace}, ns)
	if err != nil {
		return ctrl.Result{}, err
	}
	if ns.ObjectMeta.Labels[namespaceLabel] != "true" {
		ns.ObjectMeta.Labels[namespaceLabel] = "true"
		err := r.client.Update(ctx, ns.DeepCopy())
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
