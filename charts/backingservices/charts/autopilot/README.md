# Autopilot Service Helm chart

The Pega `Autopilot Service` backing service provides GenAI capabilities for Pega Platform in client-managed cloud and on-premises deployments. This chart deploys the service as an external backing service that connects directly to supported LLM providers without requiring the Pega GenAI Hub gateway. The chart also supports optional OAuth authentication between Pega Platform and the Autopilot Service, a deployment-scoped model list through a bundled or external ConfigMap, and custom OpenAI-compatible providers.

## Configuring a backing service with your Pega environment

You can deploy the Autopilot Service in the `pega` namespace or in a separate namespace. After deployment, configure the service endpoint in Pega Platform so that the platform can route Autopilot requests to the backing service.

## Pega GenAI feature support matrix

| Pega Platform version | GenAI Connect | GenAI Coach | GenAI Agent |
|---|:---:|:---:|:---:|
| 24.2 | Yes | No | No |
| 24.2.4 | Yes | Yes | No |
| 25.1.1 (`HFIX-C4307` required) | Yes | Yes | Yes |
| 25.1.2 | Yes | Yes | Yes |

## Supported LLM providers

| Provider | Authentication Methods |
|---|---|
| Azure OpenAI | API Key, Existing Secret |
| AWS Bedrock | Access Key/Secret, Existing Secret |
| Google Vertex AI | Service Account JSON (base64-encoded), Existing Secret |
| Custom OpenAI-compatible | API Key (inline or existing Secret) |

## Autopilot Service configuration

The `values.yaml` file defines the deployment, networking, authentication, provider credentials, and model-list settings for the Autopilot Service.

### Configuration settings

