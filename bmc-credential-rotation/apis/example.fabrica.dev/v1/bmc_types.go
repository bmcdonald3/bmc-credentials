package v1

import (
	"time"

	"github.com/openchami/fabrica/pkg/fabrica"
)

// Bmc represents the complete Kubernetes-style resource.
type Bmc struct {
	APIVersion string           `json:"apiVersion" validate:"required"`
	Kind       string           `json:"kind" validate:"required"`
	Metadata   fabrica.Metadata `json:"metadata"`
	Spec       BmcSpec          `json:"spec" validate:"required"`
	Status     BmcStatus        `json:"status,omitempty"`
}

// Credentials defines the authentication payload required to connect to the hardware.
type Credentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// BmcSpec represents the desired state provided by the user via the API payload.
type BmcSpec struct {
	TargetIdentifier string      `json:"targetIdentifier" validate:"required"`
	TargetIP         string      `json:"targetIp" validate:"required,ip"`
	CurrentAuth      Credentials `json:"currentAuth" validate:"required"`
	DesiredAuth      Credentials `json:"desiredAuth" validate:"required"`
}

// BmcStatus represents the observed state managed exclusively by the asynchronous Reconciliation Controller.
type BmcStatus struct {
	Phase       string     `json:"phase" validate:"omitempty,oneof=Pending Reconciling Aligned Failed"`
	Message     string     `json:"message,omitempty"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
}

func (r *Bmc) GetKind() string { return "Bmc" }
func (r *Bmc) GetName() string { return r.Metadata.Name }
func (r *Bmc) GetUID() string  { return r.Metadata.UID }
func (r *Bmc) IsHub()          {}
