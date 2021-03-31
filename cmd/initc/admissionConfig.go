package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	scope := admissionregistrationv1.AllScopes
	timeout := int32(5)

	mutateconfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: mutationCfgName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name: mutationCfgName + "." + webhookNamespace + ".svc",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caCert,
				Service: &admissionregistrationv1.ServiceReference{
					Name:      webhookService,
					Namespace: webhookNamespace,
					Path:      &path,
				},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1.OperationType{
						admissionregistrationv1.Create,
						admissionregistrationv1.Update,
					},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
						Scope:       &scope,
					},
				},
			},
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
			TimeoutSeconds:          &timeout,
			AdmissionReviewVersions: []string{"v1beta1", "v1"},
		}},
	}

	cfg, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, mutationCfgName, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		log.Infof("createMutationConfig: creating config '%s'", mutationCfgName)
		_, err = kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, mutateconfig, v1.CreateOptions{})
	} else {
		for i := range cfg.Webhooks {
			cfg.Webhooks[i].ClientConfig.CABundle = caCert
		}
		log.Infof("createMutationConfig: updating config '%s'", mutationCfgName)
		_, err = kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, cfg, v1.UpdateOptions{})
	}
	return err
}

func deleteMutationConfig(ctx context.Context) error {
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	_, err = kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, mutationCfgName, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("deleteMutationConfig: config '%s' does not exist", mutationCfgName)
			return nil
		}
		return err
	}
	return kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, mutationCfgName, v1.DeleteOptions{})
}
