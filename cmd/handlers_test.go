package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/randsw/validationwebhook/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	admissionv1 "k8s.io/api/admission/v1"
)

func CreateNamespace(t *testing.T, namespace string, annotations map[string]string, client *fake.Clientset) {

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespace,
			Annotations: annotations,
		},
	}

	ctx := context.Background()

	n, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})

	fmt.Println(n.Annotations)

	if err != nil {
		t.Fatal("error creating namespace", err)
	}

}

func TestValidateWebhookHandler(t *testing.T) {

	tt := []struct {
		name            string
		allowed         bool
		sourceJsonFile  string
		annotationKey   string
		annotationValue string
		statusCode      int
	}{
		{
			name:            "Deployment is missing label team and namespace has correct annotations",
			allowed:         false,
			sourceJsonFile:  "../tests/invalid-request.json",
			annotationKey:   "validate",
			annotationValue: "true",
			statusCode:      http.StatusOK,
		},
		{
			name:            "Deployment has correct labels and namespace has correct annotations",
			allowed:         true,
			sourceJsonFile:  "../tests/valid-request.json",
			annotationKey:   "validate",
			annotationValue: "true",
			statusCode:      http.StatusOK,
		},
		{
			name:            "Deployment has correct labels but annotation is set to false",
			allowed:         true,
			sourceJsonFile:  "../tests/valid-request.json",
			annotationKey:   "validate",
			annotationValue: "false",
			statusCode:      http.StatusOK,
		},
		{
			name:            "Test with empty Admission request object",
			allowed:         false, // this field is not checked as the response does not contain valid Response
			sourceJsonFile:  "../tests/empty-request.json",
			annotationKey:   "validate",
			annotationValue: "false",
			statusCode:      http.StatusBadRequest,
		},
	}
	logger.InitLogger()
	for _, tc := range tt {
		tc := tc // capture inner variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := fake.NewSimpleClientset()

			app := &application{
				client: client,
			}

			testAnnotations := make(map[string]string)

			testAnnotations[tc.annotationKey] = tc.annotationValue

			CreateNamespace(t, "webhook-demo", testAnnotations, client)

			f, err := os.Open(tc.sourceJsonFile)

			defer func() {
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			if err != nil {
				t.Fatalf("Failed to load input json file %v", err.Error())
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(app.Validate)

			// send Admission review loaded from the json file
			req, err := http.NewRequest("POST", "/validate", f)
			req.Header.Set("Content-Type", "application/json")

			if err != nil {
				t.Errorf("Failed to create the request object %v", err.Error())
				return
			}

			handler.ServeHTTP(rr, req)

			t.Logf("status code= %v", rr.Code)

			if rr.Code != tc.statusCode {
				t.Errorf("HTTP status code mismatch want=%v, got=%v", tc.statusCode, rr.Code)
				return
			}

			// we only marshal valid json responses from the server
			// if we did not receive a valid admissionv1.AdmissionReview{} object
			// then we don't need to decode it
			if rr.Code != 200 {
				return
			}

			result := admissionv1.AdmissionReview{}
			err = json.NewDecoder(rr.Body).Decode(&result)

			if err != nil {
				t.Errorf("Failed to decode the Json response to AdmissionReview object %v", err.Error())
				return
			}

			//t.Log(result)

			admissionReviewReqAllowed := result.Response.Allowed

			if admissionReviewReqAllowed != tc.allowed {
				t.Errorf("AdmissionReview.Request.Allowed field: want=%v got=%v", tc.allowed, admissionReviewReqAllowed)
			}

		})
	}

}
