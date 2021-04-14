package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	webhookNamespace, webhookService, mutationCfgName string
)

func init() {
	webhookNamespace, _ = os.LookupEnv("WEBHOOK_NAMESPACE")
	webhookService, _ = os.LookupEnv("WEBHOOK_SERVICE")
	mutationCfgName, _ = os.LookupEnv("MUTATING_CONFIG")
}

type labelFlag map[string]string

func (h *labelFlag) String() string {
	return "label map"
}

func (h *labelFlag) Set(value string) error {
	indx := strings.Index(value, "=")
	if indx <= 0 || indx == len(value)-1 {
		return fmt.Errorf("expected format <key=<value")
	}
	key := value[:indx]
	val := value[indx+1:]
	(*h)[key] = val
	return nil
}

func main() {
	var action, certPath, keyPath string
	labels := labelFlag{}
	flag.StringVar(&action, "action", "install", "Action: [install|uninstall]. Default: install")
	flag.StringVar(&certPath, "tls.cert.path", "/etc/webhook/certs/tls.crt", "TLS certificate filepath")
	flag.StringVar(&keyPath, "tls.key.path", "/etc/webhook/certs/tls.key", "TLS private key filepath")
	flag.Var(&labels, "label", "admission config label (repetitive)")
	flag.Parse()

	ctx := context.Background()
	log.Infof("admission config action: %s", action)

	var err error
	switch action {
	case "install":
		err = install(ctx, certPath, keyPath, labels)
	case "uninstall":
		err = uninstall(ctx)
	default:
		err = fmt.Errorf("unsupported action %q", action)
	}
	if err != nil {
		log.Panic(err)
	}
}

func install(ctx context.Context, certPath, keyPath string, labels map[string]string) error {
	caBundle, err := createCert(certPath, keyPath)
	if err != nil {
		return err
	}
	if err = createMutationConfig(ctx, labels, caBundle); err != nil {
		return err
	}
	fmt.Printf("Successfully installed mutating webhook %q\n", mutationCfgName)
	return nil
}

func uninstall(ctx context.Context) error {
	if err := deleteMutationConfig(ctx); err != nil {
		return err
	}
	fmt.Printf("Successfully uninstalled mutating webhook %q\n", mutationCfgName)
	return nil
}
