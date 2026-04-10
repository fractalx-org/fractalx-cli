package transform

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/fractalx-org/fractalx-cli/internal/model"
)

// javaReserved is the set of Java reserved words that must not appear as package segments.
var javaReserved = map[string]bool{
	"abstract": true, "assert": true, "boolean": true, "break": true, "byte": true,
	"case": true, "catch": true, "char": true, "class": true, "const": true,
	"continue": true, "default": true, "do": true, "double": true, "else": true,
	"enum": true, "extends": true, "final": true, "finally": true, "float": true,
	"for": true, "goto": true, "if": true, "implements": true, "import": true,
	"instanceof": true, "int": true, "interface": true, "long": true, "module": true,
	"native": true, "new": true, "package": true, "private": true, "protected": true,
	"public": true, "record": true, "return": true, "short": true, "static": true,
	"strictfp": true, "super": true, "switch": true, "synchronized": true, "this": true,
	"throw": true, "throws": true, "transient": true, "try": true, "var": true,
	"void": true, "volatile": true, "while": true,
}

var nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]`)

func safePkg(s string) string {
	s = strings.ToLower(nonAlphaNum.ReplaceAllString(s, ""))
	if javaReserved[s] {
		return s + "svc"
	}
	return s
}

// ResolvedPackage converts groupId + artifactId into a valid Java package name.
// e.g. "com.example" + "my-platform" → "com.example.myplatform"
func ResolvedPackage(spec *model.ProjectSpec) string {
	parts := strings.Split(spec.GroupID+"."+spec.ArtifactID, ".")
	safe := make([]string, 0, len(parts))
	for _, p := range parts {
		s := safePkg(p)
		if s != "" {
			safe = append(safe, s)
		}
	}
	return strings.Join(safe, ".")
}

// PackagePath converts a dot-separated package name to a file path.
// e.g. "com.example.myplatform" → "com/example/myplatform"
func PackagePath(pkg string) string {
	return strings.ReplaceAll(pkg, ".", "/")
}

// AppClassName converts artifactId to a PascalCase Spring Boot application class name.
// e.g. "my-platform" → "MyPlatformApplication"
func AppClassName(spec *model.ProjectSpec) string {
	return toPascal(spec.ArtifactID) + "Application"
}

// SvcPackage converts a service name to a Java package segment.
// e.g. "order-service" → "order"
func SvcPackage(svc *model.Service) string {
	name := svc.Name
	name = strings.TrimSuffix(name, "-service")
	return safePkg(name)
}

// SvcPrefix converts a service name to a PascalCase class prefix.
// e.g. "order-service" → "Order"
func SvcPrefix(svc *model.Service) string {
	name := svc.Name
	name = strings.TrimSuffix(name, "-service")
	return toPascal(name)
}

// ResolvedSchema returns the database schema name for a service.
// e.g. "order-service" → "order_db"
func ResolvedSchema(svc *model.Service) string {
	return SvcPackage(svc) + "_db"
}

// ToSnake converts PascalCase or camelCase to snake_case.
// e.g. "OrderItem" → "order_item", "createdAt" → "created_at"
func ToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteRune('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// ToCamel converts kebab-case to camelCase.
// e.g. "place-order-saga" → "placeOrderSaga"
func ToCamel(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		return s
	}
	var b strings.Builder
	b.WriteString(strings.ToLower(parts[0]))
	for _, p := range parts[1:] {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return b.String()
}

// SqlType maps a Java type to an SQL DDL column type.
func SqlType(javaType string) string {
	switch javaType {
	case "String":
		return "VARCHAR(255)"
	case "Long":
		return "BIGINT"
	case "Integer":
		return "INTEGER"
	case "BigDecimal":
		return "NUMERIC(19,4)"
	case "Boolean":
		return "BOOLEAN"
	case "LocalDateTime":
		return "TIMESTAMP"
	case "LocalDate":
		return "DATE"
	case "UUID":
		return "UUID"
	case "Double":
		return "DOUBLE PRECISION"
	default:
		return "VARCHAR(255)"
	}
}

// Capitalize returns the string with the first letter uppercased.
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// toPascal converts kebab-case or snake_case to PascalCase.
func toPascal(s string) string {
	parts := nonAlphaNum.Split(s, -1)
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return b.String()
}
