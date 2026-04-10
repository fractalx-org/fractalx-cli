package generator

import (
	"fmt"
	"strings"

	"github.com/fractalx/fractalx-init/internal/model"
	"github.com/fractalx/fractalx-init/internal/transform"
)

func genAppYml(spec *model.ProjectSpec) string {
	pkg := transform.ResolvedPackage(spec)

	var securityBlock string
	switch spec.Security {
	case "jwt":
		securityBlock = `
fractalx:
  security:
    enabled: false
    bearer:
      enabled: true
`
	case "oauth2":
		securityBlock = `
fractalx:
  security:
    enabled: false
    oauth2:
      enabled: true
      jwks-uri: ${OAUTH2_JWKS_URI:http://localhost:8080/realms/fractalx/protocol/openid-connect/certs}
`
	case "apikey":
		securityBlock = `
fractalx:
  security:
    enabled: false
    api-key:
      enabled: true
`
	}

	return fmt.Sprintf(`spring:
  application:
    name: %s
  profiles:
    active: dev
%s
fractalx:
  gateway-port: 9999
  admin-port: 9090
  registry-url: ${FRACTALX_REGISTRY_URL:http://localhost:8761}

management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics,prometheus

logging:
  level:
    root: INFO
    %s: DEBUG
    org.fractalx: DEBUG
`, spec.ArtifactID, securityBlock, pkg)
}

func genDevYml() string {
	return `spring:
  h2:
    console:
      enabled: true
      path: /h2-console
  jpa:
    show-sql: true
    properties:
      hibernate:
        format_sql: true

logging:
  level:
    root: DEBUG
    org.springframework.web: DEBUG
    org.fractalx: DEBUG
`
}

func genTestYml(spec *model.ProjectSpec) string {
	pkg := transform.ResolvedPackage(spec)

	var hasJpa bool
	for _, svc := range spec.Services {
		if svc.DB != "mongodb" && svc.DB != "redis" {
			hasJpa = true
			break
		}
	}

	var datasourceBlock string
	if hasJpa {
		datasourceBlock = `  datasource:
    url: jdbc:h2:mem:testdb;DB_CLOSE_DELAY=-1;MODE=MySQL
    driver-class-name: org.h2.Driver
    username: sa
    password: ''
  jpa:
    hibernate:
      ddl-auto: create-drop
    show-sql: false
    database-platform: org.hibernate.dialect.H2Dialect
  h2:
    console:
      enabled: false
`
	}

	return fmt.Sprintf(`spring:
%s
fractalx:
  registry-url: ''
  logger-url: ''
  otel-endpoint: ''

management:
  endpoints:
    web:
      exposure:
        include: health

logging:
  level:
    root: WARN
    %s: INFO
`, datasourceBlock, pkg)
}

func genFractalxConfig(spec *model.ProjectSpec) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`fractalx:
  spring-boot-version: %s
  registry-url: http://localhost:8761
  gateway-port: 9999
  admin-port: 9090
  logger-url: http://localhost:9099
  otel-endpoint: http://localhost:4317
  services:
`, spec.SpringBootVersion))

	for _, svc := range spec.Services {
		schema := transform.ResolvedSchema(&svc)
		b.WriteString(fmt.Sprintf("    %s:\n", svc.Name))
		b.WriteString("      datasource:\n")
		switch svc.DB {
		case "postgresql":
			b.WriteString(fmt.Sprintf("        url: jdbc:postgresql://localhost:5432/%s\n", schema))
			b.WriteString("        username: fractalx\n")
			b.WriteString("        password: fractalx\n")
		case "mysql":
			b.WriteString(fmt.Sprintf("        url: jdbc:mysql://localhost:3306/%s\n", schema))
			b.WriteString("        username: fractalx\n")
			b.WriteString("        password: fractalx\n")
		case "mongodb":
			b.WriteString(fmt.Sprintf("        uri: mongodb://localhost:27017/%s\n", schema))
		case "redis":
			b.WriteString("        host: localhost\n")
			b.WriteString("        port: 6379\n")
		default: // h2
			b.WriteString(fmt.Sprintf("        url: jdbc:h2:mem:%s;DB_CLOSE_DELAY=-1\n", schema))
		}
	}

	return b.String()
}
