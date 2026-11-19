{{- /*
imagePullSecret
backingservicesRegistrySecret
deploymentName
tlssecretsnippet
backingservices.gke.backendConfig
podAffinity
tolerations
are copied from backingservices/templates/_supplemental.tpl because helm lint requires
charts to render standalone. See: https://github.com/helm/helm/issues/11260 for more details.
*/}}

{{- define "imagePullSecret" }}
{{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .Values.docker.registry.url (printf "%s:%s" .Values.docker.registry.username .Values.docker.registry.password | b64enc) | b64enc }}
{{- end }}

{{- define "backingservicesRegistrySecret" }}
{{- $depName := printf "%s" (include "deploymentName" (dict "root" .root "defaultname" .defaultname )) -}}
{{- $depName -}}-registry-secret
{{- end }}

{{- define "deploymentName" }}{{ $deploymentNamePrefix := .defaultname }}{{ if (.root.deployment) }}{{ if (.root.deployment.name) }}{{ $deploymentNamePrefix = .root.deployment.name }}{{ end }}{{ end }}{{ if (.root.name) }}{{ $deploymentNamePrefix = .root.name }}{{ end }}{{ $deploymentNamePrefix }}{{- end }}

{{- define "tlssecretsnippet" -}}
tls:
- hosts:
  - {{ include "domainName" (dict "node" .node) }}
  secretName: {{ .node.ingress.tls.secretName }}
{{- end -}}

{{- define "domainName" }}
  {{- if .node.ingress -}}
  {{- if .node.ingress.domain -}}
    {{ .node.ingress.domain }}
  {{- end -}}
  {{- else if .node.service.domain -}}
    {{ .node.service.domain }}
  {{- end -}}
{{- end }}

{{- define "podAffinity" }}
{{- if .affinity }}
affinity:
{{- toYaml .affinity | nindent 2 }}
{{- end }}
{{ end }}

{{- define "tolerations" }}
{{- if .tolerations }}
tolerations:
{{- toYaml .tolerations | nindent 2 }}
{{- end }}
{{ end }}

{{/*
Validates and returns the OAuth public key URL when auth is enabled.
Fails render if authEnabled is true but oauthPublicKeyURL is not set.
*/}}
{{- define "autopilot.oauthPublicKeyUrl" -}}
{{- if .Values.authEnabled }}
    {{- if .Values.oauthPublicKeyURL }}
    {{- .Values.oauthPublicKeyURL | quote }}
    {{- else }}
    {{- fail "A valid entry is required for oauthPublicKeyURL when authEnabled is true. Set oauthPublicKeyURL to the IdP public key endpoint so the Autopilot service can validate incoming tokens from Pega Infinity." | quote }}
    {{- end }}
{{- end }}
{{- end }}

{{/*
Autopilot secret name - uses pre-existing secret or auto-generated one
*/}}
{{- define "autopilot.credentialsSecretName" -}}
{{- if .Values.providerCredentialsSecret -}}
{{- .Values.providerCredentialsSecret -}}
{{- else -}}
{{- $depName := include "deploymentName" (dict "root" .Values "defaultname" "autopilot") -}}
{{- printf "%s-provider-credentials" $depName -}}
{{- end -}}
{{- end -}}

{{/*
Check if any inline credentials are provided
*/}}
{{- define "autopilot.hasInlineCredentials" -}}
{{- if or .Values.azure.endpoint .Values.azure.apiKey .Values.aws.accessKeyId .Values.vertex.credentials -}}
true
{{- else if .Values.customOpenAI.providers -}}
{{- range .Values.customOpenAI.providers -}}
{{- if .apiKey -}}
true
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Check if any models config is provided (inline, existing configmap, or use default bundled file)
*/}}
{{- define "autopilot.hasCustomModels" -}}
{{- if and .Values.customModels (or .Values.customModels.existingConfigMap .Values.customModels.inline) -}}
true
{{- else if .Values.deployModelsConfigMap -}}
true
{{- end -}}
{{- end -}}

{{/*
Check if we need to create a ConfigMap (inline or default file, but NOT existing configmap)
*/}}
{{- define "autopilot.hasModelsConfig" -}}
{{- if and .Values.customModels .Values.customModels.inline -}}
true
{{- else if .Values.deployModelsConfigMap -}}
true
{{- end -}}
{{- end -}}

{{/*
Resolve the models ConfigMap name
*/}}
{{- define "autopilot.modelsConfigMapName" -}}
{{- if and .Values.customModels .Values.customModels.existingConfigMap -}}
{{- .Values.customModels.existingConfigMap -}}
{{- else -}}
{{- $depName := include "deploymentName" (dict "root" .Values "defaultname" "autopilot") -}}
{{- printf "%s-models" $depName -}}
{{- end -}}
{{- end -}}

