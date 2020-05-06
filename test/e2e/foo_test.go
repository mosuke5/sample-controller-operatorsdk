package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	"github.com/mosuke5/sample-controller-operatorsdk/pkg/apis"
	operator "github.com/mosuke5/sample-controller-operatorsdk/pkg/apis/samplecontroller/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestFoo(t *testing.T) {
	fooList := &operator.FooList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, fooList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("foo-group", func(t *testing.T) {
		t.Run("Cluster", FooCluster)
		//t.Run("Cluster2", MemcachedCluster)
	})
}

func fooScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	replicas := int32(3)
	after_replicas := int32(4)

	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	// create foo custom resource
	exampleFoo := &operator.Foo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-foo",
			Namespace: namespace,
		},
		Spec: operator.FooSpec{
			DeploymentName: "test-deployment",
			Replicas: &replicas,
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleFoo, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for test-deployment to reach 3 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "test-deployment", 3, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-foo", Namespace: namespace}, exampleFoo)
	if err != nil {
		return err
	}
	exampleFoo.Spec.Replicas = &after_replicas
	err = f.Client.Update(goctx.TODO(), exampleFoo)
	if err != nil {
		return err
	}

	// wait for example-foo to reach 4 replicas
	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "test-deployment", 4, retryInterval, timeout)
}

func FooDeleteTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	replicas := int32(1)

	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	// create foo custom resource
	exampleFoo := &operator.Foo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "delete-test-foo",
			Namespace: namespace,
		},
		Spec: operator.FooSpec{
			DeploymentName: "delete-test-deployment",
			Replicas: &replicas,
		},
	}

	err = f.Client.Create(goctx.TODO(), exampleFoo, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "delete-test-deployment", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = f.Client.Delete(goctx.TODO(), exampleFoo)
	if err != nil {
		return err
	}

	// wait for example-foo to reach 0 replicas
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "delete-test-deployment",
			Namespace: namespace,
		},
	}
	return e2eutil.WaitForDeletion(t, f.Client.Client, deployment, retryInterval, timeout)
}

func FooCluster(t *testing.T) {
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
	// wait for foo-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "foo-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = fooScaleTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}

	if err = FooDeleteTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}