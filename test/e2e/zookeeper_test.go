// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"testing"
	"time"

	goctx "context"
	apis "github.com/ronin13/zookeeper-operator/pkg/apis"
	operator "github.com/ronin13/zookeeper-operator/pkg/apis/wnohang/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 360
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestZookeeper(t *testing.T) {
	memcachedList := &operator.ZookeeperList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Zookeeper",
			APIVersion: "wnohang.net/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, memcachedList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("zookeeper-group", func(t *testing.T) {
		t.Run("Cluster", ZookeeperCluster)
		t.Run("Cluster2", ZookeeperCluster)
	})
}

func zookeeperScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Could not get namespace: %v", err)
	}
	exampleZookeeper := &operator.Zookeeper{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Zookeeper",
			APIVersion: "wnohang.net/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-zookeeper",
			Namespace: namespace,
		},
		Spec: operator.ZookeeperSpec{
			Nodes:   3,
			Storage: "10Mi",
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleZookeeper, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for example-zookeeper to reach 3 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-zookeeper", 3, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-zookeeper", Namespace: namespace}, exampleZookeeper)
	if err != nil {
		return err
	}
	exampleZookeeper.Spec.Nodes = 4
	err = f.Client.Update(goctx.TODO(), exampleZookeeper)
	if err != nil {
		return err
	}

	// wait for example-zookeeper to reach 4 replicas
	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-zookeeper", 4, retryInterval, timeout)
}

func ZookeeperCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "zookeeper-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	//if err = zookeeperScaleTest(t, f, ctx); err != nil {
	//t.Fatal(err)
	//}

}
