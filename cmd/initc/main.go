package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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

func main() {
	var action, certPath, keyPath string
	flag.StringVar(&action, "action", "install", "Action: [install|uninstall]. Default: install")
	flag.StringVar(&certPath, "tls.cert.path", "/etc/webhook/certs/tls.crt", "TLS certificate filepath")
	flag.StringVar(&keyPath, "tls.key.path", "/etc/webhook/certs/tls.key", "TLS private key filepath")
	flag.Parse()

	var err error
	switch action {
	case "install":
		err = install(certPath, keyPath)
	case "uninstall":
		err = uninstall()
	default:
		err = fmt.Errorf("unsupported action %q", action)
	}
	if err != nil {
		log.Panic(err)
	}
}

func install(certPath, keyPath string) error {
	caBundle, err := createCert(certPath, keyPath)
	if err != nil {
		return err
	}
	if err = createMutationConfig(context.Background(), caBundle); err != nil {
		return err
	}
	fmt.Printf("Successfully installed mutating webhook %q\n", mutationCfgName)
	return nil
}

func uninstall() error {
	if err := deleteMutationConfig(context.Background()); err != nil {
		return err
	}
	fmt.Printf("Successfully uninstalled mutating webhook %q\n", mutationCfgName)
	return nil
}
