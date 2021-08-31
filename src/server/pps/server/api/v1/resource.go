/*
Copyright 2018 The Kubernetes Authors.

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

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// ChaosPodSpec defines the desired state of ChaosPod
type ChaosPodSpec struct {
	Template corev1.PodTemplateSpec `json:"template"`
	// +optional
	NextStop metav1.Time `json:"nextStop,omitempty"`
}

// ChaosPodStatus defines the observed state of ChaosPod.
// It should always be reconstructable from the state of the cluster and/or outside world.
type ChaosPodStatus struct {
	LastRun metav1.Time `json:"lastRun,omitempty"`
}

// +kubebuilder:object:root=true

// ChaosPod is the Schema for the randomjobs API
// +kubebuilder:printcolumn:name="next stop",type="string",JSONPath=".spec.nextStop",format="date"
// +kubebuilder:printcolumn:name="last run",type="string",JSONPath=".status.lastRun",format="date"
type ChaosPod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChaosPodSpec   `json:"spec,omitempty"`
	Status ChaosPodStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChaosPodList contains a list of ChaosPod
type ChaosPodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChaosPod `json:"items"`
}

// +kubebuilder:webhook:path=/validate-chaosapps-metamagical-io-v1-chaospod,mutating=false,failurePolicy=fail,groups=chaosapps.metamagical.io,resources=chaospods,verbs=create;update,versions=v1,name=vchaospod.kb.io

var _ webhook.Validator = &ChaosPod{}

// ValidateCreate implements webhookutil.validator so a webhook will be registered for the type
func (c *ChaosPod) ValidateCreate() error {
	//log.Info("validate create", "name", c.Name)

	if c.Spec.NextStop.Before(&metav1.Time{Time: time.Now()}) {
		return fmt.Errorf(".spec.nextStop must be later than current time")
	}
	return nil
}

// ValidateUpdate implements webhookutil.validator so a webhook will be registered for the type
func (c *ChaosPod) ValidateUpdate(old runtime.Object) error {
	//log.Info("validate update", "name", c.Name)

	if c.Spec.NextStop.Before(&metav1.Time{Time: time.Now()}) {
		return fmt.Errorf(".spec.nextStop must be later than current time")
	}

	oldC, ok := old.(*ChaosPod)
	if !ok {
		return fmt.Errorf("expect old object to be a %T instead of %T", oldC, old)
	}
	if c.Spec.NextStop.After(oldC.Spec.NextStop.Add(time.Hour)) {
		return fmt.Errorf("it is not allowed to delay.spec.nextStop for more than 1 hour")
	}
	return nil
}

// ValidateDelete implements webhookutil.validator so a webhook will be registered for the type
func (c *ChaosPod) ValidateDelete() error {
	//log.Info("validate delete", "name", c.Name)

	if c.Spec.NextStop.Before(&metav1.Time{Time: time.Now()}) {
		return fmt.Errorf(".spec.nextStop must be later than current time")
	}
	return nil
}

// +kubebuilder:webhook:path=/mutate-chaosapps-metamagical-io-v1-chaospod,mutating=true,failurePolicy=fail,groups=chaosapps.metamagical.io,resources=chaospods,verbs=create;update,versions=v1,name=mchaospod.kb.io

var _ webhook.Defaulter = &ChaosPod{}

// Default implements webhookutil.defaulter so a webhook will be registered for the type
func (c *ChaosPod) Default() {
	//log.Info("default", "name", c.Name)

	if c.Spec.NextStop.Before(&metav1.Time{Time: time.Now()}) {
		c.Spec.NextStop = metav1.Time{Time: time.Now().Add(time.Minute)}
	}
}

func init() {
	SchemeBuilder.Register(&ChaosPod{}, &ChaosPodList{})
}
