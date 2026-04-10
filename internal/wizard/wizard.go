package wizard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fractalx-org/fractalx-cli/internal/model"
)

// Run executes the interactive 6-step wizard and returns a fully populated ProjectSpec.
func Run() (*model.ProjectSpec, error) {
	fmt.Println()
	fmt.Println("  \033[1;36mFractalX Initializr\033[0m")
	fmt.Println("  Generate a Spring Boot monolith ready for decomposition")
	fmt.Println()

	spec := &model.ProjectSpec{}

	if err := stepProject(spec); err != nil {
		return nil, err
	}
	if err := stepServices(spec); err != nil {
		return nil, err
	}
	if err := stepDependencies(spec); err != nil {
		return nil, err
	}
	if err := stepSagas(spec); err != nil {
		return nil, err
	}
	if err := stepInfra(spec); err != nil {
		return nil, err
	}
	if err := stepSecurity(spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// ──────────────────────────────────────────────────────────────
// Step 1: Project metadata
// ──────────────────────────────────────────────────────────────

func stepProject(spec *model.ProjectSpec) error {
	printStep(1, "Project", "Basic project metadata")

	qs := []*survey.Question{
		{
			Name:   "groupId",
			Prompt: &survey.Input{Message: "Group ID:", Default: "com.example"},
		},
		{
			Name:   "artifactId",
			Prompt: &survey.Input{Message: "Artifact ID:", Default: "my-platform"},
		},
		{
			Name:   "version",
			Prompt: &survey.Input{Message: "Version:", Default: "1.0.0-SNAPSHOT"},
		},
		{
			Name:   "description",
			Prompt: &survey.Input{Message: "Description:", Default: "FractalX modular monolith"},
		},
		{
			Name: "springBootVersion",
			Prompt: &survey.Select{
				Message: "Spring Boot version:",
				Options: []string{"3.3.0", "3.2.0", "3.1.0"},
				Default: "3.3.0",
			},
		},
		{
			Name: "javaVersion",
			Prompt: &survey.Select{
				Message: "Java version:",
				Options: []string{"21", "17"},
				Default: "21",
			},
		},
		{
			Name:   "fractalxVersion",
			Prompt: &survey.Input{Message: "FractalX version:", Default: "0.3.2"},
		},
	}

	answers := struct {
		GroupID           string `survey:"groupId"`
		ArtifactID        string `survey:"artifactId"`
		Version           string `survey:"version"`
		Description       string `survey:"description"`
		SpringBootVersion string `survey:"springBootVersion"`
		JavaVersion       string `survey:"javaVersion"`
		FractalXVersion   string `survey:"fractalxVersion"`
	}{}

	if err := survey.Ask(qs, &answers); err != nil {
		return err
	}

	spec.GroupID = answers.GroupID
	spec.ArtifactID = answers.ArtifactID
	spec.Version = answers.Version
	spec.Description = answers.Description
	spec.SpringBootVersion = answers.SpringBootVersion
	spec.JavaVersion = answers.JavaVersion
	spec.FractalXVersion = answers.FractalXVersion
	return nil
}

// ──────────────────────────────────────────────────────────────
// Step 2: Services
// ──────────────────────────────────────────────────────────────

func stepServices(spec *model.ProjectSpec) error {
	printStep(2, "Services", "Define bounded-context services")

	nextPort := 8081
	for {
		var addSvc bool
		label := "Add a service?"
		if len(spec.Services) > 0 {
			label = "Add another service?"
		}
		if err := survey.AskOne(&survey.Confirm{Message: label, Default: len(spec.Services) == 0}, &addSvc); err != nil {
			return err
		}
		if !addSvc {
			break
		}

		svc, err := promptService(nextPort)
		if err != nil {
			return err
		}
		if svc.Port >= nextPort {
			nextPort = svc.Port + 1
		}
		spec.Services = append(spec.Services, *svc)
	}

	if len(spec.Services) == 0 {
		fmt.Println("  \033[33m⚠  No services defined — generating a minimal project.\033[0m")
	}
	return nil
}

func promptService(defaultPort int) (*model.Service, error) {
	fmt.Println()
	qs := []*survey.Question{
		{
			Name:   "name",
			Prompt: &survey.Input{Message: "Service name (kebab-case):", Default: "order-service"},
		},
		{
			Name:   "port",
			Prompt: &survey.Input{Message: "Port:", Default: strconv.Itoa(defaultPort)},
		},
		{
			Name: "db",
			Prompt: &survey.Select{
				Message: "Database:",
				Options: []string{"h2", "postgresql", "mysql", "mongodb", "redis"},
				Default: "h2",
			},
		},
	}

	answers := struct {
		Name string `survey:"name"`
		Port string `survey:"port"`
		DB   string `survey:"db"`
	}{}

	if err := survey.Ask(qs, &answers); err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(strings.TrimSpace(answers.Port))
	if err != nil {
		port = defaultPort
	}

	svc := &model.Service{
		Name: strings.TrimSpace(answers.Name),
		Port: port,
		DB:   answers.DB,
	}

	// Entities
	for {
		var addEnt bool
		label := "Add an entity to " + svc.Name + "?"
		if len(svc.Entities) > 0 {
			label = "Add another entity?"
		}
		if err := survey.AskOne(&survey.Confirm{Message: label, Default: len(svc.Entities) == 0}, &addEnt); err != nil {
			return nil, err
		}
		if !addEnt {
			break
		}
		ent, err := promptEntity(svc.DB)
		if err != nil {
			return nil, err
		}
		svc.Entities = append(svc.Entities, *ent)
	}

	return svc, nil
}

func promptEntity(db string) (*model.Entity, error) {
	var entName string
	if err := survey.AskOne(&survey.Input{Message: "Entity name (PascalCase):", Default: "Order"}, &entName); err != nil {
		return nil, err
	}

	ent := &model.Entity{Name: strings.TrimSpace(entName)}

	javaTypes := []string{"String", "Long", "Integer", "BigDecimal", "Boolean", "LocalDateTime", "LocalDate", "UUID", "Double"}

	for {
		var addField bool
		label := "Add a field?"
		if len(ent.Fields) > 0 {
			label = "Add another field?"
		}
		if err := survey.AskOne(&survey.Confirm{Message: label, Default: true}, &addField); err != nil {
			return nil, err
		}
		if !addField {
			break
		}

		fieldAnswers := struct {
			Name string `survey:"name"`
			Type string `survey:"type"`
		}{}
		if err := survey.Ask([]*survey.Question{
			{Name: "name", Prompt: &survey.Input{Message: "Field name (camelCase):"}},
			{Name: "type", Prompt: &survey.Select{Message: "Field type:", Options: javaTypes, Default: "String"}},
		}, &fieldAnswers); err != nil {
			return nil, err
		}
		ent.Fields = append(ent.Fields, model.Field{
			Name: strings.TrimSpace(fieldAnswers.Name),
			Type: fieldAnswers.Type,
		})
	}

	return ent, nil
}

// ──────────────────────────────────────────────────────────────
// Step 3: Dependencies
// ──────────────────────────────────────────────────────────────

func stepDependencies(spec *model.ProjectSpec) error {
	if len(spec.Services) < 2 {
		return nil
	}
	printStep(3, "Dependencies", "Define service-to-service dependencies")

	allNames := make([]string, len(spec.Services))
	for i, svc := range spec.Services {
		allNames[i] = svc.Name
	}

	for i := range spec.Services {
		svc := &spec.Services[i]
		// Options are all other services
		others := make([]string, 0, len(allNames)-1)
		for _, n := range allNames {
			if n != svc.Name {
				others = append(others, n)
			}
		}
		if len(others) == 0 {
			continue
		}

		var selected []string
		if err := survey.AskOne(&survey.MultiSelect{
			Message: fmt.Sprintf("'%s' depends on:", svc.Name),
			Options: others,
		}, &selected); err != nil {
			return err
		}
		svc.Dependencies = selected
	}
	return nil
}

// ──────────────────────────────────────────────────────────────
// Step 4: Sagas
// ──────────────────────────────────────────────────────────────

func stepSagas(spec *model.ProjectSpec) error {
	printStep(4, "Sagas", "Configure distributed sagas (optional)")

	var addSagas bool
	if err := survey.AskOne(&survey.Confirm{Message: "Add distributed sagas?", Default: false}, &addSagas); err != nil {
		return err
	}
	if !addSagas {
		return nil
	}

	svcNames := make([]string, len(spec.Services))
	for i, svc := range spec.Services {
		svcNames[i] = svc.Name
	}

	for {
		saga, err := promptSaga(svcNames)
		if err != nil {
			return err
		}
		spec.Sagas = append(spec.Sagas, *saga)

		var more bool
		if err := survey.AskOne(&survey.Confirm{Message: "Add another saga?", Default: false}, &more); err != nil {
			return err
		}
		if !more {
			break
		}
	}
	return nil
}

func promptSaga(svcNames []string) (*model.Saga, error) {
	fmt.Println()
	answers := struct {
		SagaID       string `survey:"sagaId"`
		Owner        string `survey:"owner"`
		Compensation string `survey:"compensation"`
	}{}

	ownerOptions := append([]string{"(none)"}, svcNames...)

	if err := survey.Ask([]*survey.Question{
		{Name: "sagaId", Prompt: &survey.Input{Message: "Saga ID (kebab-case):", Default: "place-order-saga"}},
		{Name: "owner", Prompt: &survey.Select{Message: "Owner service:", Options: svcNames}},
		{Name: "compensation", Prompt: &survey.Input{Message: "Compensation method name (optional):"}},
	}, &answers); err != nil {
		_ = ownerOptions
		return nil, err
	}

	saga := &model.Saga{
		SagaID:       answers.SagaID,
		Owner:        answers.Owner,
		Compensation: strings.TrimSpace(answers.Compensation),
	}

	// Steps
	for {
		var addStep bool
		label := "Add a saga step?"
		if len(saga.Steps) > 0 {
			label = "Add another step?"
		}
		if err := survey.AskOne(&survey.Confirm{Message: label, Default: true}, &addStep); err != nil {
			return nil, err
		}
		if !addStep {
			break
		}

		step := struct {
			Service string `survey:"service"`
			Method  string `survey:"method"`
		}{}
		if err := survey.Ask([]*survey.Question{
			{Name: "service", Prompt: &survey.Select{Message: "Service:", Options: svcNames}},
			{Name: "method", Prompt: &survey.Input{Message: "Method name:"}},
		}, &step); err != nil {
			return nil, err
		}
		saga.Steps = append(saga.Steps, model.Step{Service: step.Service, Method: step.Method})
	}

	return saga, nil
}

// ──────────────────────────────────────────────────────────────
// Step 5: Infrastructure
// ──────────────────────────────────────────────────────────────

func stepInfra(spec *model.ProjectSpec) error {
	printStep(5, "Infrastructure", "Toggle infrastructure components")

	options := []string{
		"API Gateway",
		"Admin Dashboard",
		"Service Registry",
		"Docker Compose",
		"GitHub Actions CI",
		"Kubernetes manifests",
		"Observability (Jaeger + OTel)",
		"Saga Orchestrator",
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select infrastructure components:",
		Options: options,
		Default: []string{"API Gateway", "Admin Dashboard", "Service Registry", "Docker Compose"},
	}, &selected); err != nil {
		return err
	}

	sel := map[string]bool{}
	for _, s := range selected {
		sel[s] = true
	}

	spec.Infra = model.InfraConfig{
		Gateway:       sel["API Gateway"],
		Admin:         sel["Admin Dashboard"],
		Registry:      sel["Service Registry"],
		Docker:        sel["Docker Compose"],
		CI:            sel["GitHub Actions CI"],
		Kubernetes:    sel["Kubernetes manifests"],
		Observability: sel["Observability (Jaeger + OTel)"],
		Saga:          sel["Saga Orchestrator"],
	}
	return nil
}

// ──────────────────────────────────────────────────────────────
// Step 6: Security
// ──────────────────────────────────────────────────────────────

func stepSecurity(spec *model.ProjectSpec) error {
	printStep(6, "Security", "Authentication strategy")

	var auth string
	if err := survey.AskOne(&survey.Select{
		Message: "Authentication strategy:",
		Options: []string{"none", "jwt", "oauth2", "apikey"},
		Default: "none",
	}, &auth); err != nil {
		return err
	}
	spec.Security = auth
	return nil
}

// ──────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────

func printStep(n int, title, subtitle string) {
	fmt.Printf("\n  \033[1mStep %d — %s\033[0m  \033[2m%s\033[0m\n\n", n, title, subtitle)
}