{{/*
Check if any custom OpenAI providers are configured.
Returns "true" when the providers list is non-empty.
*/}}
{{- define "autopilot.hasCustomOpenAI" -}}
{{- if .Values.customOpenAI.providers -}}
true
{{- end -}}
{{- end -}}

{{/*
Resolve the Secret name that holds custom OpenAI API keys.
Uses customOpenAI.existingSecret when set; otherwise falls back to the
shared provider-credentials Secret (same Secret as Azure/AWS/Vertex keys).
*/}}
{{- define "autopilot.customOpenAISecretName" -}}
{{- if .Values.customOpenAI.existingSecret -}}
{{- .Values.customOpenAI.existingSecret -}}
{{- else -}}
{{- include "autopilot.credentialsSecretName" . -}}
{{- end -}}
{{- end -}}

{{/*
Convert a custom OpenAI creator alias to the upper-snake-case env var suffix.
e.g.  "openrouter"    → "OPENROUTER"
      "my-llm-server" → "MY_LLM_SERVER"
      "together.ai"   → "TOGETHER_AI"
Usage: {{ include "autopilot.creatorEnvSuffix" "openrouter" }}
*/}}
{{- define "autopilot.creatorEnvSuffix" -}}
{{- . | upper | replace "-" "_" | replace "." "_" -}}
{{- end -}}

{{/*
Infinity Rules MCP: host name of infinityRulesMcp.auth.mcpBaseUrl, used for the Ingress rule and TLS host.
*/}}
{{- define "infinityRulesMcp.publicHost" -}}
{{- (urlParse (toString .Values.infinityRulesMcp.auth.mcpBaseUrl)).hostname | lower -}}
{{- end -}}

{{/*
Infinity Rules MCP: JVM options for the trust store, rendered into JDK_JAVA_OPTIONS. Empty when no trust store is set.
$(MCP_TRUSTSTORE_PASSWORD) is expanded by Kubernetes from the variable declared before JDK_JAVA_OPTIONS.
*/}}
{{- define "infinityRulesMcp.jdkJavaOptions" -}}
{{- $trustStore := .Values.infinityRulesMcp.trustStore -}}
{{- if $trustStore.secretName -}}
{{- $options := list (printf "-Djavax.net.ssl.trustStore=/opt/mcp/truststore/%s" $trustStore.key) (printf "-Djavax.net.ssl.trustStoreType=%s" $trustStore.type) -}}
{{- if $trustStore.passwordKey -}}
{{- $options = append $options "-Djavax.net.ssl.trustStorePassword=$(MCP_TRUSTSTORE_PASSWORD)" -}}
{{- end -}}
{{- join " " $options -}}
{{- end -}}
{{- end -}}

{{/*
Infinity Rules MCP: environment variables added in bearer mode (infinityRulesMcp.auth.enabled), after the
variables shared with trusted-header mode. Only the active token mode's variables are rendered.
*/}}
{{- define "infinityRulesMcp.bearerEnv" -}}
{{- $mcp := .Values.infinityRulesMcp -}}
{{- $auth := $mcp.auth -}}
{{- /* Autopilot calls MCP with Pega tokens issued to this client, so they are accepted in both token modes. */ -}}
{{- $autopilotClientId := "26383066777993846573" -}}
- name: PEGA_INBOUND_AUTH
  value: "bearer"
- name: PEGA_TOKEN_MODE
  value: {{ $auth.tokenMode | quote }}
- name: MCP_PUBLIC_URL
  value: {{ trimSuffix "/" (toString $auth.mcpBaseUrl) | quote }}
- name: PEGA_ALLOWED_HOSTS
  value: {{ join "," $mcp.allowedHosts | quote }}
- name: SERVER_MAX_HTTP_REQUEST_HEADER_SIZE
  value: "16KB"
{{- if eq $auth.tokenMode "exchange" }}
{{- $exchange := $auth.exchange }}
- name: PEGA_EXCHANGE_ISSUER
  value: {{ $exchange.issuer | quote }}
{{- with $exchange.jwksUri }}
- name: PEGA_EXCHANGE_JWKS_URI
  value: {{ . | quote }}
- name: PEGA_EXCHANGE_DISCOVERY
  value: "false"
{{- end }}
{{- with $exchange.audience }}
- name: PEGA_EXCHANGE_AUDIENCE
  value: {{ . | quote }}
{{- end }}
- name: PEGA_EXCHANGE_ALGS
  value: {{ join "," $exchange.algorithms | quote }}
- name: PEGA_EXCHANGE_CLIENT_ID
  value: {{ toString $exchange.clientId | quote }}
