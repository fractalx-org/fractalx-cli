package generator

import (
	"fmt"

	"github.com/fractalx/fractalx-init/internal/model"
	"github.com/fractalx/fractalx-init/internal/transform"
)

func genK8s(spec *model.ProjectSpec, svc *model.Service) string {
	prefix := transform.SvcPrefix(svc)
	labelName := transform.ToSnake(prefix)

	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
        - name: %s
          image: %s/%s:latest
          ports:
            - containerPort: %d
          env:
            - name: SPRING_PROFILES_ACTIVE
              value: "prod"
            - name: FRACTALX_REGISTRY_URL
              value: "http://fractalx-registry:8761"
          readinessProbe:
            httpGet:
              path: /actuator/health
              port: %d
            initialDelaySeconds: 30
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /actuator/health
              port: %d
            initialDelaySeconds: 60
            periodSeconds: 30
---
apiVersion: v1
kind: Service
metadata:
  name: %s
spec:
  selector:
    app: %s
  ports:
    - protocol: TCP
      port: %d
      targetPort: %d
`,
		svc.Name, labelName,
		labelName,
		labelName,
		svc.Name, spec.ArtifactID, svc.Name, svc.Port,
		svc.Port,
		svc.Port,
		svc.Name,
		labelName,
		svc.Port, svc.Port,
	)
}
