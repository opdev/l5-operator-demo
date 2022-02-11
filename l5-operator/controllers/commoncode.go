/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"os"

	petsv1 "github.com/opdev/l5-operator-demo/l5-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Returns whether or not the MySQL deployment is running
func (r *BestieReconciler) isRunning(ctx context.Context, bestie *petsv1.Bestie) bool {
	dp := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{Name: bestie.Name + "-app", Namespace: bestie.Namespace}, dp)

	if err != nil {
		log.Error(err, "Deployment found")
		return false
	}
	if dp.Status.ReadyReplicas == 1 {
		return true
	}

	return false
}

func (r *BestieReconciler) applyManifests(ctx context.Context, req ctrl.Request, bestie *petsv1.Bestie, obj client.Object, fileName string) error {

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

	obj.SetNamespace(bestie.GetNamespace())
	//obj.SetName(bestie.GetName())
	controllerutil.SetControllerReference(bestie, obj, r.Scheme)

	err = r.Client.Create(ctx, obj)
	if err != nil {
		Log.Error(err, "Failed to create object", "object", obj.GetName())
		return err
	}

	return nil
}
