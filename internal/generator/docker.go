package generator

import (
	"strings"

	"github.com/fractalx-org/fractalx-cli/internal/model"
)

func genDocker(spec *model.ProjectSpec) string {
	var hasPg, hasMy, hasMg, hasRd bool
	for _, svc := range spec.Services {
		switch svc.DB {
		case "postgresql":
			hasPg = true
		case "mysql":
			hasMy = true
		case "mongodb":
			hasMg = true
		case "redis":
			hasRd = true
		}
	}
	hasObs := spec.Infra.Observability

	var services strings.Builder
	var volumes strings.Builder

	services.WriteString("services:\n")

	if hasPg {
		services.WriteString(`
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: fractalx
      POSTGRES_PASSWORD: fractalx
      POSTGRES_DB: fractalx_dev
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U fractalx"]
      interval: 10s
      retries: 5
`)
		volumes.WriteString("  postgres_data:\n")
	}

	if hasMy {
		services.WriteString(`
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: fractalx
      MYSQL_DATABASE: fractalx_dev
      MYSQL_USER: fractalx
      MYSQL_PASSWORD: fractalx
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      retries: 5
`)
		volumes.WriteString("  mysql_data:\n")
	}

	if hasMg {
		services.WriteString(`
  mongo:
    image: mongo:7.0
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db
`)
		volumes.WriteString("  mongo_data:\n")
	}

	if hasRd {
		services.WriteString(`
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
`)
		volumes.WriteString("  redis_data:\n")
	}

	if hasObs {
		services.WriteString(`
  jaeger:
    image: jaegertracing/all-in-one:1.53
    ports:
      - "16686:16686"
      - "4317:4317"
    environment:
      COLLECTOR_OTLP_ENABLED: "true"

  logger-service:
    image: fractalx/logger-service:latest
    ports:
      - "9099:9099"
`)
	}

	result := services.String()
	if volumes.Len() > 0 {
		result += "\nvolumes:\n" + volumes.String()
	}
	return result
}
