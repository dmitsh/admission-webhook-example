package main

import (
	"context"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func createMutationConfig(ctx context.Context, caCert []byte) error {
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	path := "/mutate"
	fail := admissionregistrationv1.Fail
	none := admissionregistrationv1.SideEffectClassNone

	mutateconfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: mutationCfgName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name: webhookService + "." + webhookNamespace + ".svc",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caCert,
				Service: &admissionregistrationv1.ServiceReference{
					Name:      webhookService,
					Namespace: webhookNamespace,
					Path:      &path,
				},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{{Operations: []admissionregistrationv1.OperationType{
				admissionregistrationv1.Create},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{"apps"},
					APIVersions: []string{"v1"},
					Resources:   []string{"deployments", "pods"},
				},
			}},
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "kubernetes.io/cluster-service",
						Operator: metav1.LabelSelectorOpNotIn,
						Values:   []string{"true"},
					},
				},
			},
			FailurePolicy:           &fail,
			SideEffects:             &none,
			AdmissionReviewVersions: []string{"v1beta1", "v1"},
		}},
	}

	if _, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, mutateconfig, v1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func deleteMutationConfig(ctx context.Context) error {
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	return kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, mutationCfgName, v1.DeleteOptions{})
}
