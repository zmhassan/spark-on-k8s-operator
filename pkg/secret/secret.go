/*
Copyright 2017 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secret

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"

	"k8s.io/spark-on-k8s-operator/pkg/config"
)

const (
	// GoogleApplicationCredentialsEnvVar is the environment variable used by the
	// Application Default Credentials mechanism. More details can be found at
	// https://developers.google.com/identity/protocols/application-default-credentials.
	GoogleApplicationCredentialsEnvVar = "GOOGLE_APPLICATION_CREDENTIALS"
	// ServiceAccountJSONKeyFileName is the default name of the service account
	// Json key file. This name is added to the service account secret mount path to
	// form the path to the Json key file referred to by GOOGLE_APPLICATION_CREDENTIALS.
	ServiceAccountJSONKeyFileName = "key.json"
	// ServiceAccountSecretVolumeName is the name of the GCP service account secret volume.
	ServiceAccountSecretVolumeName = "gcp-service-account-secret-volume"
)

// AddSecretVolumeToPod adds a secret volume for the secret with secretName into pod.
func AddSecretVolumeToPod(secretVolumeName string, secretName string, pod *v1.Pod) {
	volume := v1.Volume{
		Name: secretVolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
}

// MountSecretToContainer mounts the secret volume with volumeName onto the mountPath into container.
func MountSecretToContainer(volumeName string, mountPath string, container *v1.Container) {
	volumeMount := v1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: mountPath,
	}
	container.VolumeMounts = append(container.VolumeMounts, volumeMount)
}

// FindGCPServiceAccountSecret finds the annotation for specifying GCP service account
// secret and returns the name and mount path of the secret if the annotation is found.
func FindGCPServiceAccountSecret(annotations map[string]string) (string, string, bool) {
	for annotation := range annotations {
		if strings.HasPrefix(annotation, config.GCPServiceAccountSecretAnnotationPrefix) {
			name := strings.TrimPrefix(annotation, config.GCPServiceAccountSecretAnnotationPrefix)
			path := annotations[annotation]
			return name, path, true
		}
	}
	return "", "", false
}

// FindGeneralSecrets finds the annotations for specifying general secrets and returns
// an map of names of the secrets to their mount paths.
func FindGeneralSecrets(annotations map[string]string) map[string]string {
	secrets := make(map[string]string)
	for annotation := range annotations {
		if strings.HasPrefix(annotation, config.GeneralSecretsAnnotationPrefix) {
			name := strings.TrimPrefix(annotation, config.GeneralSecretsAnnotationPrefix)
			path := annotations[annotation]
			secrets[name] = path
		}
	}
	return secrets
}

// MountServiceAccountSecretToContainer mounts the service account secret volume with volumeName onto
// the mountPath into container and also sets environment variable GOOGLE_APPLICATION_CREDENTIALS to
// the service account key file in the volume.
func MountServiceAccountSecretToContainer(mountPath string, container *v1.Container) {
	MountSecretToContainer(ServiceAccountSecretVolumeName, mountPath, container)
	jsonKeyFilePath := fmt.Sprintf("%s/%s", mountPath, ServiceAccountJSONKeyFileName)
	appCredentialEnvVar := v1.EnvVar{Name: GoogleApplicationCredentialsEnvVar, Value: jsonKeyFilePath}
	container.Env = append(container.Env, appCredentialEnvVar)
}
