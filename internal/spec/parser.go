package spec

import (
	"fmt"
	"os"

	"github.com/fractalx/fractalx-init/internal/model"
	"gopkg.in/yaml.v3"
)

// yamlSpec mirrors the fractalx.yaml structure produced by the web UI and this CLI.
type yamlSpec struct {
	Project struct {
		GroupID           string `yaml:"groupId"`
		ArtifactID        string `yaml:"artifactId"`
		Version           string `yaml:"version"`
		JavaVersion       string `yaml:"javaVersion"`
		SpringBootVersion string `yaml:"springBootVersion"`
		Description       string `yaml:"description"`
	} `yaml:"project"`
	Services []struct {
		Name         string `yaml:"name"`
		Port         int    `yaml:"port"`
		Database     string `yaml:"database"`
		Dependencies []string `yaml:"dependencies"`
		Entities     []struct {
			Name   string `yaml:"name"`
			Fields []map[string]string `yaml:"fields"`
		} `yaml:"entities"`
	} `yaml:"services"`
	Sagas []struct {
		ID           string `yaml:"id"`
		Owner        string `yaml:"owner"`
		Compensation string `yaml:"compensationMethod"`
		Steps        []struct {
			Service string `yaml:"service"`
			Method  string `yaml:"method"`
		} `yaml:"steps"`
	} `yaml:"sagas"`
	Infrastructure struct {
		Gateway      bool   `yaml:"gateway"`
		Admin        bool   `yaml:"admin"`
		Registry     bool   `yaml:"serviceRegistry"`
		Docker       bool   `yaml:"docker"`
		Kubernetes   bool   `yaml:"kubernetes"`
		CI           string `yaml:"ci"` // "github-actions" | "none"
		Observability bool  `yaml:"observability"`
		Saga         bool   `yaml:"sagaOrchestrator"`
	} `yaml:"infrastructure"`
	Security struct {
		Type string `yaml:"type"`
	} `yaml:"security"`
}

// FromFile reads a fractalx.yaml file and returns a ProjectSpec.
func FromFile(path string) (*model.ProjectSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec file: %w", err)
	}
	return FromBytes(data)
}

// FromBytes parses fractalx.yaml content and returns a ProjectSpec.
func FromBytes(data []byte) (*model.ProjectSpec, error) {
	var y yamlSpec
	if err := yaml.Unmarshal(data, &y); err != nil {
		return nil, fmt.Errorf("parse fractalx.yaml: %w", err)
	}

	spec := &model.ProjectSpec{
		GroupID:           y.Project.GroupID,
		ArtifactID:        y.Project.ArtifactID,
		Version:           y.Project.Version,
		Description:       y.Project.Description,
		JavaVersion:       y.Project.JavaVersion,
		SpringBootVersion: y.Project.SpringBootVersion,
		Security:          y.Security.Type,
		Infra: model.InfraConfig{
			Gateway:       y.Infrastructure.Gateway,
			Admin:         y.Infrastructure.Admin,
			Registry:      y.Infrastructure.Registry,
			Docker:        y.Infrastructure.Docker,
			Kubernetes:    y.Infrastructure.Kubernetes,
			CI:            y.Infrastructure.CI == "github-actions",
			Observability: y.Infrastructure.Observability,
			Saga:          y.Infrastructure.Saga,
		},
	}

	for _, ys := range y.Services {
		svc := model.Service{
			Name:         ys.Name,
			Port:         ys.Port,
			DB:           ys.Database,
			Dependencies: ys.Dependencies,
		}
		for _, ye := range ys.Entities {
			ent := model.Entity{Name: ye.Name}
			for _, fm := range ye.Fields {
				for k, v := range fm {
					ent.Fields = append(ent.Fields, model.Field{Name: k, Type: v})
				}
			}
			svc.Entities = append(svc.Entities, ent)
		}
		spec.Services = append(spec.Services, svc)
	}

	for _, ys := range y.Sagas {
		saga := model.Saga{
			SagaID:       ys.ID,
			Owner:        ys.Owner,
			Compensation: ys.Compensation,
		}
		for _, st := range ys.Steps {
			saga.Steps = append(saga.Steps, model.Step{Service: st.Service, Method: st.Method})
		}
		spec.Sagas = append(spec.Sagas, saga)
	}

	// Apply defaults for missing fields
	if spec.Version == "" {
		spec.Version = "1.0.0-SNAPSHOT"
	}
	if spec.JavaVersion == "" {
		spec.JavaVersion = "17"
	}
	if spec.SpringBootVersion == "" {
		spec.SpringBootVersion = "3.3.0"
	}
	if spec.Security == "" {
		spec.Security = "none"
	}

	return spec, nil
}
