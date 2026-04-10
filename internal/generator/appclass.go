package generator

import (
	"fmt"

	"github.com/fractalx/fractalx-cli/internal/model"
	"github.com/fractalx/fractalx-cli/internal/transform"
)

func genAppClass(spec *model.ProjectSpec) string {
	pkg := transform.ResolvedPackage(spec)
	cls := transform.AppClassName(spec)
	hasSagas := len(spec.Sagas) > 0

	schedulingImport := ""
	schedulingAnnotation := ""
	if hasSagas {
		schedulingImport = "\nimport org.springframework.scheduling.annotation.EnableScheduling;"
		schedulingAnnotation = "\n@EnableScheduling"
	}

	return fmt.Sprintf(`package %s;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;%s

@SpringBootApplication%s
public class %s {

	public static void main(String[] args) {
		SpringApplication.run(%s.class, args);
	}
}
`, pkg, schedulingImport, schedulingAnnotation, cls, cls)
}

func genAppTest(spec *model.ProjectSpec) string {
	pkg := transform.ResolvedPackage(spec)
	cls := transform.AppClassName(spec)

	return fmt.Sprintf(`package %s;

import org.junit.jupiter.api.Test;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.test.context.ActiveProfiles;

@SpringBootTest(webEnvironment = WebEnvironment.NONE, properties = {"spring.main.lazy-initialization=true"})
@ActiveProfiles("test")
class %sTests {

	@Test
	void contextLoads() {
	}
}
`, pkg, cls)
}
