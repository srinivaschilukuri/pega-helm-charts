package backingservices

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

const (
	mcpName                 = "infinity-rules-mcp-service"
	mcpDeploymentTemplate   = "charts/autopilot/templates/infinity-rules-mcp-deployment.yaml"
	mcpServiceTemplate      = "charts/autopilot/templates/infinity-rules-mcp-service.yaml"
	mcpPdbTemplate          = "charts/autopilot/templates/infinity-rules-mcp-poddisruptionbudget.yaml"
	mcpIngressTemplate      = "charts/autopilot/templates/infinity-rules-mcp-ingress.yaml"
	mcpTrustedHeaderFixture = "data/infinity-rules-mcp-trusted-headers.yaml"
	autopilotClientID       = "26383066777993846573"
)

func mcpValues(options map[string]string) map[string]string {
	values := map[string]string{
		"autopilot.enabled":                  "true",
		"autopilot.infinityRulesMcp.enabled": "true",
	}
	for key, value := range options {
		values[key] = value
	}
	return values
}

func mcpBearerValues(options map[string]string) map[string]string {
	values := mcpValues(map[string]string{
		"autopilot.docker.infinityRulesMcp.image":    "registry.example.com/infinity-rules-skills-mcp:0.7.17",
		"autopilot.infinityRulesMcp.auth.enabled":    "true",
		"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "https://mcp.example.com",
		"autopilot.infinityRulesMcp.allowedHosts":    "{pega.example.com,*.pega.example.net}",
	})
	for key, value := range options {
		values[key] = value
	}
	return values
}

func mcpExchangeValues(options map[string]string) map[string]string {
	values := mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.auth.tokenMode":         "exchange",
		"autopilot.infinityRulesMcp.auth.exchange.issuer":   "https://login.example.com/tenant/v2.0",
		"autopilot.infinityRulesMcp.auth.exchange.clientId": "pega-exchange-client",
	})
	for key, value := range options {
		values[key] = value
	}
	return values
}

func renderMcpE(t *testing.T, values map[string]string, valuesFiles []string, templates ...string) (string, error) {
	chartPath, err := filepath.Abs(helmChartRelativePath)
	require.NoError(t, err)
	options := &helm.Options{SetValues: values, ValuesFiles: valuesFiles, Logger: logger.Discard}
	return helm.RenderTemplateE(t, options, chartPath, srsHelmRelease, templates)
}

func renderMcp(t *testing.T, values map[string]string, templates ...string) string {
	output, err := renderMcpE(t, values, nil, templates...)
	require.NoError(t, err)
	return output
}

// mcpValuesFile writes infinityRulesMcp values to a file. Helm's --set cannot express an empty list.
func mcpValuesFile(t *testing.T, infinityRulesMcpYAML string) string {
	path := filepath.Join(t.TempDir(), "values.yaml")
	indented := "    " + strings.ReplaceAll(strings.TrimSpace(infinityRulesMcpYAML), "\n", "\n    ")
	require.NoError(t, os.WriteFile(path, []byte("autopilot:\n  infinityRulesMcp:\n"+indented+"\n"), 0o600))
	return path
}

func without(values map[string]string, keys ...string) map[string]string {
	result := map[string]string{}
	for key, value := range values {
		result[key] = value
	}
	for _, key := range keys {
		delete(result, key)
	}
	return result
}

func mcpResources(t *testing.T, output string) map[string]map[string]interface{} {
	resources := map[string]map[string]interface{}{}
	for _, document := range strings.Split(output, "\n---") {
		var resource map[string]interface{}
		helm.UnmarshalK8SYaml(t, document, &resource)
		if resource == nil {
			continue
		}
		metadata := resource["metadata"].(map[string]interface{})
		resources[fmt.Sprintf("%s/%s", resource["kind"], metadata["name"])] = resource
	}
	return resources
}

func mcpDeployment(t *testing.T, values map[string]string) appsv1.Deployment {
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, renderMcp(t, values, mcpDeploymentTemplate), &deployment)
	require.Equal(t, mcpName, deployment.Name)
	return deployment
}

func mcpContainer(t *testing.T, values map[string]string) corev1.Container {
	deployment := mcpDeployment(t, values)
	require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	return deployment.Spec.Template.Spec.Containers[0]
}

func mcpIngress(t *testing.T, values map[string]string) networkingv1.Ingress {
	var ingress networkingv1.Ingress
	helm.UnmarshalK8SYaml(t, renderMcp(t, values, mcpIngressTemplate), &ingress)
	return ingress
}

func envValue(name, value string) corev1.EnvVar {
	return corev1.EnvVar{Name: name, Value: value}
}

