# Install Infinity Rules MCP locally

The Infinity Rules MCP Service is deployed by the Autopilot subchart of the `backingservices` chart and renders only when Autopilot is enabled. This guide installs both on Docker Desktop Kubernetes for local testing. For all settings, see [Infinity Rules MCP Service](../charts/backingservices/charts/autopilot/README.md#infinity-rules-mcp-service) in the Autopilot chart README.

## Prerequisites

- Start Docker Desktop and enable its Kubernetes cluster.
- Use pullable Autopilot Service and `infinity-rules-skills-mcp` images. Bearer mode requires **infinity-rules-mcp 0.7.17 or later**; ask the image publisher which MCP version is bundled.
- Identify an allowed Pega host name, such as `pega.example.com`.
- If the registry is private, create an image-pull Secret in the `autopilot` namespace before installing.

The commands below explicitly use the `docker-desktop` context so that they do not target your current context by accident. Confirm that the local cluster is ready:

```sh
kubectl --context docker-desktop get nodes
```

## Install

From the root of this repository, load the chart dependencies, then replace the example images:

```sh
helm dependency update charts/backingservices

export AUTOPILOT_IMAGE='registry.example.com/autopilot-service@sha256:<digest>'
export MCP_IMAGE='registry.example.com/infinity-rules-skills-mcp@sha256:<digest>'

helm --kube-context docker-desktop upgrade --install backingservices ./charts/backingservices \
  --namespace autopilot --create-namespace --wait --timeout 5m \
  --set srs.enabled=false \
  --set srs.srsStorage.provisionInternalESCluster=false \
  --set autopilot.enabled=true \
  --set autopilot.replicas=1 \
  --set-string "autopilot.docker.autopilot.image=$AUTOPILOT_IMAGE" \
  --set-string "autopilot.docker.infinityRulesMcp.image=$MCP_IMAGE" \
  --set autopilot.infinityRulesMcp.enabled=true \
  --set autopilot.infinityRulesMcp.replicas=1
```

For a private registry, add the existing Secret name to the Helm command, for example:

```sh
--set-string 'autopilot.docker.imagePullSecretNames[0]=registry-credentials'
```

Do not put registry credentials in Helm values or command-line arguments.

This installs MCP in trusted-header mode, which **does not authenticate callers or verify Pega TLS**. Use it only in a trusted local environment.

## Enable bearer mode

Add the bearer settings to the same command and run it again:

```sh
  --set autopilot.infinityRulesMcp.auth.enabled=true \
  --set-string autopilot.infinityRulesMcp.auth.publicUrl=https://mcp.localhost \
  --set-string 'autopilot.infinityRulesMcp.allowedHosts[0]=pega.example.com'
```

Bearer mode also requires an Autopilot release that sends bearer tokens to MCP. To check that MCP now requires a token, forward its port and send an anonymous request:

```sh
kubectl --context docker-desktop -n autopilot port-forward svc/infinity-rules-mcp-service 8090:80
curl -i -X POST http://localhost:8090/mcp -H 'X-Pega-Host: https://pega.example.com'
```

The response is `401` with a `WWW-Authenticate` header that contains `resource_metadata`. External MCP clients need the Ingress described in the chart README.

## Remove

```sh
helm --kube-context docker-desktop uninstall backingservices --namespace autopilot
```
