package dolphinio

const (
	LabelPrefix                         = "io.dolphin.k8s"
	CtrlPrefixPolicyStatus              = "sync-cnp-policy-status"
	BatchJobControllerUID               = "batch.kubernetes.io/controller-uid"
	DolphinIdentityAnnotationDeprecated = "dolphin-identity"
	PolicyLabelName                     = LabelPrefix + ".policy.name"
	PolicyLabelUID                      = LabelPrefix + ".policy.uid"
	PolicyLabelNamespace                = LabelPrefix + ".policy.namespace"
	PolicyLabelDerivedFrom              = LabelPrefix + ".policy.derived-from"
	PolicyLabelServiceAccount           = LabelPrefix + ".policy.serviceaccount"
	PolicyLabelCluster                  = LabelPrefix + ".policy.cluster"
	PolicyLabelIstioSidecarProxy        = LabelPrefix + ".policy.istiosidecarproxy"
	PodNamespaceMetaLabels              = LabelPrefix + ".namespace.labels"
	LabelMetadataName                   = "kubernetes.io/metadata.name"
	PodNamespaceLabel                   = "io.kubernetes.pod.namespace"
	PodNamespaceMetaNameLabel           = PodNamespaceMetaLabels + "." + LabelMetadataName
)