func envField(name, fieldPath string) corev1.EnvVar {
	return corev1.EnvVar{Name: name, ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: fieldPath}}}
}

func envSecret(name, secret, key string) corev1.EnvVar {
	return corev1.EnvVar{Name: name, ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{Name: secret}, Key: key}}}
}

func envNames(container corev1.Container) []string {
	var names []string
	for _, env := range container.Env {
		names = append(names, env.Name)
	}
	return names
}

// mcpCommonBearerEnv is the environment shared by both token modes with mcpBearerValues.
func mcpCommonBearerEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		envValue("SERVER_PORT", "8090"),
		envValue("PEGA_SKILLS_PATH", "/opt/infinity-skills"),
		envField("POD_NAMESPACE", "metadata.namespace"),
		envField("POD_NAME", "metadata.name"),
		envValue("PEGA_INBOUND_AUTH", "bearer"),
		envValue("PEGA_TOKEN_MODE", "passthrough"),
		envValue("MCP_PUBLIC_URL", "https://mcp.example.com"),
		envValue("PEGA_ALLOWED_HOSTS", "pega.example.com,*.pega.example.net"),
		envValue("SERVER_MAX_HTTP_REQUEST_HEADER_SIZE", "16KB"),
	}
}

func imagePullSecretNames(deployment appsv1.Deployment) []string {
	var names []string
	for _, secret := range deployment.Spec.Template.Spec.ImagePullSecrets {
		names = append(names, secret.Name)
	}
	return names
}

// --- Trusted-header mode -------------------------------------------------------------------------

func TestInfinityRulesMcpTrustedHeaderModeMatchesMasterFixture(t *testing.T) {
	t.Parallel()
	fixture, err := os.ReadFile(mcpTrustedHeaderFixture)
	require.NoError(t, err)
	expected := mcpResources(t, string(fixture))
	require.Len(t, expected, 3)

	for name, values := range map[string]map[string]string{
		"defaults": mcpValues(nil),
		"bearer settings without auth.enabled": mcpValues(map[string]string{
			"autopilot.infinityRulesMcp.auth.enabled":           "false",
			"autopilot.infinityRulesMcp.auth.mcpBaseUrl":        "https://mcp.example.com",
			"autopilot.infinityRulesMcp.auth.tokenMode":         "exchange",
			"autopilot.infinityRulesMcp.allowedHosts":           "{pega.example.com}",
			"autopilot.infinityRulesMcp.infinityVersion":        "26-1",
			"autopilot.infinityRulesMcp.trustStore.secretName":  "truststore",
			"autopilot.infinityRulesMcp.trustStore.passwordKey": "password",
			"autopilot.infinityRulesMcp.extraEnv[0].name":       "EXTRA",
			"autopilot.infinityRulesMcp.extraEnv[0].value":      "value",
		}),
	} {
		output := renderMcp(t, values, mcpDeploymentTemplate, mcpServiceTemplate, mcpPdbTemplate)
		require.Equal(t, expected, mcpResources(t, output), name)
	}
}

func TestInfinityRulesMcpIngressRequiresBearerAndIngressFlags(t *testing.T) {
	t.Parallel()
	ingressSource := "# Source: backingservices/charts/autopilot/templates/infinity-rules-mcp-ingress.yaml"

	require.NotContains(t, renderMcp(t, mcpValues(nil)), ingressSource)
	require.NotContains(t, renderMcp(t, mcpBearerValues(nil)), ingressSource)
	require.Contains(t, renderMcp(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.ingress.enabled": "true",
	})), ingressSource)
}

// --- Image pull Secrets --------------------------------------------------------------------------

func TestInfinityRulesMcpDeploymentDefaultImagePullSecret(t *testing.T) {
	t.Parallel()
	for _, values := range []map[string]string{mcpValues(nil), mcpBearerValues(nil)} {
		require.Equal(t, []string{"autopilot-registry-secret"}, imagePullSecretNames(mcpDeployment(t, values)))
	}
}

func TestInfinityRulesMcpDeploymentAdditionalImagePullSecrets(t *testing.T) {
	t.Parallel()
	extra := map[string]string{"autopilot.docker.imagePullSecretNames": "{secret-a,secret-b}"}
	for _, values := range []map[string]string{mcpValues(extra), mcpBearerValues(extra)} {
		require.Equal(t, []string{"autopilot-registry-secret", "secret-a", "secret-b"}, imagePullSecretNames(mcpDeployment(t, values)))
	}
}

// --- Bearer mode environment ---------------------------------------------------------------------

