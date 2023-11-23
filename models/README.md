# Models for Weave AI

Weave AI Controllers specialize in facilitating the deployment and management of AI models on Kubernetes using Flux. Our platform offers a curated selection of models, each packaged in the Flux OCI format for seamless and efficient deployment in your GitOps pipelines.

## Available Models

Our growing list of models includes:

- Dragon Yi 6B v0
- Llama-2 7B Chat
- Llama-2 7B Instruct with 32K context
- Mistral 7B Instruct v0.1
- Mistral 7B v0.1
- MistralLite 7B v0.1
- Orca-2 7B
- TinyLlama 1.1B Chat
- YARN Mistral 7B with 128K context
- Zephyr 7B Alpha
- Zephyr 7B Beta

## Getting Started

To start using these models:

1. **Prerequisites**: Ensure you have Kubernetes and Flux installed and configured on your system.
2. **Installation**: Apply the Kustomization manifest to install the catalog of the models to your cluster.
3. **Model Activation**: Choose the model that best fits your needs, change `spec.suspend` to `false` to activate or resume the model reconciliation.
4. **Deployment**: Deploy the activated model using LM-Controller.

## Instructions

Installing Flux. We have to disable network policy in Flux for now to allow the inference engine deployed by the LM Controller to download BLOBs directly from the Source Controller.

```shell
flux install --network-policy=false
```

```shell
kubectl apply -k url
```
