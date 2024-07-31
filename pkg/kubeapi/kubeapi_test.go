package kubeapi

import (
	"context"
	"testing"

	"github.com/randsw/validationwebhook/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_application_CheckNamespaceAnnotationTrue(t *testing.T) {

	tests := []struct {
		name              string
		annotationKey     string
		annotationValue   string
		annotationToCheck string
		wantErr           bool
		want              bool
		namespaceName     string
		createNameSpace   bool
	}{
		{
			name:              "check with annotationKey example.com/validate set to true",
			annotationKey:     "example.com/validate",
			annotationValue:   "true",
			annotationToCheck: "example.com/validate",
			wantErr:           false,
			want:              true,
			namespaceName:     "test-namespace1",
			createNameSpace:   true,
		},
		{
			name:              "check with annotationKey example.com/validate set to false",
			annotationKey:     "example.com/validate",
			annotationValue:   "false",
			annotationToCheck: "example.com/validate",
			wantErr:           false,
			want:              false,
			namespaceName:     "test-namespace2",
			createNameSpace:   true,
		},
		{
			name:              "check with a non-existent namespace",
			annotationKey:     "example.com/validate",
			annotationValue:   "false",
			annotationToCheck: "example.com/validate",
			wantErr:           true,
			want:              false,
			namespaceName:     "test-namespace2",
			createNameSpace:   false,
		},
	}
	logger.InitLogger()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			namespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.namespaceName,
					Annotations: map[string]string{
						tt.annotationKey: tt.annotationValue,
					},
				},
			}
			client := fake.NewSimpleClientset()
			if tt.createNameSpace {

				_, err := client.CoreV1().Namespaces().Create(context.Background(), &namespace, metav1.CreateOptions{})

				if err != nil {
					t.Errorf("failed to create namespace - %v", tt.namespaceName)
					return
				}
			}

			got, err := CheckNamespaceAnnotationTrue(client, tt.annotationToCheck, tt.namespaceName)

			t.Log("function call returned", got, err)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckNamespaceAnnotationTrue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("CheckNamespaceAnnotationTrue() - return value - got= %v, want %v", got, tt.want)
			}

		})
	}
}
