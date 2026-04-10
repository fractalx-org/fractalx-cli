package generator

import (
	"fmt"
	"strings"

	"github.com/fractalx/fractalx-init/internal/model"
	"github.com/fractalx/fractalx-init/internal/transform"
)

func genServiceClass(spec *model.ProjectSpec, svc *model.Service) string {
	pkg := transform.ResolvedPackage(spec)
	svcPkg := transform.SvcPackage(svc)
	prefix := transform.SvcPrefix(svc)
	isMongo := svc.DB == "mongodb"

	idType := "Long"
	if isMongo {
		idType = "String"
	}

	// Build constructor params and fields
	var fieldDecls, ctorParams, ctorAssigns strings.Builder
	var ctorParts []string

	for _, ent := range svc.Entities {
		rVar := strings.ToLower(ent.Name[:1]) + ent.Name[1:] + "Repository"
		fieldDecls.WriteString(fmt.Sprintf("\tprivate final %sRepository %s;\n", ent.Name, rVar))
		ctorParts = append(ctorParts, fmt.Sprintf("%sRepository %s", ent.Name, rVar))
		ctorAssigns.WriteString(fmt.Sprintf("\t\tthis.%s = %s;\n", rVar, rVar))
	}

	// Dependencies on other services
	for _, dep := range svc.Dependencies {
		depSvc := &model.Service{Name: dep}
		depPrefix := transform.SvcPrefix(depSvc)
		depVar := strings.ToLower(depPrefix[:1]) + depPrefix[1:] + "Service"
		fieldDecls.WriteString(fmt.Sprintf("\tprivate final %sService %s;\n", depPrefix, depVar))
		ctorParts = append(ctorParts, fmt.Sprintf("%sService %s", depPrefix, depVar))
		ctorAssigns.WriteString(fmt.Sprintf("\t\tthis.%s = %s;\n", depVar, depVar))
	}

	if len(ctorParts) > 0 {
		ctorParams.WriteString(strings.Join(ctorParts, ",\n\t\t\t"))
	}

	// CRUD methods per entity
	var methods strings.Builder
	for _, ent := range svc.Entities {
		rVar := strings.ToLower(ent.Name[:1]) + ent.Name[1:] + "Repository"
		varName := strings.ToLower(ent.Name[:1]) + ent.Name[1:]

		methods.WriteString(fmt.Sprintf(`
	@org.springframework.transaction.annotation.Transactional(readOnly = true)
	public java.util.List<%s> findAll%ss() {
		return %s.findAll();
	}

	@org.springframework.transaction.annotation.Transactional(readOnly = true)
	public %s find%sById(%s id) {
		return %s.findById(id).orElseThrow(() -> new IllegalArgumentException("%s not found: " + id));
	}

	@org.springframework.transaction.annotation.Transactional
	public %s create%s(%s %s) {
		return %s.save(%s);
	}

	@org.springframework.transaction.annotation.Transactional
	public void delete%s(%s id) {
		%s.deleteById(id);
	}
`, ent.Name, ent.Name, rVar,
			ent.Name, ent.Name, idType, rVar, ent.Name,
			ent.Name, ent.Name, ent.Name, varName, rVar, varName,
			ent.Name, idType, rVar))
	}

	// Saga methods for sagas owned by this service
	var sagaImports strings.Builder
	hasSagaImport := false
	for _, saga := range spec.Sagas {
		if saga.Owner != svc.Name {
			continue
		}
		if !hasSagaImport {
			sagaImports.WriteString("import org.fractalx.annotations.DistributedSaga;\n")
			hasSagaImport = true
		}
		methodName := transform.ToCamel(saga.SagaID)
		compensation := saga.Compensation
		compArg := ""
		if compensation != "" {
			compArg = fmt.Sprintf(`, compensationMethod = "%s"`, compensation)
		}

		var stepComments strings.Builder
		for _, step := range saga.Steps {
			stepComments.WriteString(fmt.Sprintf("\t\t// Step → %s.%s()\n", step.Service, step.Method))
		}

		methods.WriteString(fmt.Sprintf(`
	@DistributedSaga(sagaId = "%s"%s, timeout = 30000)
	@org.springframework.transaction.annotation.Transactional
	public void %s() {
		// TODO: implement saga orchestration
%s	}
`, saga.SagaID, compArg, methodName, stepComments.String()))

		if compensation != "" {
			methods.WriteString(fmt.Sprintf(`
	@org.springframework.transaction.annotation.Transactional
	public void %s() {
		// TODO: implement compensation / rollback
	}
`, compensation))
		}
	}

	imports := ""
	if hasSagaImport {
		imports = sagaImports.String() + "\n"
	}

	return fmt.Sprintf(`package %s.%s;

import org.springframework.stereotype.Service;
%s
@Service
public class %sService {

%s
	public %sService(%s) {
%s	}
%s}
`,
		pkg, svcPkg,
		imports,
		prefix,
		fieldDecls.String(),
		prefix, ctorParams.String(),
		ctorAssigns.String(),
		methods.String(),
	)
}
