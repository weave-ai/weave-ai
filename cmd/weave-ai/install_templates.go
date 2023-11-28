package main

// There are two components currently in the Flamingo project:
// - the LM controller
// - the BRS controller

const installTemplate = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: weave-ai
resources:
- namespace.yaml
- "https://github.com/weave-ai/lm-controller/releases/download/v0.8.0/lm-controller.crds.yaml"
- "https://github.com/weave-ai/lm-controller/releases/download/v0.8.0/lm-controller.rbac.yaml"
- "https://github.com/weave-ai/lm-controller/releases/download/v0.8.0/lm-controller.deployment.yaml"
{{- if .ModelCatalog }}
- "https://github.com/weave-ai/weave-ai/releases/download/{{ .Version }}/model-catalog.yaml"
{{- end }}
`

var namespaceTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`

const defaultClusterSecretTemplate = `
---
`
