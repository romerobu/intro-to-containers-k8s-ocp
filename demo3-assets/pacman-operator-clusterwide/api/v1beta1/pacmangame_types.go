/*


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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PacmanGameSpec defines the desired state of PacmanGame
type PacmanGameSpec struct {
	Replicas   int32  `json:"replicas"`
	AppVersion string `json:"appVersion,omitempty"`
}

// PacmanGameStatus defines the observed state of PacmanGame
type PacmanGameStatus struct {
	AppPods    []string           `json:"appPods"`
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PacmanGame is the Schema for the PacmanGames API
type PacmanGame struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PacmanGameSpec   `json:"spec,omitempty"`
	Status PacmanGameStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PacmanGameList contains a list of PacmanGame
type PacmanGameList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PacmanGame `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PacmanGame{}, &PacmanGameList{})
}

// Conditions
const (
	// ConditionTypePacmanGameDeploymentNotReady indicates if the Reverse Words Deployment is not ready

	ConditionTypePacmanGameDeploymentNotReady string = "PacmanGameDeploymentNotReady"

	// ConditionTypeReady indicates if the Reverse Words Deployment is ready
	ConditionTypeReady string = "Ready"
)
