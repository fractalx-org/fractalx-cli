package model

// ProjectSpec is the top-level configuration for a FractalX project.
type ProjectSpec struct {
	GroupID            string
	ArtifactID         string
	Version            string
	Description        string
	JavaVersion        string // "17" | "21"
	SpringBootVersion  string // e.g. "3.3.0"
	FractalXVersion    string // e.g. "0.3.2"
	Security           string // "none" | "jwt" | "oauth2" | "apikey"
	Services           []Service
	Sagas              []Saga
	Infra              InfraConfig
}

// Service represents a bounded-context service within the monolith.
type Service struct {
	Name         string
	Port         int
	DB           string   // "h2" | "postgresql" | "mysql" | "mongodb" | "redis"
	Entities     []Entity
	Dependencies []string // names of other services this one depends on
}

// Entity is a JPA/MongoDB domain entity within a service.
type Entity struct {
	Name   string
	Fields []Field
}

// Field is a single field on an entity.
type Field struct {
	Name string
	Type string // Java type: String, Long, Integer, BigDecimal, LocalDateTime, LocalDate, UUID, Boolean, Double
}

// Saga describes a distributed saga orchestrated from a single owner service.
type Saga struct {
	SagaID       string
	Owner        string // service name
	Compensation string // optional compensation method name
	Steps        []Step
}

// Step is a single step in a distributed saga.
type Step struct {
	Service string
	Method  string
}

// InfraConfig holds infrastructure toggles.
type InfraConfig struct {
	Docker        bool
	CI            bool
	Kubernetes    bool
	Gateway       bool
	Admin         bool
	Registry      bool
	Saga          bool
	Observability bool
}