func TestInfinityRulesMcpPassthroughEnvironment(t *testing.T) {
	t.Parallel()
	container := mcpContainer(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.auth.passthrough.audiences": "{ai-client-id}",
		"autopilot.infinityRulesMcp.auth.mcpBaseUrl":            "https://mcp.example.com/",
	}))

	expected := append(mcpCommonBearerEnv(), envValue("PEGA_TOKEN_AUDIENCES", autopilotClientID+",ai-client-id"))
	require.Equal(t, expected, container.Env)
}

func TestInfinityRulesMcpPassthroughDefaultAudience(t *testing.T) {
	t.Parallel()
	container := mcpContainer(t, mcpBearerValues(nil))
	require.Equal(t, append(mcpCommonBearerEnv(), envValue("PEGA_TOKEN_AUDIENCES", autopilotClientID)), container.Env)
}

func TestInfinityRulesMcpPassthroughDeduplicatesAutopilotClient(t *testing.T) {
	t.Parallel()
	container := mcpContainer(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.auth.passthrough.audiences": "{ai-client-id," + autopilotClientID + "}",
	}))
	require.Contains(t, container.Env, envValue("PEGA_TOKEN_AUDIENCES", autopilotClientID+",ai-client-id"))
}

func TestInfinityRulesMcpExchangeEnvironment(t *testing.T) {
	t.Parallel()
	container := mcpContainer(t, mcpExchangeValues(map[string]string{
		"autopilot.infinityRulesMcp.auth.exchange.jwksUri":           "https://login.example.com/tenant/discovery/v2.0/keys",
		"autopilot.infinityRulesMcp.auth.exchange.audience":          "api://mcp-app-id",
		"autopilot.infinityRulesMcp.auth.exchange.algorithms":        "{RS256,PS256}",
		"autopilot.infinityRulesMcp.auth.exchange.clientAuthMethod":  "client_secret_basic",
		"autopilot.infinityRulesMcp.auth.exchange.clientSecret.name": "mcp-exchange-client",
		"autopilot.infinityRulesMcp.auth.exchange.clientSecret.key":  "secret",
	}))

	expected := mcpCommonBearerEnv()
	expected[5] = envValue("PEGA_TOKEN_MODE", "exchange")
	expected = append(expected,
		envValue("PEGA_EXCHANGE_ISSUER", "https://login.example.com/tenant/v2.0"),
		envValue("PEGA_EXCHANGE_JWKS_URI", "https://login.example.com/tenant/discovery/v2.0/keys"),
		envValue("PEGA_EXCHANGE_DISCOVERY", "false"),
		envValue("PEGA_EXCHANGE_AUDIENCE", "api://mcp-app-id"),
		envValue("PEGA_EXCHANGE_ALGS", "RS256,PS256"),
		envValue("PEGA_EXCHANGE_CLIENT_ID", "pega-exchange-client"),
		envValue("PEGA_EXCHANGE_CLIENT_AUTH_METHOD", "client_secret_basic"),
		envSecret("PEGA_EXCHANGE_CLIENT_SECRET", "mcp-exchange-client", "secret"),
		envValue("PEGA_EXCHANGE_INTERNAL_AUDIENCES", autopilotClientID),
	)
	require.Equal(t, expected, container.Env)
}

func TestInfinityRulesMcpExchangeMinimalEnvironment(t *testing.T) {
	t.Parallel()
	container := mcpContainer(t, mcpExchangeValues(nil))

	expected := mcpCommonBearerEnv()
	expected[5] = envValue("PEGA_TOKEN_MODE", "exchange")
	expected = append(expected,
		envValue("PEGA_EXCHANGE_ISSUER", "https://login.example.com/tenant/v2.0"),
		envValue("PEGA_EXCHANGE_ALGS", "RS256"),
		envValue("PEGA_EXCHANGE_CLIENT_ID", "pega-exchange-client"),
		envValue("PEGA_EXCHANGE_CLIENT_AUTH_METHOD", "client_secret_post"),
		envValue("PEGA_EXCHANGE_INTERNAL_AUDIENCES", autopilotClientID),
	)
	require.Equal(t, expected, container.Env)
}

func TestInfinityRulesMcpExchangeInternalAudiencesAreNotConfigurable(t *testing.T) {
	t.Parallel()
	container := mcpContainer(t, mcpExchangeValues(map[string]string{
		"autopilot.infinityRulesMcp.auth.exchange.internalAudiences": "{user-client-id}",
		"autopilot.infinityRulesMcp.auth.passthrough.audiences":      "{ai-client-id}",
	}))
	require.Contains(t, container.Env, envValue("PEGA_EXCHANGE_INTERNAL_AUDIENCES", autopilotClientID))
	require.NotContains(t, envNames(container), "PEGA_TOKEN_AUDIENCES")
}

