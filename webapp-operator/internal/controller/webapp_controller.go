/*
Copyright 2026.

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

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/sysman-x/operator-lab/api/v1alpha1"
)

// WebAppReconciler reconciles a WebApp object.
type WebAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// WebApp CRD 권한
// +kubebuilder:rbac:groups=apps.sysproto.com,resources=webapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.sysproto.com,resources=webapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.sysproto.com,resources=webapps/finalizers,verbs=update

// Deployment 권한
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Service 권한
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the Kubernetes reconciliation loop.
func (r *WebAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// -------------------------------------------------
	// 1. WebApp CR 조회
	// -------------------------------------------------

	var webapp appsv1alpha1.WebApp

	if err := r.Get(ctx, req.NamespacedName, &webapp); err != nil {
		if apierrors.IsNotFound(err) {
			// WebApp이 삭제된 경우
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	log.Info(
		"Reconciling WebApp",
		"name", webapp.Name,
		"namespace", webapp.Namespace,
	)

	// -------------------------------------------------
	// 2. Deployment 생성/확인
	// -------------------------------------------------

	deployment := &appsv1.Deployment{}

	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      webapp.Name,
			Namespace: webapp.Namespace,
		},
		deployment,
	)

	if apierrors.IsNotFound(err) {

		// Deployment가 없으면 새로 생성
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      webapp.Name,
				Namespace: webapp.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &webapp.Spec.Replicas,

				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": webapp.Name,
					},
				},

				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": webapp.Name,
						},
					},

					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "webapp",
								Image: webapp.Spec.Image,

								Ports: []corev1.ContainerPort{
									{
										ContainerPort: webapp.Spec.Port,
									},
								},
							},
						},
					},
				},
			},
		}

		// WebApp을 Deployment의 Owner로 설정
		if err := controllerutil.SetControllerReference(
			&webapp,
			deployment,
			r.Scheme,
		); err != nil {
			return ctrl.Result{}, err
		}

		// Deployment 생성
		if err := r.Create(ctx, deployment); err != nil {
			return ctrl.Result{}, err
		}

		log.Info(
			"Created Deployment",
			"name", deployment.Name,
		)

	} else if err != nil {
		return ctrl.Result{}, err
	}

	// -------------------------------------------------
	// 3. Service 생성/확인
	// -------------------------------------------------

	service := &corev1.Service{}

	err = r.Get(
		ctx,
		types.NamespacedName{
			Name:      webapp.Name,
			Namespace: webapp.Namespace,
		},
		service,
	)

	if apierrors.IsNotFound(err) {

		// Service가 없으면 새로 생성
		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      webapp.Name,
				Namespace: webapp.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": webapp.Name,
				},

				Ports: []corev1.ServicePort{
					{
						Port: webapp.Spec.Port,

						TargetPort: intstr.FromInt(
							int(webapp.Spec.Port),
						),
					},
				},
			},
		}

		// WebApp을 Service의 Owner로 설정
		if err := controllerutil.SetControllerReference(
			&webapp,
			service,
			r.Scheme,
		); err != nil {
			return ctrl.Result{}, err
		}

		// Service 생성
		if err := r.Create(ctx, service); err != nil {
			return ctrl.Result{}, err
		}

		log.Info(
			"Created Service",
			"name", service.Name,
		)

	} else if err != nil {
		return ctrl.Result{}, err
	}

	// -------------------------------------------------
	// 4. Reconcile 완료
	// -------------------------------------------------

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.WebApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("webapp").
		Complete(r)
}
