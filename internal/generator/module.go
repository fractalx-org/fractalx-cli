package generator

import (
	"fmt"

	"github.com/fractalx-org/fractalx-cli/internal/model"
	"github.com/fractalx-org/fractalx-cli/internal/transform"
)

func genModuleMarker(spec *model.ProjectSpec, svc *model.Service) string {
	pkg := transform.ResolvedPackage(spec)
	svcPkg := transform.SvcPackage(svc)
	prefix := transform.SvcPrefix(svc)
	schema := transform.ResolvedSchema(svc)

	return fmt.Sprintf(`package %s.%s;

import org.fractalx.annotations.DecomposableModule;

@DecomposableModule(
    serviceName = "%s",
    port = %d,
    ownedSchemas = {"%s"},
    independentDeployment = true
)
public class %sModule {
}
`, pkg, svcPkg, svc.Name, svc.Port, schema, prefix)
}