func TestInfinityRulesMcpBearerNeverRendersUnsafeVariables(t *testing.T) {
	t.Parallel()
	for _, values := range []map[string]string{mcpBearerValues(nil), mcpExchangeValues(nil)} {
		names := envNames(mcpContainer(t, values))
		for _, name := range []string{"PEGA_WRITE_ENABLED", "PEGA_ALLOW_INSECURE_HTTP", "PEGA_HTTP_TRUST_ALL_TLS",
			"PEGA_BASE_URL", "SPRING_PROFILES_ACTIVE", "JAVA_TOOL_OPTIONS", "JDK_JAVA_OPTIONS"} {
			require.NotContains(t, names, name)
		}
	}
	trustedHeaderNames := envNames(mcpContainer(t, mcpValues(nil)))
	require.Contains(t, trustedHeaderNames, "PEGA_WRITE_ENABLED")
	require.NotContains(t, trustedHeaderNames, "PEGA_INBOUND_AUTH")
}

func TestInfinityRulesMcpBearerOptionalEnvironment(t *testing.T) {
	t.Parallel()
	container := mcpContainer(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.proxy.httpProxyHost": "proxy.example.com",
		"autopilot.infinityRulesMcp.proxy.httpProxyPort": "3128",
		"autopilot.infinityRulesMcp.extraEnv[0].name":    "PEGA_EXCHANGE_CACHE_TTL",
		"autopilot.infinityRulesMcp.extraEnv[0].value":   "10m",
		"autopilot.infinityRulesMcp.extraEnv[1].name":    "PEGA_OAUTH_CLIENT_ID",
		"autopilot.infinityRulesMcp.extraEnv[1].value":   "12345",
		"autopilot.infinityRulesMcp.extraEnv[2].name":    "PEGA_KNOWLEDGE_ENABLED",
		"autopilot.infinityRulesMcp.extraEnv[2].value":   "true",
		"autopilot.infinityRulesMcp.extraEnv[3].name":    "PEGA_INFINITY_VERSION",
		"autopilot.infinityRulesMcp.extraEnv[3].value":   "26-1",
	}))

	require.Equal(t, []corev1.EnvVar{
		envValue("SERVER_PORT", "8090"),
		envValue("PEGA_SKILLS_PATH", "/opt/infinity-skills"),
		envValue("HTTP_PROXY_HOST", "proxy.example.com"),
		envValue("HTTP_PROXY_PORT", "3128"),
		envField("POD_NAMESPACE", "metadata.namespace"),
		envField("POD_NAME", "metadata.name"),
		envValue("PEGA_INBOUND_AUTH", "bearer"),
		envValue("PEGA_TOKEN_MODE", "passthrough"),
		envValue("MCP_PUBLIC_URL", "https://mcp.example.com"),
		envValue("PEGA_ALLOWED_HOSTS", "pega.example.com,*.pega.example.net"),
		envValue("SERVER_MAX_HTTP_REQUEST_HEADER_SIZE", "16KB"),
		envValue("PEGA_TOKEN_AUDIENCES", autopilotClientID),
		envValue("PEGA_EXCHANGE_CACHE_TTL", "10m"),
		envValue("PEGA_OAUTH_CLIENT_ID", "12345"),
		envValue("PEGA_KNOWLEDGE_ENABLED", "true"),
		envValue("PEGA_INFINITY_VERSION", "26-1"),
	}, container.Env)
}

func TestInfinityRulesMcpInfinityVersionUsesServiceDefault(t *testing.T) {
	t.Parallel()
	options := map[string]string{"autopilot.infinityRulesMcp.infinityVersion": "26-1"}
	for _, values := range []map[string]string{mcpBearerValues(options), mcpExchangeValues(options)} {
		require.NotContains(t, envNames(mcpContainer(t, values)), "PEGA_INFINITY_VERSION")
	}
}

func TestInfinityRulesMcpKnowledgeUsesServiceDefault(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"true", "false"} {
		container := mcpContainer(t, mcpBearerValues(map[string]string{
			"autopilot.infinityRulesMcp.knowledge.enabled": value,
		}))
		require.NotContains(t, envNames(container), "PEGA_KNOWLEDGE_ENABLED")
	}

	container := mcpContainer(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.extraEnv[0].name":  "PEGA_KNOWLEDGE_ENABLED",
		"autopilot.infinityRulesMcp.extraEnv[0].value": "false",
	}))
	require.Contains(t, container.Env, envValue("PEGA_KNOWLEDGE_ENABLED", "false"))
}

// --- Probes and identity -------------------------------------------------------------------------