- name: PEGA_EXCHANGE_CLIENT_AUTH_METHOD
  value: {{ $exchange.clientAuthMethod | quote }}
{{- with $exchange.clientSecret.name }}
- name: PEGA_EXCHANGE_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ . | quote }}
      key: {{ $exchange.clientSecret.key | quote }}
{{- end }}
- name: PEGA_EXCHANGE_INTERNAL_AUDIENCES
  value: {{ $autopilotClientId | quote }}
{{- else }}
- name: PEGA_TOKEN_AUDIENCES
  value: {{ prepend (default (list) $auth.passthrough.audiences) $autopilotClientId | uniq | join "," | quote }}
{{- end }}
{{- if and $mcp.trustStore.secretName $mcp.trustStore.passwordKey }}
- name: MCP_TRUSTSTORE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ $mcp.trustStore.secretName | quote }}
      key: {{ $mcp.trustStore.passwordKey | quote }}
{{- end }}
{{- with include "infinityRulesMcp.jdkJavaOptions" . }}
- name: JDK_JAVA_OPTIONS
  value: {{ . | quote }}
{{- end }}
{{- range $mcp.extraEnv }}
- name: {{ .name | quote }}
  {{- if not (kindIs "invalid" .value) }}
  value: {{ toString .value | quote }}
  {{- end }}
  {{- with .valueFrom }}
  valueFrom:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Infinity Rules MCP: fails rendering on unsafe or inconsistent values. Rule IDs match the chart design document.
Called from the MCP Deployment, which renders only when enabled and infinityRulesMcp.enabled are true.
*/}}
{{- define "infinityRulesMcp.validate" -}}
{{- $mcp := .Values.infinityRulesMcp -}}
{{- $auth := $mcp.auth -}}
{{- $prefix := "infinity-rules-mcp: " -}}
{{- $hostPattern := `^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?)*$` -}}

{{- /* V1 */ -}}
{{- if and $mcp.ingress.enabled (not $auth.enabled) -}}
{{- fail (print $prefix "infinityRulesMcp.ingress requires infinityRulesMcp.auth.enabled; trusted-header mode must not be exposed") -}}
{{- end -}}

{{- if $auth.enabled -}}

{{- /* V2 */ -}}
{{- $image := toString .Values.docker.infinityRulesMcp.image -}}
{{- if or (not (trim $image)) (hasPrefix "YOUR_" $image) -}}
{{- fail (print $prefix "docker.infinityRulesMcp.image is required") -}}
{{- end -}}

{{- /* V3 */ -}}
{{- if or (not $mcp.allowedHosts) (not (kindIs "slice" $mcp.allowedHosts)) -}}
{{- fail (print $prefix "infinityRulesMcp.allowedHosts must list bare host names") -}}
{{- end -}}
{{- range $mcp.allowedHosts -}}
{{- if not (regexMatch $hostPattern (trimPrefix "*." (toString .))) -}}
{{- fail (print $prefix "infinityRulesMcp.allowedHosts must list bare host names") -}}
{{- end -}}
{{- end -}}

