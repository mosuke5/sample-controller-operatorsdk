package foo

import (
	"context"
	"testing"

	foov1alpha1 "github.com/mosuke5/sample-controller-operatorsdk/pkg/apis/samplecontroller/v1alpha1"

    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    "k8s.io/client-go/kubernetes/scheme"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"	
)

func TestFooController(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var (
        name            = "foo-test"
        namespace       = "foo-namespace"
        deploymentname  = "foo-deploy"
        replicas = int32(3)
    )

    // A foo object with metadata and spec.
    foo := &foov1alpha1.Foo{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
        },
        Spec: foov1alpha1.FooSpec{
            DeploymentName: deploymentname,
            Replicas: &replicas,
        },
	}
	
	// Objects to track in the fake client.
    objs := []runtime.Object{ foo }

    // Register operator types with the runtime scheme.
    s := scheme.Scheme
    s.AddKnownTypes(foov1alpha1.SchemeGroupVersion, foo)

    // Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	
	// Create a ReconcileFoo object with the scheme and fake client.
    r := &ReconcileFoo{client: cl, scheme: s}

    // Mock request to simulate Reconcile() being called on an event for a
    // watched resource .
    req := reconcile.Request{
        NamespacedName: types.NamespacedName{
            Name:      name,
            Namespace: namespace,
        },
    }
	_, err := r.Reconcile(req)

	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check if foo has been created.
    fooreq := reconcile.Request{
        NamespacedName: types.NamespacedName{
            Name:      name,
            Namespace: namespace,
        },
    }
	err = cl.Get(context.TODO(), fooreq.NamespacedName, foo)
	if err != nil {
		t.Fatalf("get foo: (%v)", err)
	}

	// Check if deployment has been created and has the correct size.
    depreq := reconcile.Request{
        NamespacedName: types.NamespacedName{
            Name:      deploymentname,
            Namespace: namespace,
        },
    }
	dep := &appsv1.Deployment{}
	err = cl.Get(context.TODO(), depreq.NamespacedName, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	// Check if the quantity of Replicas for this deployment is equals the specification
	dsize := *dep.Spec.Replicas
	if dsize != replicas {
		t.Errorf("dep size (%d) is not the expected size (%d)", dsize, replicas)
	}
}

func TestCurstom(t *testing.T) {
	t.Errorf("aiuei")
}