func TestInfinityRulesMcpProbes(t *testing.T) {
	t.Parallel()
	trustedHeader := mcpContainer(t, mcpValues(nil))
	require.Nil(t, trustedHeader.LivenessProbe.HTTPGet)
	require.Nil(t, trustedHeader.ReadinessProbe.HTTPGet)
	require.Equal(t, 8090, trustedHeader.LivenessProbe.TCPSocket.Port.IntValue())
	require.Equal(t, 8090, trustedHeader.ReadinessProbe.TCPSocket.Port.IntValue())

	bearer := mcpContainer(t, mcpBearerValues(nil))
	require.Nil(t, bearer.LivenessProbe.TCPSocket)
	require.Nil(t, bearer.ReadinessProbe.TCPSocket)
	require.Equal(t, "/actuator/health/liveness", bearer.LivenessProbe.HTTPGet.Path)
	require.Equal(t, 8090, bearer.LivenessProbe.HTTPGet.Port.IntValue())
	require.Equal(t, "/actuator/health/readiness", bearer.ReadinessProbe.HTTPGet.Path)
	require.Equal(t, 8090, bearer.ReadinessProbe.HTTPGet.Port.IntValue())

	for _, probes := range [][2]*corev1.Probe{
		{trustedHeader.LivenessProbe, bearer.LivenessProbe},
		{trustedHeader.ReadinessProbe, bearer.ReadinessProbe},
	} {
		require.Equal(t, probes[0].InitialDelaySeconds, probes[1].InitialDelaySeconds)
		require.Equal(t, probes[0].TimeoutSeconds, probes[1].TimeoutSeconds)
		require.Equal(t, probes[0].PeriodSeconds, probes[1].PeriodSeconds)
		require.Equal(t, probes[0].FailureThreshold, probes[1].FailureThreshold)
	}
}

func TestInfinityRulesMcpBearerKeepsResourceIdentity(t *testing.T) {
	t.Parallel()
	templates := []string{mcpDeploymentTemplate, mcpServiceTemplate, mcpPdbTemplate}
	trustedHeader := mcpResources(t, renderMcp(t, mcpValues(nil), templates...))
	bearer := mcpResources(t, renderMcp(t, mcpBearerValues(nil), templates...))

	require.Equal(t, trustedHeader["Service/"+mcpName], bearer["Service/"+mcpName])
	require.Equal(t, trustedHeader["PodDisruptionBudget/"+mcpName], bearer["PodDisruptionBudget/"+mcpName])

	trustedDeployment := mcpDeployment(t, mcpValues(nil))
	bearerDeployment := mcpDeployment(t, mcpBearerValues(nil))
	require.Equal(t, trustedDeployment.ObjectMeta, bearerDeployment.ObjectMeta)
	require.Equal(t, trustedDeployment.Spec.Selector, bearerDeployment.Spec.Selector)
	require.Equal(t, trustedDeployment.Spec.Replicas, bearerDeployment.Spec.Replicas)
	require.Equal(t, trustedDeployment.Spec.Template.ObjectMeta, bearerDeployment.Spec.Template.ObjectMeta)
	require.Equal(t, trustedDeployment.Spec.Template.Spec.SecurityContext, bearerDeployment.Spec.Template.Spec.SecurityContext)
	require.Equal(t, trustedDeployment.Spec.Template.Spec.Volumes, bearerDeployment.Spec.Template.Spec.Volumes)

	trustedContainer := trustedDeployment.Spec.Template.Spec.Containers[0]
	bearerContainer := bearerDeployment.Spec.Template.Spec.Containers[0]
	require.Equal(t, trustedContainer.Name, bearerContainer.Name)
	require.Equal(t, trustedContainer.Ports, bearerContainer.Ports)
	require.Equal(t, trustedContainer.Resources, bearerContainer.Resources)
	require.Equal(t, trustedContainer.SecurityContext, bearerContainer.SecurityContext)
	require.Equal(t, trustedContainer.VolumeMounts, bearerContainer.VolumeMounts)
}

// --- Ingress -------------------------------------------------------------------------------------

