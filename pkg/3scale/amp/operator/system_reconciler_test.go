package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSystemReconcilerCreate(t *testing.T) {
	var (
		log = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()

	apimanager := basicApimanagerSpecTestSystemOptions()
	// Objects to track in the fake client.
	objs := []runtime.Object{apimanager}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = imagev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = routev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := grafanav1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := configv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	reconciler := NewSystemReconciler(baseAPIManagerLogicReconciler)
	_, err = reconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      k8sclient.Object
	}{
		{"systemPVC", "system-storage", &v1.PersistentVolumeClaim{}},
		{"systemProviderService", "system-provider", &v1.Service{}},
		{"systemMasterService", "system-master", &v1.Service{}},
		{"systemDeveloperService", "system-developer", &v1.Service{}},
		{"systemMemcacheService", "system-memcache", &v1.Service{}},
		{"systemAppDC", "system-app", &appsv1.DeploymentConfig{}},
		{"systemSideKiqDC", "system-sidekiq", &appsv1.DeploymentConfig{}},
		{"systemCM", "system", &v1.ConfigMap{}},
		{"systemEnvironmentCM", "system-environment", &v1.ConfigMap{}},
		{"systemSMTPSecret", "system-smtp", &v1.Secret{}},
		{"systemEventsHookSecret", component.SystemSecretSystemEventsHookSecretName, &v1.Secret{}},
		{"systemMasterApicastSecret", component.SystemSecretSystemMasterApicastSecretName, &v1.Secret{}},
		{"systemSeedSecret", component.SystemSecretSystemSeedSecretName, &v1.Secret{}},
		{"systemRecaptchaSecret", component.SystemSecretSystemRecaptchaSecretName, &v1.Secret{}},
		{"systemAppSecret", component.SystemSecretSystemAppSecretName, &v1.Secret{}},
		{"systemMemcachedSecret", component.SystemSecretSystemMemcachedSecretName, &v1.Secret{}},
		{"systemMemcachedSecret", component.SystemSecretSystemMemcachedSecretName, &v1.Secret{}},
		{"systemAppPDB", "system-app", &policyv1.PodDisruptionBudget{}},
		{"systemSidekiqPDB", "system-sidekiq", &policyv1.PodDisruptionBudget{}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			obj := tc.obj
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, obj)
			// object must exist, that is all required to be tested
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}
		})
	}
}

func TestReplicaSystemReconciler(t *testing.T) {
	var (
		namespace        = "operator-unittest"
		log              = logf.Log.WithName("operator_test")
		oneValue   int32 = 1
		oneValue64 int64 = 1
		twoValue   int32 = 2
	)
	ctx := context.TODO()
	s := scheme.Scheme

	err := appsv1alpha1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}
	err = appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	if err := configv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName                 string
		objName                  string
		apimanager               *appsv1alpha1.APIManager
		expectedAmountOfReplicas int32
	}{
		{"system app replicas set", "system-app", testSystemAPIManagerCreator(&oneValue64, nil), oneValue},
		{"system app replicas not set", "system-app", testSystemAPIManagerCreator(nil, nil), twoValue},

		{"system sidekiq replicas set", "system-sidekiq", testSystemAPIManagerCreator(nil, &oneValue64), oneValue},
		{"system sidekiq replicas not set", "system-sidekiq", testSystemAPIManagerCreator(nil, nil), twoValue},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{tc.apimanager}
			// Create a fake client to mock API calls.
			cl := fake.NewFakeClient(objs...)
			clientAPIReader := fake.NewFakeClient(objs...)
			clientset := fakeclientset.NewSimpleClientset()
			recorder := record.NewFakeRecorder(10000)
			baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
			baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, tc.apimanager)

			reconciler := NewSystemReconciler(baseAPIManagerLogicReconciler)
			_, err = reconciler.Reconcile()
			if err != nil {
				t.Fatal(err)
			}

			dc := &appsv1.DeploymentConfig{}
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}

			err = cl.Get(context.TODO(), namespacedName, dc)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			// bump the amount of replicas in the dc
			dc.Spec.Replicas = twoValue
			err = cl.Update(context.TODO(), dc)
			if err != nil {
				subT.Errorf("error updating dc of %s: %v", tc.objName, err)
			}

			// re-run the reconciler
			_, err = reconciler.Reconcile()
			if err != nil {
				t.Fatal(err)
			}

			err = cl.Get(context.TODO(), namespacedName, dc)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			if tc.expectedAmountOfReplicas != dc.Spec.Replicas {
				subT.Errorf("expected replicas do not match. expected: %d actual: %d", tc.expectedAmountOfReplicas, dc.Spec.Replicas)
			}
		})
	}
}

func testSystemAPIManagerCreator(appReplicas, sidekiqReplicas *int64) *appsv1alpha1.APIManager {
	var (
		name                  = "example-apimanager"
		namespace             = "operator-unittest"
		wildcardDomain        = "test.3scale.net"
		appLabel              = "someLabel"
		tenantName            = "someTenant"
		trueValue             = true
		tmpApicastRegistryURL = apicastRegistryURL
	)

	return &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			Apicast: &appsv1alpha1.ApicastSpec{RegistryURL: &tmpApicastRegistryURL},
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
				WildcardDomain:               wildcardDomain,
				TenantName:                   &tenantName,
				ResourceRequirementsEnabled:  &trueValue,
			},
			System: &appsv1alpha1.SystemSpec{
				AppSpec:         &appsv1alpha1.SystemAppSpec{Replicas: appReplicas},
				SidekiqSpec:     &appsv1alpha1.SystemSidekiqSpec{Replicas: sidekiqReplicas},
				FileStorageSpec: &appsv1alpha1.SystemFileStorageSpec{},
				SearchdSpec:     &appsv1alpha1.SystemSearchdSpec{},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
}
