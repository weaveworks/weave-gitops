/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const ApplicationKind = "Application"

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// ConfigRepo is the address of the git repository containing the automation for this application
	ConfigRepo string `json:"config_url,omitempty"`
	// URL is the address of the git repository for this application
	URL string `json:"url,omitempty"`
	// Path is the path in the repository where the k8s yaml files for this application are stored.
	Path string `json:"path,omitempty"`
	// Branch is the branch in the repository where the k8s yaml files for this application are stored.
	Branch string `json:"branch,omitempty"`
	// DeploymentType is the deployment method used to apply the manifests
	DeploymentType DeploymentType `json:"deployment_type,omitempty"`
	// SourceType is the type of repository containing the app manifests
	SourceType SourceType `json:"source_type,omitempty"`
	// HelmTargetNamespace is the namespace in which to deploy an added Helm Chart
	HelmTargetNamespace string `json:"helm_target_namespace,omitempty"`
}

// +kubebuilder:validation:Enum=helm;kustomize
type DeploymentType string

const (
	DeploymentTypeHelm      DeploymentType = "helm"
	DeploymentTypeKustomize DeploymentType = "kustomize"
)

// +kubebuilder:validation:Enum=helm;git
type SourceType string

const (
	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"
)

// SuspendAction defines the command run to pause/unpause an application
type SuspendActionType string

const (
	SuspendAction SuspendActionType = "suspend"
	ResumeAction  SuspendActionType = "resume"
)

const (
	DefaultNamespace           = "flux-system"
	DefaultSuperUserSecretHash = "admin-password-hash"
	DefaultClaimsSubject       = "wego-admin"
)

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:singular=app,path=apps

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

func (a *Application) IsHelmRepository() bool {
	return a.Spec.SourceType == SourceTypeHelm
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
