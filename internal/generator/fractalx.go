package generator

import (
	"fmt"
	"strings"

	"github.com/fractalx-org/fractalx-cli/internal/model"
	"github.com/fractalx-org/fractalx-cli/internal/transform"
)

func genSpecYaml(spec *model.ProjectSpec) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`project:
  groupId: %s
  artifactId: %s
  version: "%s"
  javaVersion: "%s"
  springBootVersion: %s
  fractalxVersion: %s
  description: "%s"

services:
`, spec.GroupID, spec.ArtifactID, spec.Version, spec.JavaVersion, spec.SpringBootVersion, spec.FractalXVersion, spec.Description))

	for _, svc := range spec.Services {
		schema := transform.ResolvedSchema(&svc)
		b.WriteString(fmt.Sprintf("  - name: %s\n", svc.Name))
		b.WriteString(fmt.Sprintf("    port: %d\n", svc.Port))
		b.WriteString(fmt.Sprintf("    database: %s\n", svc.DB))
		b.WriteString(fmt.Sprintf("    schema: %s\n", schema))
		if len(svc.Dependencies) > 0 {
			b.WriteString("    dependencies:\n")
			for _, dep := range svc.Dependencies {
				b.WriteString(fmt.Sprintf("      - %s\n", dep))
			}
		}
		if len(svc.Entities) > 0 {
			b.WriteString("    entities:\n")
			for _, ent := range svc.Entities {
				b.WriteString(fmt.Sprintf("      - name: %s\n", ent.Name))
				if len(ent.Fields) > 0 {
					b.WriteString("        fields:\n")
					for _, f := range ent.Fields {
						b.WriteString(fmt.Sprintf("          - %s: %s\n", f.Name, f.Type))
					}
				}
			}
		}
	}

	if len(spec.Sagas) > 0 {
		b.WriteString("\nsagas:\n")
		for _, saga := range spec.Sagas {
			b.WriteString(fmt.Sprintf("  - id: %s\n", saga.SagaID))
			b.WriteString(fmt.Sprintf("    owner: %s\n", saga.Owner))
			if saga.Compensation != "" {
				b.WriteString(fmt.Sprintf("    compensationMethod: %s\n", saga.Compensation))
			}
			b.WriteString("    timeoutMs: 30000\n")
			if len(saga.Steps) > 0 {
				b.WriteString("    steps:\n")
				for _, step := range saga.Steps {
					b.WriteString(fmt.Sprintf("      - service: %s\n", step.Service))
					b.WriteString(fmt.Sprintf("        method: %s\n", step.Method))
				}
			}
		}
	}

	ciValue := "none"
	if spec.Infra.CI {
		ciValue = "github-actions"
	}

	b.WriteString(fmt.Sprintf(`
infrastructure:
  gateway: %v
  admin: %v
  serviceRegistry: %v
  docker: %v
  kubernetes: %v
  ci: %s
  observability: %v
  sagaOrchestrator: %v

security:
  type: %s
`, spec.Infra.Gateway, spec.Infra.Admin, spec.Infra.Registry,
		spec.Infra.Docker, spec.Infra.Kubernetes, ciValue,
		spec.Infra.Observability, spec.Infra.Saga, spec.Security))

	return b.String()
}
