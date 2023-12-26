# Weave AI

Weave AI is a collection of Flux controllers and CLI that manage the 
lifecycle of Large Language Models (LLMs) on Kubernetes.

**Weave AI CLI** aims to be the easiest way to onboard LLMs on Kubernetes, 
and the **Weave AI controllers** manage the lifecycle of the LLMs, including
training, serving, and monitoring of the models on production Kubernetes
clusters.

## Getting started with Weave AI CLI

### CLI installation

#### Install with CURL
```shell
# CURL
curl -s https://raw.githubusercontent.com/weave-ai/weave-ai/main/install/weave-ai.sh | sudo bash
````

#### Install with Homebrew
```shell
# or Homebrew
brew install weave-ai/tap/weave-ai
```

### Quick Start

Prerequisites: Please install Kubernetes v1.27+ and Flux v2.1.0+ before proceeding.
Minimum requirements of the Kubernetes cluster are 8 CPUs and 16GB of memory with 100GB of SSD storage.

```shell
flux install --network-policy=false

weave-ai install
weave-ai run -p zephyr-7b-beta
```

### What's next?

To learn more about Weave AI, a step by step guide can be found [here](docs/GETTING_STARTED.md).
