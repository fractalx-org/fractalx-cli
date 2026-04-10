package generator

import (
	"fmt"
	"strings"

	"github.com/fractalx/fractalx-init/internal/model"
	"github.com/fractalx/fractalx-init/internal/transform"
)

func genFlyway(spec *model.ProjectSpec) string {
	var b strings.Builder

	for _, svc := range spec.Services {
		if svc.DB == "h2" || svc.DB == "mongodb" || svc.DB == "redis" {
			continue
		}
		if len(svc.Entities) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("-- %s\n", svc.Name))
		for _, ent := range svc.Entities {
			tableName := transform.ToSnake(ent.Name) + "s"
			b.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))
			b.WriteString("    id BIGSERIAL PRIMARY KEY")
			for _, f := range ent.Fields {
				col := transform.ToSnake(f.Name)
				sqlT := transform.SqlType(f.Type)
				b.WriteString(fmt.Sprintf(",\n    %s %s", col, sqlT))
			}
			b.WriteString("\n);\n\n")
		}
	}

	return b.String()
}
