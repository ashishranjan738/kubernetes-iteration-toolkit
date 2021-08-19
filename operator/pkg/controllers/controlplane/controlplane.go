/*
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

package controlplane

import (
	"context"
	"fmt"

	"github.com/awslabs/kit/operator/pkg/apis/infrastructure/v1alpha1"
	"github.com/awslabs/kit/operator/pkg/awsprovider"
	"github.com/awslabs/kit/operator/pkg/controllers"
	"github.com/awslabs/kit/operator/pkg/controllers/etcd"
	"github.com/awslabs/kit/operator/pkg/controllers/master"
	"github.com/awslabs/kit/operator/pkg/kubeprovider"
	"github.com/awslabs/kit/operator/pkg/results"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type controlPlane struct {
	ec2api           *awsprovider.EC2
	etcdController   *etcd.Controller
	masterController *master.Controller
}

// NewController returns a controller for managing VPCs in AWS
func NewController(ec2api *awsprovider.EC2, kubeClient client.Client) *controlPlane {
	return &controlPlane{ec2api: ec2api,
		etcdController:   etcd.New(kubeprovider.New(kubeClient)),
		masterController: master.New(kubeprovider.New(kubeClient)),
	}
}

// Name returns the name of the controller
func (c *controlPlane) Name() string {
	return "control-plane"
}

// For returns the resource this controller is for.
func (c *controlPlane) For() controllers.Object {
	return &v1alpha1.ControlPlane{}
}

type reconciler func(ctx context.Context, controlPlane *v1alpha1.ControlPlane) (err error)

// Reconcile will check if the resource exists is AWS if it does sync status,
// else create the resource and then sync status with the ControlPlane.Status
// object
func (c *controlPlane) Reconcile(ctx context.Context, object controllers.Object) (res *reconcile.Result, err error) {
	for _, reconcile := range []reconciler{
		c.etcdController.Reconcile,
		c.masterController.Reconcile,
	} {
		if err := reconcile(ctx, object.(*v1alpha1.ControlPlane)); err != nil {
			return nil, fmt.Errorf("reconciling, %w", err)
		}
	}
	return results.Created, nil
}

func (c *controlPlane) Finalize(_ context.Context, _ controllers.Object) (*reconcile.Result, error) {
	return results.Terminated, nil
}