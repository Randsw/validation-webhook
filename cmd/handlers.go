package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/randsw/validationwebhook/pkg/kubeapi"
	"github.com/randsw/validationwebhook/pkg/logger"
	"go.uber.org/zap"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type application struct {
	client kubernetes.Interface
}

func writeErrorMessage(w http.ResponseWriter, msg string, code int) {

	w.Header().Set("Content-Type", "application/json")
	logger.Error(msg)
	msg = fmt.Sprintf(`{"error": "%v"}`, msg)
	http.Error(w, msg, code)

}

func GetHealth(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := map[string]string{
		"app_name": "Validation Webhook for Kubernetes clusters",
		"status":   "OK",
	}
	if err := enc.Encode(resp); err != nil {
		logger.Error("Error while encoding JSON response", zap.String("err", err.Error()))
	}
}

func (app *application) Validate(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		logger.Error("Only POST method allowed")
		writeErrorMessage(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		logger.Error("expected application/json content-type")
		writeErrorMessage(w, "expected application/json content-type", http.StatusBadRequest)
		return
	}

	// Webhooks are sent a POST request, with Content-Type: application/json, with
	// an AdmissionReview API object in the admission.k8s.io API group serialized to JSON as the body.
	input := admissionv1.AdmissionReview{}

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		writeErrorMessage(w, "Unable to decode the POST request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// check for various nil or empty values
	if input.Request == nil || input.Request.RequestKind == nil {
		logger.Error("Request object is nil")
		writeErrorMessage(w, "invalid request", http.StatusBadRequest)
		return
	}
	switch input.Request.RequestKind.Kind {
	case "Deployment":
		logger.Info("Request came from object type of Deployment")

		deploy := appsv1.Deployment{}

		var requestAllowed bool
		var respMsg string

		if len(input.Request.Object.Raw) == 0 {
			writeErrorMessage(w, "empty Deployment object in the request", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(input.Request.Object.Raw, &deploy); err != nil {
			writeErrorMessage(w, "Unable to marshal the raw payload into Deployment object: "+err.Error(), http.StatusBadRequest)
			return
		}

		// checking if the annotationKey "validate" exists with a value of true
		ok, err := kubeapi.CheckNamespaceAnnotationTrue(app.client, "validate", deploy.Namespace)

		if err != nil {
			writeErrorMessage(w, "Unable to check annotationKey on the namespace "+err.Error(), http.StatusInternalServerError)
			return
		}
		// if the annotationKey was not preset or was set to false
		if !ok {
			logger.Info("skipping validation of the Deployment", zap.String("Deployment Name", deploy.Name), zap.String("Deployment Namespace", deploy.Namespace))
			requestAllowed = true
			respMsg = "skipping validation as annotationKey " + "validate" + " is missing or set to false"
		}

		if ok && len(deploy.ObjectMeta.Labels) > 0 {

			if val, ok := deploy.ObjectMeta.Labels["team"]; ok {
				if val != "" {
					requestAllowed = true
					respMsg = "Allowed as label " + "team" + " is present in the Deployment"
				}
				logger.Info("Allowed Deployment because label is present in the Deployment", zap.String("Deployment Name", deploy.Name), zap.String("Deployment Namespace", deploy.Namespace))
			} else {
				requestAllowed = false
				respMsg = "Denied because the Deployment is missing label " + "team"
			}
		}
		output := admissionv1.AdmissionReview{

			Response: &admissionv1.AdmissionResponse{
				UID:     input.Request.UID,
				Allowed: requestAllowed,
				Result: &metav1.Status{
					Message: respMsg,
				},
			},
		}
		output.TypeMeta.Kind = input.TypeMeta.Kind
		output.TypeMeta.APIVersion = input.TypeMeta.APIVersion

		w.Header().Set("Content-Type", "application/json")

		resp, err := json.Marshal(output)

		if err != nil {
			writeErrorMessage(w, "Unable to marshal the json object: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(resp); err != nil {
			writeErrorMessage(w, "Unable to send HTTP response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}


func (app *application) Mutate(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		logger.Error("Only POST method allowed")
		writeErrorMessage(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		logger.Error("expected application/json content-type")
		writeErrorMessage(w, "expected application/json content-type", http.StatusBadRequest)
		return
	}

	// Webhooks are sent a POST request, with Content-Type: application/json, with
	// an AdmissionReview API object in the admission.k8s.io API group serialized to JSON as the body.
	input := admissionv1.AdmissionReview{} //#TODO

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		writeErrorMessage(w, "Unable to decode the POST request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// check for various nil or empty values
	if input.Request == nil || input.Request.RequestKind == nil {
		logger.Error("Request object is nil")
		writeErrorMessage(w, "invalid request", http.StatusBadRequest)
		return
	}
	switch input.Request.RequestKind.Kind {
	case "Deployment":
		logger.Info("Request came from object type of Deployment")

		deploy := appsv1.Deployment{}

		var requestAllowed bool
		var respMsg string

		if len(input.Request.Object.Raw) == 0 {
			writeErrorMessage(w, "empty Deployment object in the request", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(input.Request.Object.Raw, &deploy); err != nil {
			writeErrorMessage(w, "Unable to marshal the raw payload into Deployment object: "+err.Error(), http.StatusBadRequest)
			return
		}

		// checking if the annotationKey "mutate" exists with a value of true
		ok, err := kubeapi.CheckNamespaceAnnotationTrue(app.client, "mutate", deploy.Namespace)

		if err != nil {
			writeErrorMessage(w, "Unable to check annotationKey on the namespace "+err.Error(), http.StatusInternalServerError)
			return
		}
		// if the annotationKey was not preset or was set to false
		if !ok {
			logger.Info("skipping validation of the Deployment", zap.String("Deployment Name", deploy.Name), zap.String("Deployment Namespace", deploy.Namespace))
			requestAllowed = true
			respMsg = "skipping validation as annotationKey " + "mutate" + " is missing or set to false"
		}

		if ok && len(deploy.ObjectMeta.Labels) > 0 {

			if val, ok := deploy.ObjectMeta.Labels["team"]; ok {
				if val != "" {
					requestAllowed = true
					respMsg = "Allowed as label " + "team" + " is present in the Deployment"
				}
				logger.Info("Allowed Deployment because label is present in the Deployment", zap.String("Deployment Name", deploy.Name), zap.String("Deployment Namespace", deploy.Namespace))
			} else {
				requestAllowed = false
				respMsg = "Denied because the Deployment is missing label " + "team"
			}
		}
		output := admissionv1.AdmissionReview{

			Response: &admissionv1.AdmissionResponse{
				UID:     input.Request.UID,
				Allowed: requestAllowed,
				Result: &metav1.Status{
					Message: respMsg,
				},
			},
		}
		output.TypeMeta.Kind = input.TypeMeta.Kind
		output.TypeMeta.APIVersion = input.TypeMeta.APIVersion

		w.Header().Set("Content-Type", "application/json")

		resp, err := json.Marshal(output)

		if err != nil {
			writeErrorMessage(w, "Unable to marshal the json object: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(resp); err != nil {
			writeErrorMessage(w, "Unable to send HTTP response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}