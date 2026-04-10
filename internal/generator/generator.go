package generator

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fractalx/fractalx-cli/internal/model"
	"github.com/fractalx/fractalx-cli/internal/transform"
)

// GenerateZip generates the project as a ZIP file at the given output path.
// outputPath should be the desired .zip file path.
func GenerateZip(spec *model.ProjectSpec, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create zip file: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	return writeFiles(spec, zw)
}

// GenerateDir writes the project directly to a directory (--no-zip mode).
func GenerateDir(spec *model.ProjectSpec, dir string) error {
	files, err := buildFileMap(spec)
	if err != nil {
		return err
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func writeFiles(spec *model.ProjectSpec, zw *zip.Writer) error {
	files, err := buildFileMap(spec)
	if err != nil {
		return err
	}
	for path, content := range files {
		w, err := zw.Create(path)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, content); err != nil {
			return err
		}
	}
	return nil
}

// buildFileMap returns all generated files as a map of zip-path → content.
func buildFileMap(spec *model.ProjectSpec) (map[string]string, error) {
	root := spec.ArtifactID
	pkg := transform.ResolvedPackage(spec)
	pkgPath := transform.PackagePath(pkg)
	appCls := transform.AppClassName(spec)

	mainJava := root + "/src/main/java/" + pkgPath
	mainRes := root + "/src/main/resources"
	testJava := root + "/src/test/java/" + pkgPath
	testRes := root + "/src/test/resources"

	files := map[string]string{}

	// Root files
	files[root+"/pom.xml"] = genPom(spec)
	files[root+"/fractalx.yaml"] = genSpecYaml(spec)
	files[root+"/README.md"] = genReadme(spec)

	// Application class + test
	files[mainJava+"/"+appCls+".java"] = genAppClass(spec)
	files[testJava+"/"+appCls+"Tests.java"] = genAppTest(spec)

	// YAML configs
	files[mainRes+"/application.yml"] = genAppYml(spec)
	files[mainRes+"/application-dev.yml"] = genDevYml()
	files[mainRes+"/fractalx-config.yml"] = genFractalxConfig(spec)
	files[testRes+"/application.yml"] = genTestYml(spec)

	// Flyway
	var hasFly bool
	for _, svc := range spec.Services {
		if svc.DB == "postgresql" || svc.DB == "mysql" {
			hasFly = true
			break
		}
	}
	if hasFly {
		sql := genFlyway(spec)
		if strings.TrimSpace(sql) != "" {
			files[mainRes+"/db/migration/V1__init.sql"] = sql
		}
	}

	// Per-service files
	for i := range spec.Services {
		svc := &spec.Services[i]
		svcPkg := transform.SvcPackage(svc)
		svcJava := mainJava + "/" + svcPkg

		files[svcJava+"/"+transform.SvcPrefix(svc)+"Module.java"] = genModuleMarker(spec, svc)
		files[svcJava+"/"+transform.SvcPrefix(svc)+"Service.java"] = genServiceClass(spec, svc)

		for j := range svc.Entities {
			ent := &svc.Entities[j]
			files[svcJava+"/"+ent.Name+".java"] = genEntity(spec, svc, ent)
			files[svcJava+"/"+ent.Name+"Repository.java"] = genRepository(spec, svc, ent)
			files[svcJava+"/"+ent.Name+"Controller.java"] = genController(spec, svc, ent)
		}

		if spec.Infra.Kubernetes {
			files[root+"/k8s/"+svc.Name+"-deployment.yml"] = genK8s(spec, svc)
		}
	}

	// Optional infra files
	if spec.Infra.Docker {
		files[root+"/docker-compose.dev.yml"] = genDocker(spec)
	}
	if spec.Infra.CI {
		files[root+"/.github/workflows/ci.yml"] = genCI(spec)
	}

	return files, nil
}