{{- /* V4 */ -}}
{{- if not (regexMatch `^https://(\[[0-9A-Fa-f:.]+\]|[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?)(:[0-9]{1,5})?/?$` (toString $auth.mcpBaseUrl)) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.mcpBaseUrl must be an https URL without a path") -}}
{{- end -}}

{{- /* V5 */ -}}
{{- if not (has $auth.tokenMode (list "passthrough" "exchange")) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.tokenMode must be passthrough or exchange") -}}
{{- end -}}

{{- /* V6 */ -}}
{{- if $mcp.ingress.enabled -}}
{{- $publicHost := include "infinityRulesMcp.publicHost" . -}}
{{- if or (not (regexMatch $hostPattern $publicHost)) (regexMatch `^[0-9.]+$` $publicHost) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.mcpBaseUrl must use a DNS name when ingress is enabled") -}}
{{- end -}}
{{- end -}}

{{- /* V14: also rejects unquoted numeric IDs, which YAML turns into imprecise floating-point numbers. */ -}}
{{- $audiences := $auth.passthrough.audiences -}}
{{- if and $audiences (not (kindIs "slice" $audiences)) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.passthrough.audiences must be a list") -}}
{{- end -}}
{{- range $audiences -}}
{{- if kindIs "float64" . -}}
{{- fail (print $prefix "infinityRulesMcp.auth.passthrough.audiences must be quoted strings, for example \"26383066777993846573\"") -}}
{{- end -}}
{{- if or (kindIs "invalid" .) (not (regexMatch `^[^,\s]+$` (toString .))) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.passthrough.audiences must not contain commas or whitespace") -}}
{{- end -}}
{{- end -}}

{{- /* V12 */ -}}
{{- $trustStore := $mcp.trustStore -}}
{{- if or (not (has (toString $trustStore.type) (list "PKCS12" "JKS"))) (and $trustStore.passwordKey (not $trustStore.secretName)) (and $trustStore.secretName (not $trustStore.key)) -}}
{{- fail (print $prefix "infinityRulesMcp.trustStore is invalid") -}}
{{- end -}}

{{- /* V13: names are compared as Spring maps them to properties, so pega.inbound-auth matches PEGA_INBOUND_AUTH. */ -}}
{{- $managed := list "SERVER_PORT" "PEGA_SKILLS_PATH" "HTTP_PROXY_HOST" "HTTP_PROXY_PORT" "HTTP_PROXY_USERNAME" "HTTP_PROXY_PASSWORD" "NO_PROXY" "POD_NAMESPACE" "POD_NAME" "PEGA_INBOUND_AUTH" "PEGA_TOKEN_MODE" "MCP_PUBLIC_URL" "PEGA_ALLOWED_HOSTS" "SERVER_MAX_HTTP_REQUEST_HEADER_SIZE" "MCP_TRUSTSTORE_PASSWORD" "JDK_JAVA_OPTIONS" "PEGA_TOKEN_AUDIENCES" "PEGA_EXCHANGE_ISSUER" "PEGA_EXCHANGE_JWKS_URI" "PEGA_EXCHANGE_DISCOVERY" "PEGA_EXCHANGE_AUDIENCE" "PEGA_EXCHANGE_ALGS" "PEGA_EXCHANGE_CLIENT_ID" "PEGA_EXCHANGE_CLIENT_AUTH_METHOD" "PEGA_EXCHANGE_CLIENT_SECRET" "PEGA_EXCHANGE_INTERNAL_AUDIENCES" "PEGA_BASE_URL" "PEGA_ALLOW_INSECURE_HTTP" "PEGA_HTTP_TRUST_ALL_TLS" "SPRING_PROFILES_ACTIVE" "JAVA_TOOL_OPTIONS" "SPRING_APPLICATION_JSON" -}}
{{- if and $mcp.extraEnv (not (kindIs "slice" $mcp.extraEnv)) -}}
{{- fail (print $prefix "infinityRulesMcp.extraEnv must be a list of name and value entries") -}}
{{- end -}}
{{- range $mcp.extraEnv -}}
{{- if not (and (kindIs "map" .) .name) -}}
{{- fail (print $prefix "infinityRulesMcp.extraEnv must be a list of name and value entries") -}}
{{- end -}}
{{- if has (toString .name | upper | replace "." "_" | replace "-" "_") $managed -}}
{{- fail (printf "%sinfinityRulesMcp.extraEnv %s is managed by the chart" $prefix .name) -}}
{{- end -}}
{{- end -}}

{{- if eq $auth.tokenMode "exchange" -}}
{{- $exchange := $auth.exchange -}}

{{- /* V7 */ -}}
{{- $issuer := toString $exchange.issuer -}}
{{- if or (not (regexMatch `^https://[^/?#@\s]+(/[^?#\s]*)?$` $issuer)) (not (urlParse $issuer).hostname) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.exchange.issuer must be an https URL") -}}
{{- end -}}

{{- /* V8 */ -}}
{{- if kindIs "float64" $exchange.clientId -}}
{{- fail (print $prefix "infinityRulesMcp.auth.exchange.clientId must be a quoted string") -}}
{{- end -}}
{{- if or (kindIs "invalid" $exchange.clientId) (not (trim (toString $exchange.clientId))) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.exchange.clientId is required") -}}
{{- end -}}

{{- /* V9 */ -}}
{{- with $exchange.jwksUri -}}
{{- if or (not (regexMatch `^https://[^/?#@\s]+([/?][^#\s]*)?$` (toString .))) (not (urlParse (toString .)).hostname) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.exchange.jwksUri must be an https URL") -}}
{{- end -}}
{{- end -}}

{{- /* V10 */ -}}
{{- if or (not $exchange.algorithms) (not (kindIs "slice" $exchange.algorithms)) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.exchange.algorithms is invalid") -}}
{{- end -}}
{{- range $exchange.algorithms -}}
{{- if not (regexMatch `^(RS|PS|ES)(256|384|512)$` (toString .)) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.exchange.algorithms is invalid") -}}
{{- end -}}
{{- end -}}
{{- if not (has $exchange.clientAuthMethod (list "client_secret_post" "client_secret_basic")) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.exchange.clientAuthMethod is invalid") -}}
{{- end -}}
{{- if and $exchange.clientSecret.name (not $exchange.clientSecret.key) -}}
{{- fail (print $prefix "infinityRulesMcp.auth.exchange.clientSecret.key is invalid") -}}
{{- end -}}
{{- end -}}

{{- end -}}
{{- end -}}
