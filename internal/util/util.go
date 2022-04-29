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

package util

import (
	"context"
	"fmt"
	"os"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	petsv1 "github.com/opdev/l5-operator-demo/api/v1"
)

func ApplyManifests(client client.Client, scheme *runtime.Scheme, ctx context.Context, bestie *petsv1.Bestie, obj client.Object, fileName string) error {
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

	err = controllerutil.SetControllerReference(bestie, obj, scheme)
	if err != nil {
		return err
	}

	err = client.Create(ctx, obj)
	if err != nil {
		Log.Error(err, "Failed to create object", "object", obj.GetName())
		return err
	}

	return nil
}

func IsRouteAPIAvailable() bool {
	// TODO add logging
	found, err := verifyOpenShiftCluster(routev1.GroupName, routev1.GroupVersion.Version)
	if err != nil || !found {
		return false
	}
	return true
}

func verifyOpenShiftCluster(group string, version string) (bool, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return false, err
	}

	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}

	gv := schema.GroupVersion{
		Group:   group,
		Version: version,
	}
	if err = discovery.ServerSupportsVersion(k8s, gv); err != nil {
		return false, nil
	}
	return true, nil
}
