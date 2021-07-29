/*
Copyright 2021 Wim Henderickx.

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

package v1

/*
import (
	nddv1 "github.com/netw-device-driver/ndd-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A TargetConfigSpec defines the desired state of a TargetConfig.
type TargetConfigSpec struct {
	// Name required to reference the network node name.
	Name string `json:"name"`
}

// A TargetConfigStatus reflects the observed state of a TargetConfig.
type TargetConfigStatus struct {
	nddv1.TargetConfigStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A TargetConfig configures a Template provider.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentials.secretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster
type TargetConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetConfigSpec   `json:"spec"`
	Status TargetConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TargetConfigList contains a list of TargetConfig.
type TargetConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetConfig `json:"items"`
}

// +kubebuilder:object:root=true

// A TargetConfigUsage indicates that a resource is using a TargetConfig.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONFIG-NAME",type="string",JSONPath=".TargetConfigRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Cluster,categories={ndd,target,srl}
type TargetConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	nddv1.TargetConfigUsage `json:",inline"`
}

// +kubebuilder:object:root=true

// TargetConfigUsageList contains a list of TargetConfigUsage
type TargetConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetConfigUsage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TargetConfig{}, &TargetConfigList{})
	SchemeBuilder.Register(&TargetConfigUsage{}, &TargetConfigUsageList{})
}
*/
