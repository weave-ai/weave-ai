# Running Your First LLM with Weave AI: A Step-by-Step Guide

## Step 1: Install Weave AI CLI

First, install the Weave AI Command Line Interface (CLI). Use the following command in your terminal:

```shell
curl -s https://raw.githubusercontent.com/weave-ai/weave-ai/main/install/weave-ai.sh | sudo bash
```

You should see information about the download and installation process, confirming successful completion.

```shell
[INFO]  Downloading metadata https://api.github.com/repos/weave-ai/weave-ai/releases/latest
[INFO]  Using 0.11.0 as release
[INFO]  Downloading hash https://github.com/weave-ai/weave-ai/releases/download/v0.11.0/weave-ai_0.11.0_checksums.txt
[INFO]  Downloading binary https://github.com/weave-ai/weave-ai/releases/download/v0.11.0/weave-ai_0.11.0_linux_amd64.tar.gz
[INFO]  Verifying binary download
[INFO]  Installing weave-ai to /usr/local/bin/weave-ai
```

## Step 2: Create a KIND Cluster

Next, set up a Kubernetes in Docker (KIND) cluster with this command:

```shell
kind create cluster
```

The output will confirm the creation of your cluster, detailing the steps like node image preparation, nodes configuration, and control-plane initiation.

```shell
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.27.3) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Have a question, bug, or feature request? Let us know! https://kind.sigs.k8s.io/#community 🙂
```

## Step 3: Install Flux

Now, install Flux to manage your cluster resources. Note: You need to disable the default network policy:

```shell
flux install --network-policy=false
```

The output will show the successful installation of various components in the flux-system namespace.

```shell
✚ generating manifests
✔ manifests build completed
► installing components in flux-system namespace
CustomResourceDefinition/alerts.notification.toolkit.fluxcd.io created
CustomResourceDefinition/buckets.source.toolkit.fluxcd.io created
CustomResourceDefinition/gitrepositories.source.toolkit.fluxcd.io created
CustomResourceDefinition/helmcharts.source.toolkit.fluxcd.io created
CustomResourceDefinition/helmreleases.helm.toolkit.fluxcd.io created
CustomResourceDefinition/helmrepositories.source.toolkit.fluxcd.io created
CustomResourceDefinition/kustomizations.kustomize.toolkit.fluxcd.io created
CustomResourceDefinition/ocirepositories.source.toolkit.fluxcd.io created
CustomResourceDefinition/providers.notification.toolkit.fluxcd.io created
CustomResourceDefinition/receivers.notification.toolkit.fluxcd.io created
Namespace/flux-system created
ResourceQuota/flux-system/critical-pods-flux-system created
ServiceAccount/flux-system/helm-controller created
ServiceAccount/flux-system/kustomize-controller created
ServiceAccount/flux-system/notification-controller created
ServiceAccount/flux-system/source-controller created
ClusterRole/crd-controller-flux-system created
ClusterRole/flux-edit-flux-system created
ClusterRole/flux-view-flux-system created
ClusterRoleBinding/cluster-reconciler-flux-system created
ClusterRoleBinding/crd-controller-flux-system created
Service/flux-system/notification-controller created
Service/flux-system/source-controller created
Service/flux-system/webhook-receiver created
Deployment/flux-system/helm-controller created
Deployment/flux-system/kustomize-controller created
Deployment/flux-system/notification-controller created
Deployment/flux-system/source-controller created
◎ verifying installation
✔ helm-controller: deployment ready
✔ kustomize-controller: deployment ready
✔ notification-controller: deployment ready
✔ source-controller: deployment ready
✔ install finished
```

## Step 4: Install Weave AI in Your Cluster

At this stage, you're ready to install Weave AI and its controller(s):

```shell
weave-ai install
```

After installation, you'll see confirmation messages, indicating that the Weave AI controllers and the default model catalog are set up.

```shell
✚ generating manifests
✔ manifests build completed
► installing components in weave-ai namespace
CustomResourceDefinition/languagemodels.ai.contrib.fluxcd.io created
Namespace/weave-ai created
Role/default/lm-tenant-role created
Role/weave-ai/lm-leader-election-role created
ClusterRole/lm-manager-role created
RoleBinding/default/lm-tenant-role-binding created
RoleBinding/weave-ai/lm-leader-election-rolebinding created
ClusterRoleBinding/lm-cluster-reconciler created
ClusterRoleBinding/lm-manager-rolebinding created
Deployment/weave-ai/lm-controller created
◎ verifying installation
✔ lm-controller: deployment ready
✔ install finished
```

## Step 5: Listing Models

To view the available models in your cluster, use:

```shell
weave-ai models
```

This command lists all OCI models, which are initially in an `INACTIVE` state to conserve resources.


```shell
NAME                               VERSION           FAMILY  STATUS    CREATED
weave-ai/dragon-yi-6b              v0.0.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/llama-2-7b-chat           v1.0.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/llama-2-7b-instruct-32k   v1.0.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/llamaguard-7b             v0.1.0-q4km-gguf          INACTIVE  1 minute ago
weave-ai/mistral-7b-instruct-v0.1  v0.1.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/mistral-7b-v0.1           v0.1.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/mistrallite-7b            v1.0.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/orca-2-7b                 v1.0.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/stablelm-zephyr-3b        v0.1.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/tinyllama-1.1b-chat       v0.3.0-q3ks-gguf          INACTIVE  1 minute ago
weave-ai/yarn-mistral-7b-128k      v0.1.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/zephyr-7b-alpha           v1.0.0-q5km-gguf          INACTIVE  1 minute ago
weave-ai/zephyr-7b-beta            v1.0.0-q5km-gguf          INACTIVE  1 minute ago
```

## 6. Create and run a model instance.

Step 6: Create and Run a Model Instance
To activate and run a model, use the following command:

```shell
weave-ai run -d --ui --name my-model weave-ai/zephyr-7b-beta
```

This command activates the model and sets up a UI for interaction. Follow the instructions for port forwarding to access the LLM and the UI.

```shell
► checking if model weave-ai/zephyr-7b-beta exists and is active
► activate model weave-ai/zephyr-7b-beta
◎ waiting for model weave-ai/zephyr-7b-beta to be active
► creating new LLM instance default/my-model
◎ waiting for default/my-model to be ready
◎ waiting for default/my-model-chat-app to be ready
✔ to connect to your LLM:
  kubectl port-forward -n default svc/my-model 8000:8000
✔ to connect to the UI:
  kubectl port-forward -n default deploy/my-model-chat-app 8501:8501
```

Simply run the UI port-forward command.

```shell
kubectl port-forward -n default deploy/my-model-chat-app 8501:8501
```

Then open your browser to `http://localhost:8501` to try the model via our quick chat app.

![Chat UI](https://github.com/weave-ai/weave-ai/assets/10666/ff6e624e-90d5-42d9-9197-245619b1c4fa)

## Delete the LM instance.

Finally, to remove the LM instance and the associated UI, execute:

```shell
kubectl delete lm/my-model
```

This command deletes the specified language model instance from your cluster, also with the default chat UI if you've created one.
```
languagemodel.ai.contrib.fluxcd.io "my-model" deleted
```