func TestInfinityRulesMcpIngress(t *testing.T) {
	t.Parallel()
	ingress := mcpIngress(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.auth.mcpBaseUrl":                       "https://MCP.Example.com:8443/",
		"autopilot.infinityRulesMcp.ingress.enabled":                       "true",
		"autopilot.infinityRulesMcp.ingress.ingressClassName":              "nginx",
		"autopilot.infinityRulesMcp.ingress.annotations.example\\.com/key": "value",
		"autopilot.infinityRulesMcp.ingress.tls.secretName":                "mcp-tls",
		"autopilot.infinityRulesMcp.service.port":                          "8080",
	}))

	require.Equal(t, mcpName, ingress.Name)
	require.Equal(t, map[string]string{"app": mcpName}, ingress.Labels)
	require.Equal(t, map[string]string{"example.com/key": "value"}, ingress.Annotations)
	require.Equal(t, "nginx", *ingress.Spec.IngressClassName)
	require.Len(t, ingress.Spec.Rules, 1)
	rule := ingress.Spec.Rules[0]
	require.Equal(t, "mcp.example.com", rule.Host)

	exact := networkingv1.PathTypeExact
	prefix := networkingv1.PathTypePrefix
	backend := networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{
		Name: mcpName, Port: networkingv1.ServiceBackendPort{Number: 8080}}}
	require.Equal(t, []networkingv1.HTTPIngressPath{
		{Path: "/mcp", PathType: &exact, Backend: backend},
		{Path: "/.well-known/oauth-protected-resource", PathType: &prefix, Backend: backend},
	}, rule.HTTP.Paths)
	require.Equal(t, []networkingv1.IngressTLS{{Hosts: []string{"mcp.example.com"}, SecretName: "mcp-tls"}}, ingress.Spec.TLS)
}

func TestInfinityRulesMcpIngressWithoutTLSClassOrAnnotations(t *testing.T) {
	t.Parallel()
	ingress := mcpIngress(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.ingress.enabled": "true",
	}))
	require.Empty(t, ingress.Spec.TLS)
	require.Nil(t, ingress.Spec.IngressClassName)
	require.Empty(t, ingress.Annotations)
	require.Equal(t, "mcp.example.com", ingress.Spec.Rules[0].Host)
}

// --- Trust store ---------------------------------------------------------------------------------

func TestInfinityRulesMcpTrustStoreWithPassword(t *testing.T) {
	t.Parallel()
	deployment := mcpDeployment(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.trustStore.secretName":  "pega-ca",
		"autopilot.infinityRulesMcp.trustStore.key":         "ca.p12",
		"autopilot.infinityRulesMcp.trustStore.passwordKey": "password",
	}))
	container := deployment.Spec.Template.Spec.Containers[0]

	require.Contains(t, deployment.Spec.Template.Spec.Volumes, corev1.Volume{Name: "mcp-truststore", VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{SecretName: "pega-ca", Items: []corev1.KeyToPath{{Key: "ca.p12", Path: "ca.p12"}}}}})
	require.Contains(t, container.VolumeMounts, corev1.VolumeMount{Name: "mcp-truststore", MountPath: "/opt/mcp/truststore", ReadOnly: true})

	names := envNames(container)
	require.Equal(t, "JDK_JAVA_OPTIONS", names[len(names)-1])
	require.Equal(t, "MCP_TRUSTSTORE_PASSWORD", names[len(names)-2])
	require.Equal(t, envSecret("MCP_TRUSTSTORE_PASSWORD", "pega-ca", "password"), container.Env[len(names)-2])
	require.Equal(t, envValue("JDK_JAVA_OPTIONS", "-Djavax.net.ssl.trustStore=/opt/mcp/truststore/ca.p12 "+
		"-Djavax.net.ssl.trustStoreType=PKCS12 -Djavax.net.ssl.trustStorePassword=$(MCP_TRUSTSTORE_PASSWORD)"),
		container.Env[len(names)-1])
}

func TestInfinityRulesMcpTrustStoreWithoutPassword(t *testing.T) {
	t.Parallel()
	container := mcpContainer(t, mcpBearerValues(map[string]string{
		"autopilot.infinityRulesMcp.trustStore.secretName": "pega-ca",
		"autopilot.infinityRulesMcp.trustStore.key":        "cacerts",
		"autopilot.infinityRulesMcp.trustStore.type":       "JKS",
	}))
	require.NotContains(t, envNames(container), "MCP_TRUSTSTORE_PASSWORD")
	require.Contains(t, container.Env, envValue("JDK_JAVA_OPTIONS",
		"-Djavax.net.ssl.trustStore=/opt/mcp/truststore/cacerts -Djavax.net.ssl.trustStoreType=JKS"))
}

// --- Validation ----------------------------------------------------------------------------------

