package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/randsw/validationwebhook/pkg/kubeapi"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	tlsCert    string
	tlsKey     string
	port       int
	codecs     = serializer.NewCodecFactory(runtime.NewScheme())
	httplogger = log.New(os.Stdout, "http: ", log.LstdFlags)
)

var rootCmd = &cobra.Command{
	Use:   "validating-webhook",
	Short: "Kubernetes validating webhook example",
	Long: `Example showing how to implement a basic validating webhook in Kubernetes.

Example:
$ validating-webhook --tls-cert <tls_cert> --tls-key <tls_key> --port <port>`,
	Run: func(cmd *cobra.Command, args []string) {
		if tlsCert == "" || tlsKey == "" {
			fmt.Println("--tls-cert and --tls-key required")
			os.Exit(1)
		}
		runWebhookServer(tlsCert, tlsKey)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().StringVar(&tlsCert, "tls-cert", "", "Certificate for TLS")
	rootCmd.Flags().StringVar(&tlsKey, "tls-key", "", "Private key file for TLS")
	rootCmd.Flags().IntVar(&port, "port", 443, "Port to listen on for HTTPS traffic")
}

func runWebhookServer(certFile, keyFile string) {
	client, err := kubeapi.InitKubeApiConnection()
	if err != nil {
		httplogger.Fatalf("Fail to obtain kubeapi client")
		return
	}

	app := &application{
		client: client,
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting webhook server")

	Handler := app.setupRoutes()

	server := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		ErrorLog:     httplogger,
		Handler:      Handler,
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		panic(err)
	}
}