| Configuration | Usage |
|---|---|
| `enabled` | Set this parameter to `true` to deploy the Autopilot Service as a backing service. |
| `deployment.name` | Specify the deployment name. The chart prefixes Autopilot resources with this value. Default: `autopilot` |
| `docker.registry.url` | Specify the image registry URL. |
| `docker.registry.username` | Specify the image registry username. |
| `docker.registry.password` | Specify the image registry password. |
| `docker.imagePullSecretNames` | Specify an array of existing image pull Secrets to use when pulling images. |
| `docker.autopilot.image` | Specify the Autopilot Service image and tag. |
| `docker.autopilot.imagePullPolicy` | Specify the Autopilot Service image pull policy. Default: `IfNotPresent` |
| `docker.infinityRulesMcp.image` | Specify the Infinity Rules MCP Service image and tag. Required when `infinityRulesMcp.enabled` is `true`. Bearer mode requires MCP 0.7.17 or later. |
| `docker.infinityRulesMcp.imagePullPolicy` | Specify the Infinity Rules MCP Service image pull policy. Default: `IfNotPresent` |
| `replicas` | Specify the number of Autopilot pods. Default: `2` |
| `resources` | Define CPU and memory requests and limits for the Autopilot container. |
| `service.port` | Specify the Kubernetes Service port. Default: `80` |
| `service.targetPort` | Specify the container port exposed by the service. Default: `8080` |
| `service.serviceType` | Specify the Kubernetes Service type. Default: `ClusterIP` |
| `enableGenaiHub` | Leave this parameter as `false` for client-managed cloud or on-premises deployments that connect directly to providers. Set it to `true` only when requests must be routed through the Pega GenAI Hub gateway. |
| `authEnabled` | Set this parameter to `true` to require OAuth bearer tokens from Pega Platform. Default: `false` |
| `oauthPublicKeyURL` | Specify the identity provider public key endpoint used to validate incoming bearer tokens. This value is required when `authEnabled` is `true`. |
| `isInternalDeployment` | Leave this parameter as `false` for client-managed cloud or on-premises deployments. |
| `modelProviders` | Specify a comma-separated list of providers to expose, such as `"Azure,Vertex,Bedrock"`. If this value is empty, the service relies on the configured credentials and model list. For FedRAMP or PCFG environments, set `"Bedrock"` as the provider. |
| `awsRegion` | Specify the AWS region used by Bedrock. Default: `us-east-1` |
| `vcs.gatewayDefaultModel` | Default model used by the AI Assistant. |
| `vcs.githubCopilotBaseUrl` | GitHub Copilot models base URL used by the AI Assistant. |
| `vcs.openaiBaseUrl` | OpenAI base URL used by the AI Assistant. |
| `vcs.anthropicBaseUrl` | Anthropic base URL used by the AI Assistant. |
| `vcs.infinityRulesMcpUrl` | Infinity Rules MCP URL used by the AI Assistant. |
| `infinityRulesMcp.enabled` | Set this parameter to `true` to deploy the optional Infinity Rules MCP Service in the same namespace as Autopilot. Default: `false` |
| `infinityRulesMcp.replicas` | Specify the number of Infinity Rules MCP Service pods. Default: `2` |
| `infinityRulesMcp.service.port` | Specify the Infinity Rules MCP Kubernetes Service port. Default: `80` |
| `infinityRulesMcp.service.targetPort` | Specify the Infinity Rules MCP container port. Default: `8090` |
| `infinityRulesMcp.service.serviceType` | Specify the Infinity Rules MCP Kubernetes Service type. Default: `ClusterIP` |
| `infinityRulesMcp.resources` | Define CPU and memory requests and limits for the Infinity Rules MCP container. |
| `infinityRulesMcp.auth`, `infinityRulesMcp.allowedHosts`, `infinityRulesMcp.ingress`, `infinityRulesMcp.trustStore` | Configure bearer authentication and external access for the Infinity Rules MCP Service. See [Infinity Rules MCP Service](#infinity-rules-mcp-service). |
| `autopilot.proxy.httpProxy` | Optional `HTTP_PROXY` value for the Autopilot container. |
| `autopilot.proxy.httpsProxy` | Optional `HTTPS_PROXY` value for the Autopilot container. |
| `autopilot.proxy.noProxy` | Optional comma-separated `NO_PROXY` value for the Autopilot container. |
| `infinityRulesMcp.proxy.httpProxyHost` | Optional `HTTP_PROXY_HOST` value for the Infinity Rules MCP container. |
| `infinityRulesMcp.proxy.httpProxyPort` | Optional `HTTP_PROXY_PORT` value for the Infinity Rules MCP container. |
| `infinityRulesMcp.proxy.httpProxyUsername` | Optional `HTTP_PROXY_USERNAME` value for the Infinity Rules MCP container. |
| `infinityRulesMcp.proxy.httpProxyPassword` | Optional `HTTP_PROXY_PASSWORD` value for the Infinity Rules MCP container. |
| `infinityRulesMcp.proxy.noProxy` | Optional pipe-separated `NO_PROXY` value for the Infinity Rules MCP container. |
| `affinity` | Define pod affinity rules. |
| `tolerations` | Define pod tolerations. |

### AI Assistant settings

The `vcs` settings configure the AI Assistant environment variables in the
Autopilot container. The values keys use Helm-friendly names, while the
rendered environment variable names remain unchanged.

```yaml
vcs:
  gatewayDefaultModel: "bedrock/anthropic/claude-sonnet-4-6/v1"
  githubCopilotBaseUrl: "https://api.githubcopilot.com"
  openaiBaseUrl: "https://api.openai.com/v1"
  anthropicBaseUrl: "https://api.anthropic.com/v1"
  infinityRulesMcpUrl: "http://infinity-rules-mcp-service/mcp"
```

These values render as `VCS_GATEWAY_DEFAULT_MODEL`,
`VCS_GITHUB_MODELS_BASE_URL`, `VCS_OPENAI_BASE_URL`,
`VCS_ANTHROPIC_BASE_URL`, and `VCS_INFINITY_RULES_MCP_URL`.

### Proxy configuration

Proxy environment variables are optional and are configured independently for
the Autopilot and Infinity Rules MCP containers. Supply proxy credentials through
your deployment-specific values file rather than the chart defaults.

```yaml
autopilot:
  proxy:
    httpProxy: "http://<username>:<password>@<proxy-host>:<proxy-port>"
    httpsProxy: "http://<username>:<password>@<proxy-host>:<proxy-port>"
    noProxy: "localhost,127.0.0.1,::1,.svc,.cluster.local"

infinityRulesMcp:
  proxy:
    httpProxyHost: "<proxy-host>"
    httpProxyPort: "<proxy-port>"
    httpProxyUsername: "<username>"
    httpProxyPassword: "<password>"
    noProxy: "localhost|127.0.0.1|::1|.svc|.cluster.local"
```

The Autopilot service requires comma-separated `NO_PROXY` entries, while the
Infinity Rules MCP service requires pipe-separated entries.

Both service containers are configured to run without root privileges by
default. The Autopilot container uses the non-root user defined by its image,
while the Infinity Rules MCP container uses UID/GID `1001`, matching its
Dockerfile. Both containers also drop Linux capabilities and disable privilege
escalation. The root filesystem remains writable by default because Autopilot
uses `/tmp/autopilot/models-cache` and `/lean/prometheus/metrics` at runtime.

The security settings can be overridden independently when required:

```yaml
securityContext:
  runAsNonRoot: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  seccompProfile:
    type: RuntimeDefault

infinityRulesMcp:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    runAsGroup: 1001
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - ALL
    seccompProfile:
      type: RuntimeDefault
```

For Autopilot, proxy credentials are embedded in the `httpProxy` and
`httpsProxy` URLs when authentication is required. For example:

```yaml
autopilot:
  proxy:
    httpProxy: "http://<username>:<password>@<proxy-host>:<proxy-port>"
    httpsProxy: "http://<username>:<password>@<proxy-host>:<proxy-port>"
```

## Infinity Rules MCP Service

The chart can optionally deploy the Infinity Rules MCP Service (`infinity-rules-mcp-service`) in the same namespace as the Autopilot Service. Autopilot calls it at `vcs.infinityRulesMcpUrl`, which defaults to the internal endpoint `http://infinity-rules-mcp-service/mcp`. If `infinityRulesMcp.enabled` is `false`, the MCP Deployment, Service, PodDisruptionBudget, and Ingress are not rendered. MCP is rendered only when the Autopilot Service is also enabled (`enabled: true`).

```yaml
autopilot:
  enabled: true

  docker:
    infinityRulesMcp:
      image: "YOUR_REGISTRY_URL/infinity-rules-skills-mcp:YOUR_TAG"
      imagePullPolicy: IfNotPresent

  infinityRulesMcp:
    enabled: true
```

### Authentication modes

`infinityRulesMcp.auth.enabled` selects how MCP authenticates its callers:

| `auth.enabled` | Behavior |
|---|---|
| `false` (default) | **Trusted-header mode (deprecated).** MCP does not authenticate callers and does not verify the TLS certificates of Pega. The MCP resources render exactly as in earlier chart releases, so upgrading the chart does not restart MCP. |
| `true` | **Bearer mode.** Every caller, including Autopilot, sends `Authorization: Bearer <token>`. MCP validates the token before it calls Pega, always verifies TLS, and accepts only Pega hosts on `allowedHosts`. External AI and MCP clients, such as VS Code, can use the same deployment through an optional Ingress. |

> **Warning:** Do not expose trusted-header mode outside the cluster. The chart rejects `infinityRulesMcp.ingress.enabled: true` in this mode. Do not set `infinityRulesMcp.service.serviceType` to `NodePort` or `LoadBalancer` unless bearer mode is enabled.

In bearer mode, each installation selects one token mode with `auth.tokenMode`:

| Token mode | External clients present | MCP calls Pega with |
|---|---|---|
| `passthrough` (default) | A Pega access token | The same token |
| `exchange` | An identity provider (IdP) token issued for MCP | A Pega token that MCP obtains with the JWT bearer grant. MCP caches it in memory for at most 30 minutes and never returns it to the client. |

Autopilot sends its own Pega access token in both token modes. Its OAuth client, `26383066777993846573`, is in the default `auth.passthrough.audiences` and `auth.exchange.internalAudiences` lists, so keep it in whichever list you override.

### Bearer mode prerequisites

| Component | Requirement |
|---|---|
| MCP image | `infinity-rules-skills-mcp` with **infinity-rules-mcp 0.7.17 or later**. Bearer authentication and the health endpoints need 0.7.13, exchange-mode internal audiences need 0.7.14, and the context-path headers that Autopilot sends need 0.7.17. |
| Autopilot | An Autopilot Service release that sends `Authorization: Bearer <access token>` and a bare-origin `X-Pega-Host` header to MCP. Earlier releases send only `X-Pega-Authorization`, which bearer mode rejects with `401`. |
| Pega: Autopilot client | The OAuth 2.0 client `26383066777993846573`, which Autopilot uses for its JWT assertion token exchange, allows only that grant. Only Autopilot, which holds the signing key, can then obtain tokens for it. |
| Pega: OAuth endpoints | OIDC discovery, JWKS, and token endpoints are reachable from the MCP pods over HTTPS. Pega tokens carry `iss = urn:<pega-host>` and are signed with RS256. |
| Pega: TLS | The Pega certificate chain is trusted by a standard JVM, or you configure `infinityRulesMcp.trustStore`. Trusted-header mode does not verify TLS, so this can be a new requirement. |

Every caller sends `X-Pega-Host` as a bare HTTPS origin, such as `https://pega.example.com`, on every request. Hosts that use `http://` or include a path such as `/prweb` are rejected. Callers can select an application with `X-Pega-App-Alias`.

Additional prerequisites depend on the token mode:

| | `passthrough` | `exchange` |
|---|---|---|
| Identity provider | None | An application that represents MCP, whose tokens carry the MCP audience, and a client registration for each supported MCP client with its redirect URIs. Issuer discovery (or the pinned `jwksUri`) must be reachable from the MCP pods over HTTPS. |
| Pega client for external users | A public OAuth 2.0 client registration with the authorization code grant, PKCE, the redirect URIs of each MCP client, and a least-privilege access group. Add its client ID to `auth.passthrough.audiences`. | An OAuth 2.0 client registration with the JWT bearer grant (`urn:ietf:params:oauth:grant-type:jwt-bearer`) that trusts the IdP and maps an assertion claim to a Pega operator. It can be confidential, with its secret in a Kubernetes Secret, or public. Set its client ID in `auth.exchange.clientId`. |
| Autopilot token accepted through | `auth.passthrough.audiences` | `auth.exchange.internalAudiences` |
| Users sign in at | Pega, which can delegate to corporate single sign-on | The IdP |
| MCP client configuration | MCP URL, `X-Pega-Host`, Pega client ID | MCP URL, `X-Pega-Host`, IdP client ID |
| Additional MCP egress | None | IdP discovery and JWKS endpoints |

### Infinity Rules MCP settings

All settings are under `infinityRulesMcp`. Except for `ingress.enabled`, the settings below take effect only when `auth.enabled` is `true`.

| Configuration | Usage |
|---|---|
| `auth.enabled` | Set to `true` to enable bearer mode. Default: `false` |
| `auth.publicUrl` | Required. External HTTPS URL of MCP, host root only, for example `https://mcp.example.com`. MCP uses it in its OAuth metadata and `401` challenges, and the Ingress uses its host name. |
| `auth.tokenMode` | `passthrough` or `exchange`. Default: `passthrough` |
| `auth.passthrough.audiences` | Pega OAuth client IDs whose tokens are accepted, in addition to the built-in default client of MCP. Default: `["26383066777993846573"]` (Autopilot) |
| `auth.exchange.issuer` | Required in exchange mode. IdP issuer, exactly as it appears in tokens. |
| `auth.exchange.jwksUri` | Optional pinned HTTPS JWKS URL. Setting it disables issuer discovery. |
| `auth.exchange.audience` | Expected token audience. Default: `<publicUrl>/mcp` |
| `auth.exchange.algorithms` | Accepted signing algorithms (`RS`, `PS`, or `ES` with `256`, `384`, or `512`). Default: `["RS256"]` |
| `auth.exchange.clientId` | Required in exchange mode. Pega OAuth client registration that allows the JWT bearer grant. |
| `auth.exchange.clientAuthMethod` | `client_secret_post` or `client_secret_basic`. Default: `client_secret_post` |
| `auth.exchange.clientSecret.name`, `auth.exchange.clientSecret.key` | Existing Secret and key that hold the client secret of the exchange client. Leave `name` empty for a public client registration. Default key: `client-secret` |
| `auth.exchange.internalAudiences` | Pega client IDs whose Pega-issued tokens are still accepted in exchange mode. Never list a client that users can obtain tokens for. Default: `["26383066777993846573"]` (Autopilot) |
| `allowedHosts` | Required. Pega host names that callers can target with `X-Pega-Host`: exact names such as `pega.example.com` or wildcards such as `*.example.com`, with no scheme, port, or path. |
| `infinityVersion` | Default skills version when a request has no `X-Pega-Infinity-Version` header, for example `26-1`. |
| `ingress.enabled` | Set to `true` to publish MCP for external clients. Allowed only in bearer mode. Default: `false` |
| `ingress.ingressClassName` | Ingress class name. |
| `ingress.annotations` | Annotations for the Ingress, such as controller timeouts and rate limits. |
| `ingress.tls.secretName` | Secret with the TLS certificate for the `auth.publicUrl` host. |
| `trustStore.secretName`, `trustStore.key` | Existing Secret and key that hold a trust store for Pega or IdP hosts signed by a private certificate authority. Default key: `truststore.p12` |
| `trustStore.type` | `PKCS12` or `JKS`. Default: `PKCS12` |
| `trustStore.passwordKey` | Optional key in the same Secret that holds the trust store password. |
| `knowledge.enabled` | Set to `true` to enable the tell-me documentation tool, which sends questions to a Pega-hosted knowledge service. Default: `false` |
| `extraEnv` | Additional non-sensitive environment variables, as `name` and `value` entries. Variables that the chart manages, such as authentication, host policy, and TLS settings, are rejected. |

Quote OAuth client IDs in values files. YAML reads an unquoted `26383066777993846573` as an imprecise floating-point number, and the chart rejects it.

The chart fails to render, with a message that starts with `infinity-rules-mcp:`, when bearer-mode values are missing, unsafe, or inconsistent.

### Example: passthrough mode

```yaml
autopilot:
  docker:
    infinityRulesMcp:
      image: "registry.example.com/platformservices/infinity-rules-skills-mcp@sha256:<digest>"
  infinityRulesMcp:
    enabled: true
    auth:
      enabled: true
      publicUrl: "https://mcp.example.com"
      passthrough:
        audiences:
          - "26383066777993846573"
          - "<ai-client-id>"
    allowedHosts:
      - "pega.example.com"
    ingress:
      enabled: true
      ingressClassName: "nginx"
      tls:
        secretName: "mcp-tls"
```

In Pega, create the public OAuth 2.0 client registration for AI clients (authorization code with PKCE, the redirect URIs of each MCP client, and a least-privilege access group), and add its client ID in place of `<ai-client-id>`.

### Example: exchange mode with Microsoft Entra ID

```yaml
autopilot:
  docker:
    infinityRulesMcp:
      image: "registry.example.com/platformservices/infinity-rules-skills-mcp@sha256:<digest>"
  infinityRulesMcp:
    enabled: true
    auth:
      enabled: true
      publicUrl: "https://mcp.example.com"
      tokenMode: exchange
      exchange:
        issuer: "https://login.microsoftonline.com/<tenant-id>/v2.0"
        audience: "<mcp-application-client-id>"
        clientId: "<pega-jwt-bearer-client-id>"
        clientSecret:
          name: "mcp-exchange-client"
          key: "client-secret"
        internalAudiences:
          - "26383066777993846573"
    allowedHosts:
      - "pega.example.com"
    ingress:
      enabled: true
      ingressClassName: "nginx"
      tls:
        secretName: "mcp-tls"
```

1. In Microsoft Entra ID, register an application for MCP. Under **Expose an API**, set an Application ID URI and add a scope for MCP clients. In the application manifest, set `requestedAccessTokenVersion` (`accessTokenAcceptedVersion` in older manifests) to `2`, so that tokens use the `v2.0` issuer above. Version 2 access tokens carry the MCP application's client ID as their audience, so set `auth.exchange.audience` to that client ID.
2. Register each MCP client as a public client with its redirect URIs, and grant it the MCP scope.
3. In Pega, create an OAuth 2.0 client registration that allows the JWT bearer grant, trusts the Entra ID issuer and its keys, and maps a token claim, such as `preferred_username`, to a Pega operator. Set its client ID in `auth.exchange.clientId`.
4. For a confidential registration, store its secret in the namespace of the release:

   ```sh
   kubectl create secret generic mcp-exchange-client \
     --from-literal=client-secret='<pega-client-secret>' -n <namespace>
   ```

### Ingress for external MCP clients

When `auth.enabled` and `ingress.enabled` are both `true`, the chart renders an Ingress named `infinity-rules-mcp-service` for the `auth.publicUrl` host with two paths: `/mcp` (`Exact`) and `/.well-known/oauth-protected-resource` (`Prefix`). The health endpoints under `/actuator` are never routed. The Ingress includes a TLS section when `ingress.tls.secretName` is set; `auth.publicUrl` must use a DNS name.

The chart cannot set the following controller settings portably. Configure them with `ingress.annotations` or your gateway:

- a read timeout of at least 300 seconds for long-running tool calls;
- no automatic retries of `POST /mcp`, because retries can repeat write operations;
- request-size limits;
- rate limiting. MCP has no per-caller limits, and Autopilot and external clients share the same pods.

For example, with the NGINX Ingress Controller:

```yaml
autopilot:
  infinityRulesMcp:
    ingress:
      annotations:
        nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
        nginx.ingress.kubernetes.io/proxy-send-timeout: "300"
        nginx.ingress.kubernetes.io/proxy-next-upstream: "off"
        nginx.ingress.kubernetes.io/proxy-body-size: "1m"
        nginx.ingress.kubernetes.io/limit-rps: "10"
```

### Configuring MCP clients

Configure each client with the MCP URL `<publicUrl>/mcp` and an `X-Pega-Host` header. For example, in VS Code `mcp.json`:

```json
{
  "servers": {
    "infinity-rules": {
      "type": "http",
      "url": "https://mcp.example.com/mcp",
      "headers": {
        "X-Pega-Host": "https://pega.example.com",
        "X-Pega-App-Alias": "<application-alias>"
      }
    }
  }
}
```

An unauthenticated request receives `401` with a `WWW-Authenticate` header that points to the OAuth protected resource metadata of MCP. The metadata names the authorization server: the Pega host in passthrough mode, or `auth.exchange.issuer` in exchange mode. When the client asks for an OAuth client ID, enter the Pega client ID in passthrough mode or the IdP client ID in exchange mode. Clients that cannot send a custom `X-Pega-Host` header are not supported.

> **Note:** MCP exposes write tools. Pega authorizes every call as the signed-in or mapped operator, so give AI-client users least-privilege access groups.

### Private certificate authority

The trust store replaces the default JVM trust store. When MCP must also reach publicly signed hosts, such as a cloud IdP, build the trust store from the JDK `cacerts` file plus your CA:

```sh
keytool -importkeystore -srckeystore "$JAVA_HOME/lib/security/cacerts" -srcstorepass changeit \
  -destkeystore truststore.p12 -deststoretype PKCS12 -deststorepass '<password>'
keytool -importcert -noprompt -alias pega-ca -file pega-ca.pem \
  -keystore truststore.p12 -storetype PKCS12 -storepass '<password>'
kubectl create secret generic mcp-truststore --from-file=truststore.p12 \
  --from-literal=password='<password>' -n <namespace>
```

```yaml
autopilot:
  infinityRulesMcp:
    trustStore:
      secretName: "mcp-truststore"
      key: "truststore.p12"
      type: PKCS12
      passwordKey: "password"
```

The chart mounts the Secret key read-only at `/opt/mcp/truststore` and passes the location, type, and password to the JVM through `JDK_JAVA_OPTIONS`. The password must not contain whitespace. The JVM prints `JDK_JAVA_OPTIONS` to the container log at startup, so the password appears in the log; a trust store holds only public certificates, so use a password that is not used anywhere else.

### Enabling bearer mode

1. **Upgrade the chart.** With `auth.enabled: false`, the MCP resources render identically, so MCP pods do not restart.
2. **Upgrade Autopilot** to a release that sends bearer tokens to MCP. It continues to work with trusted-header MCP.
3. **Check before switching:**
   - Autopilot debug logs (`MCP-HEADERS`, with credential values redacted) show `Authorization` on MCP calls, and no requests lack an operator. A request without an operator cannot obtain a token and fails in bearer mode.
   - The `X-Pega-Host` values that Autopilot sends are bare `https://` origins on the planned `allowedHosts`.
   - The Pega certificate is trusted by a standard JVM, or a trust store Secret is ready.
   - In exchange mode, the Autopilot client is in `auth.exchange.internalAudiences`, and the exchange client and its Secret exist.
4. **Enable bearer mode** in one `helm upgrade`: set `auth.enabled: true`, `auth.publicUrl`, the token-mode settings, `allowedHosts` and, if needed, `trustStore`, and keep `ingress.enabled: false`. The Deployment rolls one pod at a time. Upgraded Autopilot releases send both credential headers, so old and new pods both serve Autopilot during the rollout.
5. **Verify** an Autopilot turn that calls a rules tool. Readiness checks do not call Pega, so this is the functional check.
6. **Onboard external clients.** In passthrough mode, register the AI client in Pega and add its client ID to `auth.passthrough.audiences`. In exchange mode, register MCP and its clients in the IdP, configure the Pega JWT bearer client, and confirm the operator mapping. Then set `ingress.enabled: true`.

To **roll back**, set `ingress.enabled: false` and `auth.enabled: false` in one `helm upgrade`, or use `helm rollback`. MCP returns to trusted-header mode. Do not roll Autopilot back to a release that does not send bearer tokens while `auth.enabled` is `true`.

To **change the token mode** later, run one `helm upgrade`. Every external client must sign in again with the new authorization server. Autopilot is unaffected because its client is in the default list for both modes.

### Operations

- **Exchange client secret rotation:** MCP reads the secret at startup. Update the Secret, then restart the Deployment with `kubectl rollout restart deployment/infinity-rules-mcp-service -n <namespace>`.
- **IdP key rotation:** handled by issuer discovery. A pinned `jwksUri` must keep serving the current keys.
- **Adding a Pega host:** add it to `allowedHosts` and upgrade; the pods restart. In exchange mode, the new host also needs the JWT bearer client registration.
- **Network:** in exchange mode, egress controls and any proxy must allow the IdP as well as Pega.
- **In-cluster traffic:** Autopilot sends bearer tokens to MCP over cluster HTTP, as it sends trusted headers today. Restrict access with a service mesh or NetworkPolicy.
- **Exchange tuning:** `PEGA_EXCHANGE_CACHE_TTL` (default `30m`) and `PEGA_EXCHANGE_FAILURE_TTL` (default `60s`) can be set with `extraEnv`. In passthrough mode, the built-in default OAuth client of MCP is also accepted; to replace it with your AI-client registration, set `PEGA_OAUTH_CLIENT_ID` with `extraEnv`.

### Troubleshooting

| Symptom | Likely cause |
|---|---|
| `helm install` or `helm upgrade` fails with `infinity-rules-mcp: ...` | A missing or invalid value. The message names the setting. |
| MCP pods never become ready after you enable bearer mode | The MCP image is too old (images before 0.7.13 do not serve `/actuator/health`), or MCP stopped at startup because of a configuration error. Check the pod log. |
| Autopilot tool calls fail with `401` | Autopilot does not send `Authorization: Bearer`, the request has no operator, or the Autopilot client is missing from `auth.passthrough.audiences` (passthrough) or `auth.exchange.internalAudiences` (exchange). |
| `400 missing_pega_host` | The caller did not send `X-Pega-Host`. |
| `403 pega_host_rejected` | The `X-Pega-Host` value is not a bare `https://` origin on `allowedHosts`. |
| `401 invalid_token` | The token audience, issuer, or signature does not match: for example, a Pega client that is not in `auth.passthrough.audiences`, a token from another Pega host, or an IdP token with the wrong audience. |
| `401` with "The Pega instance rejected the token exchange" | The Pega JWT bearer client or operator mapping rejected the IdP token. MCP retries after 60 seconds. |
| `503 idp_trust_unavailable` or `503 pega_trust_unavailable` | MCP cannot retrieve signing keys from the IdP or the Pega host. Check egress, the proxy settings, and TLS trust. |
| TLS errors such as `PKIX path building failed` in the MCP log | The Pega or IdP certificate is not trusted. Configure `trustStore`. |
| External clients time out or see interrupted responses | The ingress read timeout is too short. |

## Provider credentials

The Autopilot Service supports two credential patterns for Azure OpenAI, AWS Bedrock, and Google Vertex AI.

### Option 1: Inline credentials

Specify provider credentials directly in your values file. When you use inline credentials, the chart automatically creates a Kubernetes Secret named `<deployment.name>-provider-credentials`.

| Configuration | Usage |
|---|---|
| `azure.endpoint` | Azure OpenAI endpoint URL, for example `https://YOUR_AZURE_RESOURCE.openai.azure.com/`. |
| `azure.apiKey` | Azure OpenAI API key. |
| `azure.apiVersion` | Azure OpenAI API version. Default: `2024-10-21` |
| `aws.accessKeyId` | AWS access key ID for Bedrock. |
| `aws.secretAccessKey` | AWS secret access key for Bedrock. |
| `aws.sessionToken` | Optional AWS session token for temporary credentials. |
| `vertex.credentials` | Base64-encoded Google service account JSON. The `project_id` is automatically extracted from the JSON. |
| `vertex.location` | Vertex AI location. Default: `us-central1` |

```yaml
autopilot:
  enabled: true
  azure:
    endpoint: "https://YOUR_AZURE_RESOURCE.openai.azure.com/"
    apiKey: "YOUR_AZURE_OPENAI_API_KEY"
  aws:
    accessKeyId: "YOUR_AWS_ACCESS_KEY_ID"
    secretAccessKey: "YOUR_AWS_SECRET_ACCESS_KEY"
    sessionToken: ""
  vertex:
    credentials: "YOUR_BASE64_ENCODED_VERTEX_SERVICE_ACCOUNT_JSON"
    location: "us-central1"
```

#### Updating inline credentials after deployment

When credentials change, run `helm upgrade` with the updated values file. The chart automatically re-encodes and replaces the Kubernetes Secret:

```bash
helm upgrade <RELEASE-NAME> <CHART-PATH> \
  --namespace <NAMESPACE> \
  -f <VALUES-FILE> \
  --wait --timeout 5m
```

> **Do not manually patch the Secret when using inline credentials.** The chart uses `b64enc` to encode values into the Secret. Patching the Secret directly can cause double base64 encoding and result in invalid credentials.

### Option 2: Existing Kubernetes Secret

Use a Secret that you create and manage outside of this chart. Set `providerCredentialsSecret` to the name of your Secret. The Secret should contain keys that match the environment variable names used by the service.

| Configuration | Usage |
|---|---|
| `providerCredentialsSecret` | Name of an existing Kubernetes Secret that contains provider credentials. When set, inline credentials are ignored. |

The Secret can contain credentials for all providers in a single Secret. Only the keys that are relevant to the providers you are using need to be present. All keys are mounted as `optional: true`, so missing keys are silently ignored.

| Key | Provider | Description |
|---|---|---|
| `AZURE_ENDPOINT` | Azure OpenAI | Azure OpenAI endpoint URL |
| `AZURE_OPENAI_KEY` | Azure OpenAI | Azure OpenAI API key |
| `AWS_ACCESS_KEY_ID` | AWS Bedrock | AWS access key ID |
| `AWS_SECRET_ACCESS_KEY` | AWS Bedrock | AWS secret access key |
| `AWS_SESSION_TOKEN` | AWS Bedrock | Optional AWS session token |
| `VERTEX_AUTH` | Google Vertex AI | Base64-encoded Google service account JSON. The Autopilot Service base64-decodes this value itself, so the Secret must contain the base64 string rather than the raw JSON. |

```yaml
autopilot:
  enabled: true
  providerCredentialsSecret: "autopilot-provider-credentials"
```

Create the Secret with credentials for all providers that you intend to use. For `VERTEX_AUTH`, pass the base64-encoded JSON so the pod receives the encoded string the service expects:

```bash
kubectl create secret generic autopilot-provider-credentials \
  --namespace <NAMESPACE> \
  --from-literal=AZURE_ENDPOINT="https://YOUR_AZURE_RESOURCE.openai.azure.com/" \
  --from-literal=AZURE_OPENAI_KEY="YOUR_AZURE_OPENAI_API_KEY" \
  --from-literal=AWS_ACCESS_KEY_ID="YOUR_AWS_ACCESS_KEY_ID" \
  --from-literal=AWS_SECRET_ACCESS_KEY="YOUR_AWS_SECRET_ACCESS_KEY" \
  --from-literal=AWS_SESSION_TOKEN="YOUR_AWS_SESSION_TOKEN" \
  --from-literal=VERTEX_AUTH="$(base64 -w0 /path/to/gcp-service-account.json)"
```

You can also create the Secret from a YAML manifest to manage it in source control. For `VERTEX_AUTH`, put the base64-encoded JSON string directly in `stringData`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: autopilot-provider-credentials
  namespace: <NAMESPACE>
type: Opaque
stringData:
  AZURE_ENDPOINT: "https://YOUR_AZURE_RESOURCE.openai.azure.com/"
  AZURE_OPENAI_KEY: "YOUR_AZURE_OPENAI_API_KEY"
  AWS_ACCESS_KEY_ID: "YOUR_AWS_ACCESS_KEY_ID"
  AWS_SECRET_ACCESS_KEY: "YOUR_AWS_SECRET_ACCESS_KEY"
  AWS_SESSION_TOKEN: "YOUR_AWS_SESSION_TOKEN"   # omit if not using temporary credentials
  VERTEX_AUTH: "YOUR_BASE64_ENCODED_VERTEX_SERVICE_ACCOUNT_JSON"  # base64 encode the JSON file: base64 -w0 gcp-service-account.json
```

#### Updating a credentials Secret after deployment

When credentials change, patch the existing Secret directly and restart the pod to pick up the new values:

```bash
# Replace the Secret with updated values
kubectl create secret generic autopilot-provider-credentials \
  --namespace <NAMESPACE> \
  --from-literal=AZURE_ENDPOINT="https://YOUR_AZURE_RESOURCE.openai.azure.com/" \
  --from-literal=AZURE_OPENAI_KEY="YOUR_AZURE_OPENAI_API_KEY" \
  --from-literal=AWS_ACCESS_KEY_ID="YOUR_AWS_ACCESS_KEY_ID" \
  --from-literal=AWS_SECRET_ACCESS_KEY="YOUR_AWS_SECRET_ACCESS_KEY" \
  --from-literal=AWS_SESSION_TOKEN="YOUR_AWS_SESSION_TOKEN" \
  --from-literal=VERTEX_AUTH="$(base64 -w0 /path/to/gcp-service-account.json)" \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart the pod to pick up the updated Secret
kubectl rollout restart deployment/<DEPLOYMENT-NAME> -n <NAMESPACE>
kubectl rollout status deployment/<DEPLOYMENT-NAME> -n <NAMESPACE> --timeout=120s
```

## Model list configuration

The Autopilot Service uses a model list to determine which models are available and how requests are routed to each provider. When a model list is mounted by the chart, the container receives it through the `LOCAL_MODELS_FILE` environment variable at `/config/models.json`. The `model_id` field format differs by provider.

### Building model list

#### For Azure OpenAI and Vertex AI

- **`name`** must match the deployment name as it appears in the LLM provider console (for example, the Azure OpenAI Studio deployment name or the Vertex AI model ID shown in Model Garden). This value is used as the display name and routing key.
- **`model_path`** must be provided as an array of API endpoint paths relative to the provider base URL. For Azure OpenAI, the path embeds the Azure portal deployment name (for example, `["/openai/deployments/gpt-5/chat/completions"]`). For Vertex AI, it embeds the model identifier (for example, `["/google/deployments/gemini-2.5-pro/chat/completions"]`).

#### For AWS Bedrock

- **`model_id`** must be the exact model ID as shown in the AWS Bedrock console, including the cross-region inference prefix and version suffix (for example, `us.anthropic.claude-3-7-sonnet-20250219-v1:0`, `us.amazon.nova-pro-v1:0`, `amazon.titan-embed-text-v2:0`).

### Option 1: Bundled models file (recommended)

By default, the chart creates a ConfigMap from `files/default-models.json` during `helm install` or `helm upgrade`.

| Configuration | Usage |
|---|---|
| `deployModelsConfigMap` | Set to `true` to create a ConfigMap from the bundled `files/default-models.json` file. Default: `true` |

```yaml
autopilot:
  enabled: true
  deployModelsConfigMap: true
```

To customize the default model list before deployment, edit `files/default-models.json` in the chart directory.

#### Updating the bundled model list after deployment

After you edit `files/default-models.json`, apply the changes to the running deployment without a full `helm upgrade`:

```bash
# Update the ConfigMap from the bundled file
kubectl create configmap autopilot-models \
  --namespace <NAMESPACE> \
  --from-file=models.json=./files/default-models.json \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart the pod to pick up the new ConfigMap
kubectl rollout restart deployment/<DEPLOYMENT-NAME> -n <NAMESPACE>
kubectl rollout status deployment/<DEPLOYMENT-NAME> -n <NAMESPACE> --timeout=120s
```

### Option 2: Existing ConfigMap

Use a ConfigMap that you create and manage outside of this chart. The ConfigMap must include a `models.json` key.

| Configuration | Usage |
|---|---|
| `customModels.existingConfigMap` | Name of an existing ConfigMap that contains `models.json`. |

```yaml
autopilot:
  enabled: true
  deployModelsConfigMap: false
  customModels:
    existingConfigMap: "autopilot-models"
```

Create the ConfigMap:

```bash
kubectl create configmap autopilot-models \
  --namespace <NAMESPACE> \
  --from-file=models.json=./my-models.json
```

#### Updating a model list ConfigMap after deployment

When the contents of your external ConfigMap change, re-apply it and restart the pod:

```bash
# Re-apply the updated ConfigMap
kubectl create configmap autopilot-models \
  --namespace <NAMESPACE> \
  --from-file=models.json=./my-models.json \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart the pod to pick up the updated ConfigMap
kubectl rollout restart deployment/<DEPLOYMENT-NAME> -n <NAMESPACE>
kubectl rollout status deployment/<DEPLOYMENT-NAME> -n <NAMESPACE> --timeout=120s
```

### Option 3: Inline model list

Specify the JSON content directly in `values.yaml`. The chart creates a ConfigMap from the inline content.

| Configuration | Usage |
|---|---|
| `customModels.inline` | JSON string containing the model list. |

```yaml
autopilot:
  enabled: true
  deployModelsConfigMap: false
  customModels:
    inline: |
      [
        {
          "provider": "azure",
          "creator": "openai",
          "model_name": "GPT-5",
          "model_mapping_id": "gpt-5-2025-08-07",
          "name": "gpt-5-2025-08-07",
          "model_id": "gpt-5-2025-08-07",
          "input_tokens": 400000,
          "output_tokens": 128000,
          "type": "chat_completion",
          "version": "2025-08-07",
          "model_path": ["/openai/deployments/gpt-5-2025-08-07/chat/completions"],
          "supported_capabilities": {
            "streaming": true,
            "functions": true,
            "json_mode": true
          }
        }
      ]
```

If you set `deployModelsConfigMap: false` and do not provide `customModels.existingConfigMap` or `customModels.inline`, the service uses its built-in model list.

### Model JSON format

| Field | Required | Description |
|---|---|---|
| `provider` | Yes | Provider name: `azure`, `bedrock`, `vertex`, or `custom`. |
| `creator` | Yes | Model creator, for example `openai`, `anthropic`, `google`, or `amazon`. |
| `model_name` | Yes | Display name for the model. |
| `name` | Yes | Routing name used by the Autopilot Service. For Azure OpenAI and Vertex AI, this value must match the deployment name shown in the provider console. |
| `model_id` | Yes | Provider-native model identifier. For Azure OpenAI and Vertex AI, use the deployment name as shown in the provider console. For AWS Bedrock, use the exact model ID from the AWS Bedrock console, including any cross-region prefix or version suffix. |
| `model_mapping_id` | Yes | Deployment or provider-specific identifier used to build `model_path` entries. |
| `model_path` | Yes | Array of API endpoint paths used by the service to reach the provider API. For Azure OpenAI, use `["/openai/deployments/<model_mapping_id>/chat/completions"]`. For Vertex AI, use `["/google/deployments/<model_mapping_id>/chat/completions"]`. For Bedrock, use `["/anthropic/deployments/<model_mapping_id>/converse", "/anthropic/deployments/<model_mapping_id>/converse-stream"]` (adjust the creator prefix to match the model creator). |
| `type` | Yes | Model type, for example `chat_completion`, `embedding`, or `image`. |
| `version` | Yes | Model version string. |
| `input_tokens` | No | Maximum supported input tokens. |
| `output_tokens` | No | Maximum supported output tokens. |
| `supported_capabilities` | No | Object that describes supported capabilities such as streaming, function calling, or multimodal support. |
| `parameters` | No | Object that describes tunable model parameters, such as `temperature`, `top_p`, or `max_tokens`. |

## Default fast and smart models

The `fast` and `smart` default models are defined explicitly in a dedicated `default_models` section at the top level of the models JSON file.

The `default_models` section contains two keys — `fast` and `smart` — each holding the full model object of the intended default:

```json
{
  "models": [ ... ],
  "default_models": {
    "fast": {
      "provider": "azure",
      "model_id": "gpt-5-chat-2025-08-07",
      ...
    },
    "smart": {
      "provider": "azure",
      "model_id": "gpt-5-mini-2025-08-07",
      ...
    }
  }
}
```

To change which models serve as defaults, update the `default_models.fast` and `default_models.smart` entries in your models JSON file to the desired model objects. The model referenced must also be present in the `models` array.

## Liveness and readiness probes

The Autopilot Service uses liveness and readiness probes on the `/v1/health` endpoint. Configure probes as part of a `livenessProbe` or `readinessProbe` configuration.

| Parameter | Description | Default `livenessProbe` | Default `readinessProbe` |
|---|---|---|---|
| `initialDelaySeconds` | Number of seconds to wait before starting probes. | `10` | `10` |
| `timeoutSeconds` | Probe timeout in seconds. | `10` | `10` |
| `periodSeconds` | Probe interval in seconds. | `10` | `10` |
| `successThreshold` | Minimum consecutive successes required after a failure. | `1` | `1` |
| `failureThreshold` | Number of failures before Kubernetes marks the probe as failed. | `3` | `3` |

Example:

```yaml
livenessProbe:
  initialDelaySeconds: 10
  timeoutSeconds: 10
  periodSeconds: 10
  successThreshold: 1
  failureThreshold: 3
readinessProbe:
  initialDelaySeconds: 10
  timeoutSeconds: 10
  periodSeconds: 10
  successThreshold: 1
  failureThreshold: 3
```

## Ingress

To expose the Autopilot Service outside the cluster, enable ingress.

| Configuration | Usage |
|---|---|
| `ingress.enabled` | Leave as `false` unless you need to expose the Autopilot Service outside the cluster. |
| `ingress.domain` | Specify the ingress host name. |
| `ingress.ingressClassName` | Specify the ingress class name. |
| `ingress.tls.enabled` | Set to `true` to enable TLS termination at the ingress. |
| `ingress.tls.secretName` | Specify the Kubernetes Secret that stores the TLS certificate. |
| `ingress.annotations` | Specify additional annotations for the ingress. |

```yaml
ingress:
  enabled: true
  domain: "autopilot.example.com"
  ingressClassName: "nginx"
  tls:
    enabled: true
    secretName: "autopilot-tls-secret"
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
```

## Connecting Pega Platform to the Autopilot Service

After you deploy the Autopilot Service, configure Pega Platform to use the service URL using Dynamic System Setting (DSS) or `prconfig.xml`.

> If both methods are configured, `prconfig.xml` takes precedence over the DSS.

### Option 1: Dynamic System Setting (DSS)

1. Log in to Pega Platform as an administrator.
2. Navigate to **Records > SysAdmin > Dynamic System Settings** and create a new DSS with the following values:

| Field | Value |
|---|---|
| Owning Ruleset | `Pega-Engine` |
| Setting Purpose | `prconfig/services/genai/autopilot/servicebaseurl/default` |
| Value | `http://<service-name>.<namespace>.svc.cluster.local:<service-port>/` |

3. Replace `<service-name>` with the Autopilot Service name (default: `autopilot`) and `<namespace>` with the Kubernetes namespace where the service is deployed.
4. Save the DSS.

**Example:** If the Autopilot Service is deployed with the default name `autopilot` in the `autopilot` namespace:

```text
http://autopilot.autopilot.svc.cluster.local/
```

5. Restart Pega Platform nodes for the DSS change to take effect.

### Option 2: `prconfig.xml`

Add the Autopilot Service URL directly to `charts/pega/config/deploy/prconfig.xml` in the Pega Helm charts repository:

```xml
<env name="services/genai/autopilot/servicebaseurl" value="http://<service-name>.<namespace>.svc.cluster.local:<service-port>/"/>
```

Example:

```xml
<env name="services/genai/autopilot/servicebaseurl" value="http://autopilot.autopilot.svc.cluster.local/"/>
```

After you edit `prconfig.xml`, apply the change with a `helm upgrade` followed by a rollout restart:

```bash
helm upgrade pega <PEGA-CHART-PATH> -f my-values.yaml -n <PEGA-NAMESPACE>
kubectl rollout restart statefulset/<PEGA-DEPLOYMENT-NAME> -n <PEGA-NAMESPACE>
```

A restart of Pega Platform is required whenever the `prconfig.xml` entry is added or changed.

## OAuth authentication between Pega Platform and the Autopilot Service

You can enable OAuth authentication to secure requests between Pega Platform and the Autopilot Service using an Identity Provider (IdP). The Autopilot Service only supports the `private_key_jwt` authentication type.

### How it works

- Pega Platform obtains a bearer token from your IdP using an OAuth 2.0 `client_credentials` grant.
- The token is attached as an `Authorization: Bearer <token>` header on every request to the Autopilot Service.
- The Autopilot Service validates incoming tokens against the IdP public key endpoint (`oauthPublicKeyURL`).

### Scopes

The Autopilot Service does not require any OAuth scopes. Leave `autopilot.autopilotAuth.scopes` empty (or omit it entirely) when configuring the Pega chart.

### Autopilot Service configuration in the backingservices chart

Enable auth on the Autopilot Service and set the IdP public key URL so it can validate incoming tokens:

| Parameter | Description | Default |
|---|---|---|
| `authEnabled` | Enables token validation on incoming requests to the Autopilot Service. | `false` |
| `oauthPublicKeyURL` | URL of the IdP public key endpoint used to verify bearer tokens. Required when `authEnabled` is `true`. | `""` |

```yaml
autopilot:
  enabled: true
  authEnabled: true
  oauthPublicKeyURL: "https://YOUR_IDP_HOST/oauth2/v1/keys"
```

### Pega Platform configuration in the `pega` chart

Configure the Pega chart to mint tokens and attach them to Autopilot requests. The Autopilot Service URL itself is configured via a DSS or `prconfig.xml` (see [Connecting Pega Platform to the Autopilot Service](#connecting-pega-platform-to-the-autopilot-service)) — no URL parameter is needed in the `pega` chart.

| Parameter | Description | Default |
|---|---|---|
| `autopilot.autopilotAuth.enabled` | Enables OAuth token minting on the Pega Platform side. | `false` |
| `autopilot.autopilotAuth.url` | URL of the OAuth service endpoint to obtain a token. | `""` |
| `autopilot.autopilotAuth.clientId` | OAuth client ID. | `""` |
| `autopilot.autopilotAuth.authType` | Authentication type. Only `private_key_jwt` is supported. | `"private_key_jwt"` |
| `autopilot.autopilotAuth.privateKey` | Base64-encoded PKCS8 private key. | `""` |
| `autopilot.autopilotAuth.privateKeyAlgorithm` | Algorithm for the private key. Allowed values: `RS256`, `RS384`, `RS512`, `ES256`, `ES384`, `ES512`. Default: `RS256` | `""` |
| `autopilot.autopilotAuth.scopes` | OAuth scopes to request. The Autopilot Service does not require any scopes — leave this empty. | `""` |
| `autopilot.autopilotAuth.external_secret_name` | Name of an existing Kubernetes Secret containing the key (key: `AUTOPILOT_OAUTH_PRIVATE_KEY`). When set, `privateKey` is ignored and no Secret is created by the chart. | `""` |

```yaml
autopilot:
  autopilotAuth:
    enabled: true
    url: "https://YOUR_IDP_HOST/oauth2/v1/token"
    clientId: "YOUR_CLIENT_ID"
    authType: "private_key_jwt"
    privateKey: "YOUR_BASE64_ENCODED_PKCS8_PRIVATE_KEY"
    privateKeyAlgorithm: "RS256"
    scopes: ""   # no scopes required for Autopilot
```

### Using an existing Secret for the private key

To avoid placing the private key directly in `values.yaml`, create a Kubernetes Secret beforehand and reference it:

```bash
kubectl create secret generic autopilot-auth-secret \
  --namespace <PEGA-NAMESPACE> \
  --from-literal=AUTOPILOT_OAUTH_PRIVATE_KEY="YOUR_BASE64_ENCODED_PKCS8_PRIVATE_KEY"
```

Then reference the Secret in the `pega` chart:

```yaml
autopilot:
  autopilotAuth:
    enabled: true
    external_secret_name: "autopilot-auth-secret"
```

### Shared credentials with SRS and token-minting precedence

Pega Platform uses a single set of `SERV_AUTH_*` environment variables to mint bearer tokens for backing services. When both SRS auth (`pegasearch.srsAuth`) and Autopilot auth (`autopilot.autopilotAuth`) are enabled at the same time, **the SRS credentials take precedence** and are used to mint tokens sent to both SRS and the Autopilot Service.

This means:

- Both SRS and Autopilot can share the same IdP client application and credentials.
- You do not need separate client registrations for each backing service if both are pointed at the same IdP authorization server.
- If only `autopilot.autopilotAuth` is enabled (SRS auth is disabled), the Autopilot credentials are used to mint tokens.
- If both are enabled and the credentials differ, the SRS credentials win — the Autopilot-specific credentials are not used for token minting.

In practice, configure both backing services to trust the same IdP public key endpoint and issue tokens from the same client application. The recommended setup when both services are deployed together:

```yaml
pegasearch:
  srsAuth:
    enabled: true
    url: "https://YOUR_IDP_HOST/oauth2/v1/token"
    clientId: "YOUR_SHARED_CLIENT_ID"
    authType: "private_key_jwt"
    privateKey: "YOUR_BASE64_ENCODED_PKCS8_PRIVATE_KEY"
autopilot:
  autopilotAuth:
    enabled: true
    url: "https://YOUR_IDP_HOST/oauth2/v1/token"
    clientId: "YOUR_SHARED_CLIENT_ID"
    privateKey: "YOUR_BASE64_ENCODED_PKCS8_PRIVATE_KEY"
    scopes: ""   # Autopilot requires no scopes
```

Because SRS takes precedence, the token sent to Autopilot is minted using the SRS credentials above. Both services must therefore trust tokens issued for the same client.

## Custom OpenAI-compatible providers

The Autopilot Service supports OpenAI-compatible providers (such as OpenRouter, private LLM servers, or other hosted endpoints) through the `customOpenAI` section. Each provider is identified by a `creator` value that maps to environment variables and to the `creator` field in the model JSON.

### Configuration settings

| Configuration | Usage |
|---|---|
| `customOpenAI.providers` | List of custom OpenAI-compatible provider configurations. Each entry produces a set of environment variables in the pod. |
| `customOpenAI.providers[].creator` | Provider identifier (for example, `openrouter`, `my-llm-server`). Must match the `creator` field in model entries. |
| `customOpenAI.providers[].baseUrl` | Base URL for the provider API (for example, `https://openrouter.ai/api/v1`). |
| `customOpenAI.providers[].apiKey` | Provider API key. The chart stores inline values in a Secret. |
| `customOpenAI.existingSecret` | Name of an existing Secret that contains provider API keys. When set, `apiKey` fields are not required. |

For each entry in `customOpenAI.providers`, the chart injects these environment variables into the pod:

- `CUSTOM_OPENAI_PROVIDERS` as a comma-separated list of all `creator` values
- `CUSTOM_OPENAI_<CREATOR>_BASE_URL` for the provider base URL
- `CUSTOM_OPENAI_<CREATOR>_API_KEY` for the provider API key

`<CREATOR>` is the `creator` value uppercased with hyphens and dots replaced by underscores. For example, `creator: openrouter` produces `CUSTOM_OPENAI_OPENROUTER_BASE_URL` and `CUSTOM_OPENAI_OPENROUTER_API_KEY`.

### Option 1: Inline API key

```yaml
autopilot:
  customOpenAI:
    providers:
      - creator: openrouter
        baseUrl: "https://openrouter.ai/api/v1"
        apiKey: "YOUR_OPENROUTER_API_KEY"
```

Multiple providers can be configured at once:

```yaml
autopilot:
  customOpenAI:
    providers:
      - creator: openrouter
        baseUrl: "https://openrouter.ai/api/v1"
        apiKey: "YOUR_OPENROUTER_API_KEY"
      - creator: my-llm-server
        baseUrl: "https://my-llm.internal/v1"
        apiKey: "YOUR_INTERNAL_API_KEY"
```

### Option 2: Existing Kubernetes Secret

Create a Secret manually and reference it via `existingSecret`. When set, inline `apiKey` values are not required.

```yaml
autopilot:
  customOpenAI:
    existingSecret: "custom-openai-credentials"
    providers:
      - creator: openrouter
        baseUrl: "https://openrouter.ai/api/v1"
        # apiKey omitted — read from existingSecret
```

The Secret must use keys in the format `CUSTOM_OPENAI_<CREATOR>_API_KEY`. For `creator: openrouter`, the required key is `CUSTOM_OPENAI_OPENROUTER_API_KEY`.

Create the Secret:

```bash
kubectl create secret generic custom-openai-credentials \
  --namespace <NAMESPACE> \
  --from-literal=CUSTOM_OPENAI_OPENROUTER_API_KEY="YOUR_OPENROUTER_API_KEY"
```

### Adding custom OpenAI models to the model list

Models for custom OpenAI-compatible providers use `provider: "custom"` in the model JSON. The `creator` field must match the `creator` value configured in `customOpenAI.providers`. The `name` field is the model identifier sent to the provider API.

> **Important:** `model_name` must not contain slashes. Use `name` for the actual API model identifier when it contains slashes (for example, `openrouter/free`).

Example model entry:

```json
{
  "provider": "custom",
  "creator": "openrouter",
  "model_name": "Free",
  "model_mapping_id": "openrouter-free",
  "name": "openrouter/free",
  "input_tokens": 200000,
  "output_tokens": 8192,
  "type": "chat_completion",
  "model_id": "custom/openrouter/Free/1",
  "version": "1",
  "model_path": ["/chat/completions"],
  "supported_capabilities": {
    "streaming": true,
    "functions": true,
    "json_mode": true,
    "is_multimodal": false
  }
}
```

To use this model as the default fast and smart model, set it in the `default_models` section of your models JSON:

```json
{
  "models": [ ... ],
  "default_models": {
    "fast": { "provider": "custom", "creator": "openrouter", "model_id": "custom/openrouter/Free/1", ... },
    "smart": { "provider": "custom", "creator": "openrouter", "model_id": "custom/openrouter/Free/1", ... }
  }
}
```

### Model JSON fields for custom providers

| Field | Description |
|---|---|
| `provider` | Must be `custom`. |
| `creator` | Must match the `creator` value in `customOpenAI.providers`. Used to look up `CUSTOM_OPENAI_<CREATOR>_BASE_URL` and `CUSTOM_OPENAI_<CREATOR>_API_KEY`. |
| `model_name` | Display name for the model. Must not contain slashes. |
| `name` | Actual model identifier sent to the provider API as the `model` parameter (for example, `openrouter/free`). |
| `model_path` | Array with the API path suffix appended to `baseUrl` (for example, `["/chat/completions"]`). |
| `model_id` | Must follow the format `custom/<creator>/<model_name>/<version>` (no slashes in `model_name`). |

## Example deployment configuration

```yaml
autopilot:
  enabled: true
  deployment:
    name: "autopilot"

  docker:
    registry:
      url: "YOUR_REGISTRY_URL"
      username: "YOUR_REGISTRY_USERNAME"
      password: "YOUR_REGISTRY_PASSWORD"
    autopilot:
      image: "YOUR_REGISTRY_URL/autopilot-service:YOUR_TAG"
      imagePullPolicy: Always
    infinityRulesMcp:
      image: "YOUR_REGISTRY_URL/infinity-rules-mcp:YOUR_TAG"
      imagePullPolicy: IfNotPresent

  replicas: 2
  enableGenaiHub: false
  authEnabled: false
  isInternalDeployment: false
  modelProviders: "Azure,Vertex,Bedrock"
  deployModelsConfigMap: true

  infinityRulesMcp:
    enabled: false

  azure:
    endpoint: "https://YOUR_AZURE_RESOURCE.openai.azure.com/"
    apiKey: "YOUR_AZURE_OPENAI_API_KEY"

  aws:
    accessKeyId: "YOUR_AWS_ACCESS_KEY_ID"
    secretAccessKey: "YOUR_AWS_SECRET_ACCESS_KEY"

  vertex:
    credentials: "YOUR_BASE64_ENCODED_VERTEX_SERVICE_ACCOUNT_JSON"
    location: "us-central1"

  customOpenAI:
    providers:
      - creator: openrouter
        baseUrl: "https://openrouter.ai/api/v1"
        apiKey: "YOUR_OPENROUTER_API_KEY"

  resources:
    requests:
      cpu: 500m
      memory: "1Gi"
    limits:
      cpu: "1"
      memory: "2Gi"
```
