package v2alpha1

// +k8s:deepcopy-gen=package

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories={dolphin},singular="dolphingatewayclassconfig",path="dolphingatewayclassconfigs",scope="Namespaced",shortName={cgcc}
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.spec.description`,priority=1
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type DolphinGatewayClassConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   DolphinGatewayClassConfigSpec   `json:"spec,omitempty"`
	Status DolphinGatewayClassConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=false
// +deepequal-gen=false

// DolphinGatewayClassConfigList is a list of
// DolphinGatewayClassConfig objects.
type DolphinGatewayClassConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DolphinGatewayClassConfig `json:"items"`
}

// +deepequal-gen=true

type LoadBalancerSourceRangesPolicyType string

const (
	LoadBalancerSourceRangesPolicyAllow LoadBalancerSourceRangesPolicyType = "Allow"

	// LoadBalancerSourceRangesPolicyDeny denies traffic for the given source ranges.
	LoadBalancerSourceRangesPolicyDeny LoadBalancerSourceRangesPolicyType = "Deny"
)

type ServiceConfig struct {
	Type                  corev1.ServiceType                  `json:"type,omitempty"`
	ExternalTrafficPolicy corev1.ServiceExternalTrafficPolicy `json:"externalTrafficPolicy,omitempty"`
	LoadBalancerClass     *string                             `json:"loadBalancerClass,omitempty"`
	IPFamilies            []corev1.IPFamily                   `json:"ipFamilies,omitempty"`
	IPFamilyPolicy        *corev1.IPFamilyPolicy              `json:"ipFamilyPolicy,omitempty"`

	AllocateLoadBalancerNodePorts  *bool                              `json:"allocateLoadBalancerNodePorts,omitempty"`
	LoadBalancerSourceRanges       []string                           `json:"loadBalancerSourceRanges,omitempty"`
	LoadBalancerSourceRangesPolicy LoadBalancerSourceRangesPolicyType `json:"loadBalancerSourceRangesPolicy,omitempty"`
	TrafficDistribution            *string                            `json:"trafficDistribution,omitempty"`
}

type Telemetry struct {
	AccessLogs []AccessLogs `json:"accessLogs,omitempty"`
}

type AccessLogs struct {
	// Format specifies the access log output format.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=JSON;Text
	Format AccessLogsFormat `json:"format"`
	// JSON maps access log field names to Envoy command operators.
	// It is used when Format is "JSON".
	// For available format specifiers, see the Envoy documentation:
	// - https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage
	// Note: Always refer to the documentation matching the specific Envoy version you are running.
	// The following Dolphin-specific formatters are also supported:
	// - %DOLPHIN_GATEWAY_NAME% -- replaced with the Gateway resource name.
	// - %DOLPHIN_GATEWAY_NAMESPACE% -- replaced with the Gateway resource namespace.
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinProperties=1
	// +kubebuilder:validation:MaxProperties=64
	// +kubebuilder:default={start_time:"%START_TIME%",method:"%REQUEST_HEADER(:METHOD)%",path:"%REQUEST_HEADER(X-ENVOY-ORIGINAL-PATH?:PATH)%",protocol:"%PROTOCOL%",response_code:"%RESPONSE_CODE%",response_flags:"%RESPONSE_FLAGS%",bytes_received:"%BYTES_RECEIVED%",bytes_sent:"%BYTES_SENT%",duration:"%DURATION%",upstream_service_time:"%RESPONSE_HEADER(X-ENVOY-UPSTREAM-SERVICE-TIME)%",x_forwarded_for:"%REQUEST_HEADER(X-FORWARDED-FOR)%",user_agent:"%REQUEST_HEADER(USER-AGENT)%",request_id:"%REQUEST_HEADER(X-REQUEST-ID)%",authority:"%REQUEST_HEADER(:AUTHORITY)%",upstream_host:"%UPSTREAM_HOST%"}
	JSON map[string]string `json:"json,omitempty"`
	// Text specifies the Envoy access log format string.
	// It is used when Format is "Text".
	// For available format specifiers, see the Envoy documentation:
	// - https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage
	// Note: Always refer to the documentation matching the specific Envoy version you are running.
	// The following Dolphin-specific formatters are also supported:
	// - %DOLPHIN_GATEWAY_NAME% -- replaced with the Gateway resource name.
	// - %DOLPHIN_GATEWAY_NAMESPACE% -- replaced with the Gateway resource namespace.
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	// +kubebuilder:default="[%START_TIME%] \"%REQUEST_HEADER(:METHOD)% %REQUEST_HEADER(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESPONSE_HEADER(X-ENVOY-UPSTREAM-SERVICE-TIME)% \"%REQUEST_HEADER(X-FORWARDED-FOR)%\" \"%REQUEST_HEADER(USER-AGENT)%\" \"%REQUEST_HEADER(X-REQUEST-ID)%\" \"%REQUEST_HEADER(:AUTHORITY)%\" \"%UPSTREAM_HOST%\""
	Text string `json:"text,omitempty"`
	// Targets specifies the generated Envoy proxy components where access logs
	// are emitted. If omitted, access logs are emitted for HTTP traffic only.
	// HTTP targets Envoy HTTP connection managers. TCP targets Envoy TCP proxies,
	// including TLS passthrough.
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:default={HTTP}
	// +listType=set
	Targets []AccessLogsTarget `json:"targets,omitempty"`
}

type AccessLogsFormat string

const (
	AccessLogsFormatJSON AccessLogsFormat = "JSON"
	AccessLogsFormatText AccessLogsFormat = "Text"
)

type AccessLogsTarget string

const (
	AccessLogsTargetHTTP AccessLogsTarget = "HTTP"
	// AccessLogsTargetTCP emits access logs from Envoy TCP proxies, including TLS passthrough.
	AccessLogsTargetTCP AccessLogsTarget = "TCP"
)

type DolphinGatewayClassConfigSpec struct {
	Description *string        `json:"description,omitempty"`
	Service     *ServiceConfig `json:"service,omitempty"`
	Telemetry   *Telemetry     `json:"telemetry,omitempty"`
}

type DolphinGatewayClassConfigStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (c *DolphinGatewayClassConfig) IsTelemetryConfigured() bool {
	return c != nil &&
		c.Spec.Telemetry != nil
}

func (t *Telemetry) IsAccessLogsConfigured() bool {
	return t != nil && len(t.AccessLogs) > 0
}
