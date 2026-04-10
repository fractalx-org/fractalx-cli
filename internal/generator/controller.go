package generator

import (
	"fmt"
	"strings"

	"github.com/fractalx/fractalx-cli/internal/model"
	"github.com/fractalx/fractalx-cli/internal/transform"
)

func genController(spec *model.ProjectSpec, svc *model.Service, ent *model.Entity) string {
	pkg := transform.ResolvedPackage(spec)
	svcPkg := transform.SvcPackage(svc)
	prefix := transform.SvcPrefix(svc)
	isMongo := svc.DB == "mongodb"

	svcClass := prefix + "Service"
	svcVar := strings.ToLower(prefix[:1]) + prefix[1:] + "Service"
	varName := strings.ToLower(ent.Name[:1]) + ent.Name[1:]
	idType := "Long"
	if isMongo {
		idType = "String"
	}

	return fmt.Sprintf(`package %s.%s;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import jakarta.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/%ss")
public class %sController {

	private final %s %s;

	public %sController(%s %s) {
		this.%s = %s;
	}

	@GetMapping
	public ResponseEntity<List<%s>> getAll() {
		return ResponseEntity.ok(%s.findAll%ss());
	}

	@GetMapping("/{id}")
	public ResponseEntity<%s> getById(@PathVariable %s id) {
		return ResponseEntity.ok(%s.find%sById(id));
	}

	@PostMapping
	public ResponseEntity<%s> create(@Valid @RequestBody %s %s) {
		return ResponseEntity.status(201).body(%s.create%s(%s));
	}

	@DeleteMapping("/{id}")
	public ResponseEntity<Void> delete(@PathVariable %s id) {
		%s.delete%s(id);
		return ResponseEntity.noContent().build();
	}
}
`,
		pkg, svcPkg,
		varName, ent.Name,
		svcClass, svcVar,
		ent.Name, svcClass, svcVar,
		svcVar, svcVar,
		ent.Name, svcVar, ent.Name,
		ent.Name, idType, svcVar, ent.Name,
		ent.Name, ent.Name, varName, svcVar, ent.Name, varName,
		idType, svcVar, ent.Name,
	)
}
