package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/randsw/validationwebhook/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHTTPRoutesUsingCorrectHTTPMethods(t *testing.T) {

	tt := []struct {
		name           string
		path           string
		wantStatusCode int
	}{
		{name: "test /healthcheck with GET method",
			path:           "/healthcheck",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "test /healthz with GET method",
			path:           "/healthz",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "test /thisdoesnotexist with GET method",
			path:           "/thisdoesnotexist",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "test /validate with GET method",
			path:           "/validate", // this exists but only POST method is allowed
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:           "test /mutate with GET method",
			path:           "/mutate", // this exists but only POST method is allowed
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}
	logger.InitLogger()
	for _, tc := range tt {
		tc := tc
		t.Run(fmt.Sprintf("testing route %v with HTTP GET method", tc.path), func(t *testing.T) {
			t.Parallel()
			app := &application{}
			srv := httptest.NewServer(app.setupRoutes())
			defer srv.Close()

			res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, tc.path))

			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			if err != nil {
				t.Fatalf("Could not send GET request to %v; %v", tc.path, err)
			}

			if err != nil {
				t.Fatalf("Could not read response, %v", err)
			}

			got := res.StatusCode

			if got != tc.wantStatusCode {
				t.Errorf("call to HTTP route %v failed; HTTP status code - want=%v got=%v", tc.path, tc.wantStatusCode, got)
			}

		})
	}

}

// TestHTTPRoutePOSTMethodValidRequest - Test with valid Deployment namespace, valid Admission review object
func TestHTTPValidateRoutePOSTMethodValidRequest(t *testing.T) {
	logger.InitLogger()
	client := fake.NewSimpleClientset()

	app := &application{
		client: client,
	}

	srv := httptest.NewServer(app.setupRoutes())
	defer srv.Close()

	f, err := os.Open("../tests/invalid-request.json")

	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("%s/validate", srv.URL)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "webhook-demo",
		},
	}

	ctx := context.Background()

	_, err = client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})

	if err != nil {
		t.Fatal("error creating namespace", err)
	}

	resp, err := http.Post(url, "application/json", f)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	//t.Log(string(body))

	output := &admissionv1.AdmissionReview{}

	if err := json.Unmarshal(body, output); err != nil {
		t.Fatalf("error unmarshaling response body into AdmissionReview object - %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("HTTP post with valid AdmissionReview failed, returned status code was not 200OK")
		return
	}

	if !output.Response.Allowed {
		t.Errorf("Admission response - want=%v, got=%v", output.Response.Allowed, true)
		return
	}

	wantMessage := "skipping validation as annotationKey validate is missing or set to false"
	gotMessage := output.Response.Result.Message

	if wantMessage != gotMessage {
		t.Errorf("Mismatch in status message - want=%q, got=%q", wantMessage, gotMessage)
		return
	}

}

func TestHTTPMutateRoutePOSTMethodValidRequest(t *testing.T) {
	logger.InitLogger()
	client := fake.NewSimpleClientset()

	app := &application{
		client: client,
	}

	srv := httptest.NewServer(app.setupRoutes())
	defer srv.Close()

	f, err := os.Open("../tests/invalid-request.json")

	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("%s/mutate", srv.URL)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "webhook-demo",
		},
	}

	ctx := context.Background()

	_, err = client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})

	if err != nil {
		t.Fatal("error creating namespace", err)
	}

	resp, err := http.Post(url, "application/json", f)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	//t.Log(string(body))

	output := &admissionv1.AdmissionReview{}

	if err := json.Unmarshal(body, output); err != nil {
		t.Fatalf("error unmarshaling response body into AdmissionReview object - %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("HTTP post with valid AdmissionReview failed, returned status code was not 200OK")
		return
	}

	if !output.Response.Allowed {
		t.Errorf("Admission response - want=%v, got=%v", output.Response.Allowed, true)
		return
	}

	wantMessage := "skipping mutation as annotationKey mutate is missing or set to false"
	gotMessage := output.Response.Result.Message

	if wantMessage != gotMessage {
		t.Errorf("Mismatch in status message - want=%q, got=%q", wantMessage, gotMessage)
		return
	}

}
