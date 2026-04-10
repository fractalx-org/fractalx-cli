package generator

import (
	"fmt"
	"strings"

	"github.com/fractalx/fractalx-cli/internal/model"
)

func genPom(spec *model.ProjectSpec) string {
	var hasJpa, hasPg, hasMy, hasH2, hasMg, hasRd, hasFly bool
	for _, svc := range spec.Services {
		switch svc.DB {
		case "postgresql":
			hasPg = true
			hasJpa = true
			hasFly = true
		case "mysql":
			hasMy = true
			hasJpa = true
			hasFly = true
		case "h2":
			hasH2 = true
			hasJpa = true
		case "mongodb":
			hasMg = true
		case "redis":
			hasRd = true
		}
	}
	isJwt := spec.Security == "jwt"
	isOAuth := spec.Security == "oauth2"

	var deps strings.Builder

	if hasJpa {
		deps.WriteString(`
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-data-jpa</artifactId>
		</dependency>`)
	}
	if hasPg {
		deps.WriteString(`
		<dependency>
			<groupId>org.postgresql</groupId>
			<artifactId>postgresql</artifactId>
			<scope>runtime</scope>
		</dependency>`)
	}
	if hasMy {
		deps.WriteString(`
		<dependency>
			<groupId>com.mysql</groupId>
			<artifactId>mysql-connector-j</artifactId>
			<scope>runtime</scope>
		</dependency>`)
	}
	if hasH2 {
		deps.WriteString(`
		<dependency>
			<groupId>com.h2database</groupId>
			<artifactId>h2</artifactId>
			<scope>runtime</scope>
		</dependency>`)
	}
	if hasFly {
		deps.WriteString(`
		<dependency>
			<groupId>org.flywaydb</groupId>
			<artifactId>flyway-core</artifactId>
		</dependency>`)
		if hasMy {
			deps.WriteString(`
		<dependency>
			<groupId>org.flywaydb</groupId>
			<artifactId>flyway-mysql</artifactId>
		</dependency>`)
		}
	}
	if hasMg {
		deps.WriteString(`
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-data-mongodb</artifactId>
		</dependency>`)
	}
	if hasRd {
		deps.WriteString(`
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-data-redis</artifactId>
		</dependency>`)
	}
	if isJwt {
		deps.WriteString(`
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-security</artifactId>
		</dependency>
		<dependency>
			<groupId>io.jsonwebtoken</groupId>
			<artifactId>jjwt-api</artifactId>
			<version>0.12.3</version>
		</dependency>
		<dependency>
			<groupId>io.jsonwebtoken</groupId>
			<artifactId>jjwt-impl</artifactId>
			<version>0.12.3</version>
			<scope>runtime</scope>
		</dependency>
		<dependency>
			<groupId>io.jsonwebtoken</groupId>
			<artifactId>jjwt-jackson</artifactId>
			<version>0.12.3</version>
			<scope>runtime</scope>
		</dependency>`)
	}
	if isOAuth {
		deps.WriteString(`
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-oauth2-resource-server</artifactId>
		</dependency>`)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>

	<parent>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-parent</artifactId>
		<version>%s</version>
		<relativePath/>
	</parent>

	<groupId>%s</groupId>
	<artifactId>%s</artifactId>
	<version>%s</version>
	<name>%s</name>
	<description>%s</description>

	<properties>
		<java.version>%s</java.version>
	</properties>

	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
		</dependency>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-actuator</artifactId>
		</dependency>
		<dependency>
			<groupId>org.fractalx</groupId>
			<artifactId>fractalx-annotations</artifactId>
			<version>0.3.2</version>
		</dependency>
		<dependency>
			<groupId>org.fractalx</groupId>
			<artifactId>fractalx-runtime</artifactId>
			<version>0.3.2</version>
		</dependency>
		<dependency>
			<groupId>org.projectlombok</groupId>
			<artifactId>lombok</artifactId>
			<optional>true</optional>
		</dependency>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-validation</artifactId>
		</dependency>%s
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-test</artifactId>
			<scope>test</scope>
		</dependency>
		<dependency>
			<groupId>com.h2database</groupId>
			<artifactId>h2</artifactId>
			<scope>test</scope>
		</dependency>
	</dependencies>

	<build>
		<plugins>
			<plugin>
				<groupId>org.springframework.boot</groupId>
				<artifactId>spring-boot-maven-plugin</artifactId>
				<configuration>
					<excludes>
						<exclude>
							<groupId>org.projectlombok</groupId>
							<artifactId>lombok</artifactId>
						</exclude>
					</excludes>
				</configuration>
			</plugin>
			<plugin>
				<groupId>org.fractalx</groupId>
				<artifactId>fractalx-maven-plugin</artifactId>
				<version>0.3.2</version>
			</plugin>
		</plugins>
	</build>

</project>
`,
		spec.SpringBootVersion,
		spec.GroupID,
		spec.ArtifactID,
		spec.Version,
		spec.ArtifactID,
		spec.Description,
		spec.JavaVersion,
		deps.String(),
	)
}
