# Weave AI

Weave AI is a collection of Flux controllers that manage the lifecycle of Large Language Models (LLMs) on Kubernetes.

## Getting started with Weave AI CLI

### CLI installation

```shell
# CURL
curl -s https://raw.githubusercontent.com/weave-ai/weave-ai/main/install/weave-ai.sh | sudo bash

# or Homebrew
brew install weave-ai/tap/weave-ai
```

### Quick Start

Prerequisites: Please install Kubernetes v1.27+ and Flux v2.1.0+ before proceeding.

```shell
flux install --network-policy=false

weave-ai install
weave-ai run -p zephyr-7b-beta
```
