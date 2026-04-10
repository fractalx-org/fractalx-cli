package generator

import (
	"fmt"
	"strings"

	"github.com/fractalx/fractalx-cli/internal/model"
	"github.com/fractalx/fractalx-cli/internal/transform"
)

func genEntity(spec *model.ProjectSpec, svc *model.Service, ent *model.Entity) string {
	pkg := transform.ResolvedPackage(spec)
	svcPkg := transform.SvcPackage(svc)
	isMongo := svc.DB == "mongodb"

	// Compute dynamic imports
	importSet := map[string]bool{}
	for _, f := range ent.Fields {
		switch f.Type {
		case "BigDecimal":
			importSet["java.math.BigDecimal"] = true
		case "LocalDateTime":
			importSet["java.time.LocalDateTime"] = true
		case "LocalDate":
			importSet["java.time.LocalDate"] = true
		case "UUID":
			importSet["java.util.UUID"] = true
		}
	}
	var imports strings.Builder
	for imp := range importSet {
		imports.WriteString(fmt.Sprintf("import %s;\n", imp))
	}

	tableName := transform.ToSnake(ent.Name) + "s"

	// Fields block
	var fields strings.Builder
	for _, f := range ent.Fields {
		fields.WriteString(fmt.Sprintf("\n\tprivate %s %s;\n", f.Type, f.Name))
	}

	// Getters/setters
	var accessors strings.Builder
	for _, f := range ent.Fields {
		cap := transform.Capitalize(f.Name)
		accessors.WriteString(fmt.Sprintf(`
	public %s get%s() { return %s; }
	public void set%s(%s %s) { this.%s = %s; }
`, f.Type, cap, f.Name, cap, f.Type, f.Name, f.Name, f.Name))
	}

	if isMongo {
		return fmt.Sprintf(`package %s.%s;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
%s
@Document(collection = "%ss")
public class %s {

	@Id
	private String id;
%s%s}
`, pkg, svcPkg, imports.String(), strings.ToLower(ent.Name), ent.Name, fields.String(), accessors.String())
	}

	return fmt.Sprintf(`package %s.%s;

import jakarta.persistence.*;
%s
@Entity
@Table(name = "%s")
public class %s {

	@Id
	@GeneratedValue(strategy = GenerationType.IDENTITY)
	private Long id;
%s%s}
`, pkg, svcPkg, imports.String(), tableName, ent.Name, fields.String(), accessors.String())
}