func TestInfinityRulesMcpValidation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		id      string
		values  map[string]string
		message string
	}{
		{"V1", mcpValues(map[string]string{"autopilot.infinityRulesMcp.ingress.enabled": "true"}),
			"infinityRulesMcp.ingress requires infinityRulesMcp.auth.enabled; trusted-header mode must not be exposed"},
		{"V2 placeholder", mcpBearerValues(map[string]string{"autopilot.docker.infinityRulesMcp.image": "YOUR_INFINITY_RULES_MCP_IMAGE:TAG"}),
			"docker.infinityRulesMcp.image is required"},
		{"V2 empty", mcpBearerValues(map[string]string{"autopilot.docker.infinityRulesMcp.image": ""}),
			"docker.infinityRulesMcp.image is required"},
		{"V3 empty", without(mcpBearerValues(nil), "autopilot.infinityRulesMcp.allowedHosts"),
			"infinityRulesMcp.allowedHosts must list bare host names"},
		{"V3 scheme", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.allowedHosts": "{https://pega.example.com}"}),
			"infinityRulesMcp.allowedHosts must list bare host names"},
		{"V3 port", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.allowedHosts": "{pega.example.com:443}"}),
			"infinityRulesMcp.allowedHosts must list bare host names"},
		{"V3 bare wildcard", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.allowedHosts": "{*}"}),
			"infinityRulesMcp.allowedHosts must list bare host names"},
		{"V4 empty", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.mcpBaseUrl": ""}),
			"infinityRulesMcp.auth.mcpBaseUrl must be an https URL without a path"},
		{"V4 old name", without(mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.publicUrl": "https://mcp.example.com"}), "autopilot.infinityRulesMcp.auth.mcpBaseUrl"),
			"infinityRulesMcp.auth.mcpBaseUrl must be an https URL without a path"},
		{"V4 http", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "http://mcp.example.com"}),
			"infinityRulesMcp.auth.mcpBaseUrl must be an https URL without a path"},
		{"V4 path", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "https://mcp.example.com/mcp"}),
			"infinityRulesMcp.auth.mcpBaseUrl must be an https URL without a path"},
		{"V4 credentials", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "https://user@mcp.example.com"}),
			"infinityRulesMcp.auth.mcpBaseUrl must be an https URL without a path"},
		{"V4 query", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "https://mcp.example.com?a=b"}),
			"infinityRulesMcp.auth.mcpBaseUrl must be an https URL without a path"},
		{"V5", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.tokenMode": "both"}),
			"infinityRulesMcp.auth.tokenMode must be passthrough or exchange"},
		{"V6 IPv4", mcpBearerValues(map[string]string{
			"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "https://10.0.0.1",
			"autopilot.infinityRulesMcp.ingress.enabled": "true",
		}), "infinityRulesMcp.auth.mcpBaseUrl must use a DNS name when ingress is enabled"},
		{"V6 IPv6", mcpBearerValues(map[string]string{
			"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "https://[fd00::1]",
			"autopilot.infinityRulesMcp.ingress.enabled": "true",
		}), "infinityRulesMcp.auth.mcpBaseUrl must use a DNS name when ingress is enabled"},
		{"V7 empty", mcpExchangeValues(map[string]string{"autopilot.infinityRulesMcp.auth.exchange.issuer": ""}),
			"infinityRulesMcp.auth.exchange.issuer must be an https URL"},
		{"V7 http", mcpExchangeValues(map[string]string{"autopilot.infinityRulesMcp.auth.exchange.issuer": "http://login.example.com"}),
			"infinityRulesMcp.auth.exchange.issuer must be an https URL"},
		{"V7 query", mcpExchangeValues(map[string]string{"autopilot.infinityRulesMcp.auth.exchange.issuer": "https://login.example.com/?a=b"}),
			"infinityRulesMcp.auth.exchange.issuer must be an https URL"},
		{"V8", mcpExchangeValues(map[string]string{"autopilot.infinityRulesMcp.auth.exchange.clientId": ""}),
			"infinityRulesMcp.auth.exchange.clientId is required"},
		{"V9 http", mcpExchangeValues(map[string]string{"autopilot.infinityRulesMcp.auth.exchange.jwksUri": "http://login.example.com/keys"}),
			"infinityRulesMcp.auth.exchange.jwksUri must be an https URL"},
		{"V9 fragment", mcpExchangeValues(map[string]string{"autopilot.infinityRulesMcp.auth.exchange.jwksUri": "https://login.example.com/keys#a"}),
			"infinityRulesMcp.auth.exchange.jwksUri must be an https URL"},
		{"V10 algorithm", mcpExchangeValues(map[string]string{"autopilot.infinityRulesMcp.auth.exchange.algorithms": "{HS256}"}),
			"infinityRulesMcp.auth.exchange.algorithms is invalid"},
		{"V10 client auth", mcpExchangeValues(map[string]string{"autopilot.infinityRulesMcp.auth.exchange.clientAuthMethod": "private_key_jwt"}),
			"infinityRulesMcp.auth.exchange.clientAuthMethod is invalid"},
		{"V12 type", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.trustStore.type": "PEM"}),
			"infinityRulesMcp.trustStore is invalid"},
		{"V12 password without secret", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.trustStore.passwordKey": "password"}),
			"infinityRulesMcp.trustStore is invalid"},
		{"V13", mcpBearerValues(map[string]string{
			"autopilot.infinityRulesMcp.extraEnv[0].name":  "PEGA_HTTP_TRUST_ALL_TLS",
			"autopilot.infinityRulesMcp.extraEnv[0].value": "true",
		}), "infinityRulesMcp.extraEnv PEGA_HTTP_TRUST_ALL_TLS is managed by the chart"},
		{"V13 relaxed name", mcpBearerValues(map[string]string{
			"autopilot.infinityRulesMcp.extraEnv[0].name":  "pega.inbound-auth",
			"autopilot.infinityRulesMcp.extraEnv[0].value": "trusted-headers",
		}), "infinityRulesMcp.extraEnv pega.inbound-auth is managed by the chart"},
		{"V13 exchange variable in passthrough", mcpBearerValues(map[string]string{
			"autopilot.infinityRulesMcp.extraEnv[0].name":  "PEGA_EXCHANGE_INTERNAL_AUDIENCES",
			"autopilot.infinityRulesMcp.extraEnv[0].value": "id",
		}), "infinityRulesMcp.extraEnv PEGA_EXCHANGE_INTERNAL_AUDIENCES is managed by the chart"},
		{"V14 whitespace", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.passthrough.audiences": "{bad id}"}),
			"infinityRulesMcp.auth.passthrough.audiences must not contain commas or whitespace"},
		{"V14 comma", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.passthrough.audiences[0]": "a\\,b"}),
			"infinityRulesMcp.auth.passthrough.audiences must not contain commas or whitespace"},
		{"V14 empty", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.passthrough.audiences[0]": ""}),
			"infinityRulesMcp.auth.passthrough.audiences must not contain commas or whitespace"},
		{"V14 not a list", mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.passthrough.audiences": "ai-client-id"}),
			"infinityRulesMcp.auth.passthrough.audiences must be a list"},
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.id, func(t *testing.T) {
			t.Parallel()
			_, err := renderMcpE(t, testCase.values, nil, mcpDeploymentTemplate)
			require.Error(t, err)
			require.Contains(t, err.Error(), "infinity-rules-mcp: "+testCase.message)
		})
	}
}

