package generator

import (
	"fmt"

	"github.com/fractalx-org/fractalx-cli/internal/model"
	"github.com/fractalx-org/fractalx-cli/internal/transform"
)

func genRepository(spec *model.ProjectSpec, svc *model.Service, ent *model.Entity) string {
	pkg := transform.ResolvedPackage(spec)
	svcPkg := transform.SvcPackage(svc)
	isMongo := svc.DB == "mongodb"

	var repoIf, importPkg, idType string
	if isMongo {
		repoIf = "MongoRepository"
		importPkg = "org.springframework.data.mongodb.repository.MongoRepository"
		idType = "String"
	} else {
		repoIf = "JpaRepository"
		importPkg = "org.springframework.data.jpa.repository.JpaRepository"
		idType = "Long"
	}

	return fmt.Sprintf(`package %s.%s;

import %s;
import org.springframework.stereotype.Repository;

@Repository
public interface %sRepository extends %s<%s, %s> {

	// Add custom query methods here
}
`, pkg, svcPkg, importPkg, ent.Name, repoIf, ent.Name, idType)
}
