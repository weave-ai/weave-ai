package main

// There are two components currently in the Flamingo project:
// - the LM controller
// - the BRS controller

const installTemplate = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- namespace.yaml
- "https://github.com/weave-ai/lm-controller/releases/download/v0.9.0/lm-controller.crds.yaml"
- "https://github.com/weave-ai/lm-controller/releases/download/v0.9.0/lm-controller.rbac.yaml"
- "https://github.com/weave-ai/lm-controller/releases/download/v0.9.0/lm-controller.deployment.yaml"
{{- if .WithModelCatalog }}
- "https://github.com/weave-ai/weave-ai/releases/download/v{{ .Version }}/model-catalog.yaml"
{{- end }}
{{- if .WithDefaultTenant }}
- default_tenant.yaml
{{- end }}
`

var namespaceTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`

var defaultTenantTemplate = `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: lm-tenant-role
  namespace: %s
rules:
  - apiGroups: ["apps"]
    resources: ["deployments", "replicasets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["services", "persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["serving.knative.dev"]
    resources: ["services"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: lm-tenant-role-binding
  namespace: %s
subjects:
  - kind: ServiceAccount
    name: default
    namespace: %s
roleRef:
  kind: Role
  name: lm-tenant-role
  apiGroup: rbac.authorization.k8s.io
`

const defaultClusterSecretTemplate = `
---
`