func TestInfinityRulesMcpValidationWithValuesFiles(t *testing.T) {
	t.Parallel()
	cases := []struct {
		id         string
		values     map[string]string
		valuesYAML string
		message    string
	}{
		{"V10 no algorithms", mcpExchangeValues(nil), "auth:\n  exchange:\n    algorithms: []",
			"infinityRulesMcp.auth.exchange.algorithms is invalid"},
		{"unquoted audience", mcpBearerValues(nil), "auth:\n  passthrough:\n    audiences:\n      - 26383066777993846573",
			"infinityRulesMcp.auth.passthrough.audiences must be quoted strings"},
		{"unquoted client ID", without(mcpExchangeValues(nil), "autopilot.infinityRulesMcp.auth.exchange.clientId"),
			"auth:\n  exchange:\n    clientId: 26383066777993846573",
			"infinityRulesMcp.auth.exchange.clientId must be a quoted string"},
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.id, func(t *testing.T) {
			t.Parallel()
			_, err := renderMcpE(t, testCase.values, []string{mcpValuesFile(t, testCase.valuesYAML)}, mcpDeploymentTemplate)
			require.Error(t, err)
			require.Contains(t, err.Error(), "infinity-rules-mcp: "+testCase.message)
		})
	}
}

func TestInfinityRulesMcpValidationAcceptsValidBearerValues(t *testing.T) {
	t.Parallel()
	for _, values := range []map[string]string{
		mcpBearerValues(map[string]string{
			"autopilot.infinityRulesMcp.ingress.enabled": "true",
			"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "https://mcp.example.com:8443/",
		}),
		mcpBearerValues(map[string]string{"autopilot.infinityRulesMcp.auth.mcpBaseUrl": "https://10.0.0.1"}),
		mcpExchangeValues(map[string]string{
			"autopilot.infinityRulesMcp.auth.exchange.jwksUri": "https://login.example.com/keys?appid=1",
			"autopilot.infinityRulesMcp.extraEnv[0].name":      "PEGA_OAUTH_CLIENT_ID",
			"autopilot.infinityRulesMcp.extraEnv[0].value":     "ai-client-id",
		}),
	} {
		_, err := renderMcpE(t, values, nil, mcpDeploymentTemplate)
		require.NoError(t, err)
	}
}
