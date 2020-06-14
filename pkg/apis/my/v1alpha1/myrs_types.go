package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MyRSSpec defines the desired state of MyRS
type MyRSSpec struct {
	Replicas *int32 `json:"replicas,omitempty"`

	Selector MyRSSpecSelector `json:"selector,omitempty"`
	Template v1.PodTemplateSpec `json:"template,omitempty"`
}

// MyRSStatus defines the observed state of MyRS
type MyRSStatus struct {
	Replicas int32    `json:"replicas"`
	PodNames []string `json:"podNames"`
}


type MyRSSpecSelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MyRS is the Schema for the myrs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=myrs,scope=Namespaced
type MyRS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyRSSpec   `json:"spec,omitempty"`
	Status MyRSStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MyRSList contains a list of MyRS
type MyRSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyRS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyRS{}, &MyRSList{})
}
