package generator

import (
	"strings"

	"github.com/fractalx-org/fractalx-cli/internal/model"
)

func genCI(spec *model.ProjectSpec) string {
	var hasPg, hasMy bool
	for _, svc := range spec.Services {
		if svc.DB == "postgresql" {
			hasPg = true
		}
		if svc.DB == "mysql" {
			hasMy = true
		}
	}

	var dbServices strings.Builder
	if hasPg {
		dbServices.WriteString(`
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: fractalx
          POSTGRES_PASSWORD: fractalx
          POSTGRES_DB: fractalx_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-retries 5
`)
	}
	if hasMy {
		dbServices.WriteString(`
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: fractalx
          MYSQL_DATABASE: fractalx_test
          MYSQL_USER: fractalx
          MYSQL_PASSWORD: fractalx
        ports:
          - 3306:3306
        options: >-
          --health-cmd "mysqladmin ping -h localhost"
          --health-interval 10s
          --health-retries 5
`)
	}

	return `name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
` + dbServices.String() + `
    steps:
      - uses: actions/checkout@v4

      - name: Set up Java ` + spec.JavaVersion + `
        uses: actions/setup-java@v4
        with:
          java-version: '` + spec.JavaVersion + `'
          distribution: temurin
          cache: maven

      - name: Build and test
        run: mvn -B verify --no-transfer-progress

      - name: Upload test results
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: test-results
          path: target/surefire-reports/
`
}
