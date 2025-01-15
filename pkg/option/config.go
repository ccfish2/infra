package option

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// myself
	"github.com/ccfish2/infra/pkg/defaults"
)

const (
	EnableK8s             = "enable-k8s"
	K8sAPIServer          = "k8s-api-server"
	K8sKubeConfigPath     = "k8s-kubeconfig-path"
	K8sServiceCacheSize   = "k8s-service-cache-size"
	K8sClientQPSLimit     = "k8s-client-qps"
	K8sClientBurst        = "k8s-client-burst"
	K8sHeartbeatTimeout   = "k8s-heartbeat-timeout"
	K8sEnableAPIDiscovery = "enable-k8s-api-discovery"

	IdentityAllocationModeCRD     = "crd"
	IdentityAllocationModeKVstore = "kvstore"
	MonitorAggregationName        = "monitor-aggregation"
	CTMapEntriesGlobalTCPName     = "bpf-ct-global-tcp-max"
	CTMapEntriesGlobalAnyName     = "bpf-ct-global-any-max"

	AllowICMPFragNeeded             = "allow-icmp-frag-needed"
	AnnotateK8sNode                 = "annotate-k8s-node"
	EnableIPv4Name                  = "enable-ipv4"
	EnableIPv6Name                  = "enable-ipv6"
	EnableIPv6NDPName               = "enable-ipv6-ndp"
	EnableSCTPName                  = "enable-sctp"
	EnableEndpointRoutes            = "enable-endpoint-routes"
	EnableHealthChecking            = "enable-health-checking"
	EnableEndpointHealthChecking    = "enable-endpoint-health-checking"
	EnableHealthCheckNodePort       = "enable-health-check-nodeport"
	EnableHealthCheckLoadBalancerIP = "enable-health-check-loadbalancer-ip"
	IdentityChangeGracePeriod       = "identity-change-grace-period"
	IdentityRestoreGracePeriod      = "identity-restore-grace-period"
	AllocatorListTimeoutName        = "allocator-list-timeout"
	KVstorePeriodicSync             = "kvstore-periodic-sync"
	KVstoreConnectivityTimeout      = "kvstore-connectivity-timeout"
	IPAllocationTimeout             = "ip-allocation-timeout"

	EndpointStatus = "endpoint-status"
	GopsPort       = "gops-port"

	EndpointStatusPolicy      = "policy"
	EndpointStatusHealth      = "health"
	EndpointStatusControllers = "controllers"
	EndpointStatusLog         = "log"
	EndpointStatusState       = "state"

	EnableDolphinEndpointSlice = "enable-dolphin-endpointslice"
)

func (c *DaemonConfig) DolphinNamespaceName() string {
	return c.K8sNamespace
}

// operator or application runs as a daemonset on
// each node
type DaemonConfig struct {
	ConfigFile   string
	ConfigDir    string
	CreationTime time.Time
	Opts         *IntOptions

	K8sNamespace string

	IPv6ClusterAllocCIDR     string
	IPv6ClusterAllocCIDRBase string
	IPAMDefaultIPPool        string

	IPAM        string
	JoinCluster bool

	EnableHostIPRestore  bool
	EnableEndpointRoutes bool

	EnableHealthChecking            bool
	EnableEndpointHealthChecking    bool
	EnableHealthCheckLoadBalancerIP bool
	EnableHealthCheckNodePort       bool

	EnableIPv4    bool
	EnableIPv6    bool
	EnableIPv6NDP bool
	EnableSCTP    bool

	EnableL7Proxy              bool
	EndpointStatus             map[string]struct{}
	EnableDolphinEndpointSlice bool

	DNSMaxIPsPerRestoredRule int

	ToFQDNsProxyPort     int
	ToFQDNsMaxIPsPerHost int

	KVstorePeriodicSync        time.Duration
	KVstoreConnectivityTimeout time.Duration
	KVStoreOpt                 map[string]string

	IPAllocationTimeout time.Duration

	IdentityChangeGracePeriod  time.Duration
	IdentityRestoreGracePeriod time.Duration

	FixedIdentityMapping map[string]string
	LogOpt               map[string]string

	LoopbackIPv4    string
	AnnotateK8sNode bool

	K8sServiceCacheSize uint

	IdentityAllocationMode    string
	AllowICMPFragNeeded       bool
	EnableWellKnownIdentities bool

	AllocatorListTimeout  time.Duration
	EnableICMPRules       bool
	ExternalClusterIP     bool
	EnableBGPControlPlane bool

	EnableK8sNetworkPolicy bool
	EnableEnvoyConfig      bool
	EnableVTEP             bool

	MaxControllerInterval int
	LogDriver             []string
	Debug                 bool

	DisableDolphinEndpointCRD bool
}

var (
	Config = &DaemonConfig{
		CreationTime:                    time.Now(),
		Opts:                            NewIntOptions(&DaemonOptionLibrary),
		IPv6ClusterAllocCIDR:            defaults.IPv6ClusterAllocCIDR,
		IPv6ClusterAllocCIDRBase:        defaults.IPv6ClusterAllocCIDRBase,
		IPAMDefaultIPPool:               defaults.IPAMDefaultIPPool,
		EnableHostIPRestore:             defaults.EnableHostIPRestore,
		EnableHealthChecking:            defaults.EnableHealthChecking,
		EnableEndpointHealthChecking:    defaults.EnableEndpointHealthChecking,
		EnableHealthCheckLoadBalancerIP: defaults.EnableHealthCheckLoadBalancerIP,
		EnableHealthCheckNodePort:       defaults.EnableHealthCheckNodePort,
		EnableIPv4:                      defaults.EnableIPv4,
		EnableIPv6:                      defaults.EnableIPv6,
		EnableIPv6NDP:                   defaults.EnableIPv6NDP,
		EnableSCTP:                      defaults.EnableSCTP,
		EnableL7Proxy:                   defaults.EnableL7Proxy,
		EndpointStatus:                  make(map[string]struct{}),
		DNSMaxIPsPerRestoredRule:        defaults.DNSMaxIPsPerRestoredRule,
		ToFQDNsMaxIPsPerHost:            defaults.ToFQDNsMaxIPsPerHost,
		KVstorePeriodicSync:             defaults.KVstorePeriodicSync,
		KVstoreConnectivityTimeout:      defaults.KVstoreConnectivityTimeout,
		IPAllocationTimeout:             defaults.IPAllocationTimeout,
		IdentityChangeGracePeriod:       defaults.IdentityChangeGracePeriod,
		IdentityRestoreGracePeriod:      defaults.IdentityRestoreGracePeriod,
		FixedIdentityMapping:            make(map[string]string),
		KVStoreOpt:                      make(map[string]string),
		LogOpt:                          make(map[string]string),
		LoopbackIPv4:                    defaults.LoopbackIPv4,
		EnableEndpointRoutes:            defaults.EnableEndpointRoutes,
		AnnotateK8sNode:                 defaults.AnnotateK8sNode,
		K8sServiceCacheSize:             defaults.K8sServiceCacheSize,

		IdentityAllocationMode:    IdentityAllocationModeKVstore,
		AllowICMPFragNeeded:       defaults.AllowICMPFragNeeded,
		EnableWellKnownIdentities: defaults.EnableWellKnownIdentities,
		AllocatorListTimeout:      defaults.AllocatorListTimeout,
		EnableICMPRules:           defaults.EnableICMPRules,

		ExternalClusterIP:      defaults.ExternalClusterIP,
		EnableVTEP:             defaults.EnableVTEP,
		EnableBGPControlPlane:  defaults.EnableBGPControlPlane,
		EnableK8sNetworkPolicy: defaults.EnableK8sNetworkPolicy,
		EnableEnvoyConfig:      defaults.EnableEnvoyConfig,
	}
)

func (c *DaemonConfig) Populate(vp *viper.Viper) {
	c.AllowICMPFragNeeded = vp.GetBool(AllowICMPFragNeeded)
	c.AnnotateK8sNode = vp.GetBool(AnnotateK8sNode)
	c.EnableIPv4 = vp.GetBool(EnableIPv4Name)
	c.EnableIPv6 = vp.GetBool(EnableIPv6Name)
	c.EnableIPv6NDP = vp.GetBool(EnableIPv6NDPName)
	c.EnableSCTP = vp.GetBool(EnableSCTPName)
	c.EnableEndpointRoutes = vp.GetBool(EnableEndpointRoutes)
	c.EnableHealthChecking = vp.GetBool(EnableHealthChecking)
	c.EnableEndpointHealthChecking = vp.GetBool(EnableEndpointHealthChecking)
	c.EnableHealthCheckNodePort = vp.GetBool(EnableHealthCheckNodePort)
	c.EnableHealthCheckLoadBalancerIP = vp.GetBool(EnableHealthCheckLoadBalancerIP)
	c.IdentityChangeGracePeriod = vp.GetDuration(IdentityChangeGracePeriod)
	c.IdentityRestoreGracePeriod = vp.GetDuration(IdentityRestoreGracePeriod)
	c.K8sServiceCacheSize = uint(vp.GetInt(K8sServiceCacheSize))
	c.AllocatorListTimeout = vp.GetDuration(AllocatorListTimeoutName)
	c.KVstorePeriodicSync = vp.GetDuration(KVstorePeriodicSync)
	c.KVstoreConnectivityTimeout = vp.GetDuration(KVstoreConnectivityTimeout)
	c.IPAllocationTimeout = vp.GetDuration(IPAllocationTimeout)
}

func LogRegisteredOptions(vp *viper.Viper, entry *logrus.Entry) {
	keys := vp.AllKeys()
	sort.Strings(keys)
	for _, k := range keys {
		ss := vp.GetStringSlice(k)
		if len(ss) == 0 {
			sm := vp.GetStringMap(k)
			for k, v := range sm {
				ss = append(ss, fmt.Sprintf("%s=%s", k, v))
			}
		}
		if len(ss) > 0 {
			entry.Infof("  --%s='%s'", k, strings.Join(ss, ""))
		} else {
			entry.Infof("  --%s='%s'", k, vp.GetString(k))
		}
	}
}

func InitConfig(cmd *cobra.Command, programName, configName string, vp *viper.Viper) func() {
	return func() {
		if vp.GetBool("version") {
			fmt.Printf("%s 1.0.0\n", programName)
			os.Exit(0)
		}

		if vp.GetString("cmdref") != "" {
			return
		}

		Config.ConfigFile = vp.GetString("conigFile")
		Config.ConfigDir = vp.GetString("configDir")
		vp.SetEnvPrefix("dolphin")

		if Config.ConfigDir != "" {
			if _, err := os.Stat(Config.ConfigDir); os.IsNotExist(err) {
				panic("Non existing configuration file dir")
			}
			if m, err := ReadDirConfig(Config.ConfigDir); err != nil {
				panic("failed to read configuration dir")
			} else {
				ReplaceDeprecatedFields(m)

				if err := validateConfigMap(cmd, m); err != nil {

				}

				if err := MergeConfig(vp, m); err != nil {
					panic("failed to merge configuration into the viper")
				}
			}
		}

		if Config.ConfigFile != "" {
			vp.SetConfigFile(Config.ConfigFile)
		} else {
			vp.SetConfigName(configName)
			vp.AddConfigPath("$HOME")
		}

		if err := vp.ReadInConfig(); err == nil {
			fmt.Println("successfully read configuration from configuration file")
		} else if Config.ConfigFile != "" {
			fmt.Printf("error reading configuration from file")
		} else {
			fmt.Printf("skipping reading configuration file")
		}

	}
}

func validateConfigMap(cmd *cobra.Command, m map[string]interface{}) error {
	flags := cmd.Flags()

	for key, value := range m {
		flag := flags.Lookup(key)
		if flag == nil {
			continue
		}

		var err error

		switch t := flag.Value.Type(); t {
		case "bool":
			_, err = cast.ToBoolE(value)
		case "duration":
			_, err = cast.ToDurationE(value)
		case "float32":
			_, err = cast.ToFloat32E(value)
		case "float64":
			_, err = cast.ToFloat64E(value)
		case "int":
			_, err = cast.ToIntE(value)
		case "int8":
			_, err = cast.ToInt8E(value)
		case "int16":
			_, err = cast.ToInt16E(value)
		case "int32":
			_, err = cast.ToInt32E(value)
		case "int64":
			_, err = cast.ToInt64E(value)
		case "map":
			// custom type, see pkg/option/map_options.go
			err = flag.Value.Set(fmt.Sprintf("%s", value))
		case "stringSlice":
			_, err = cast.ToStringSliceE(value)
		case "string":
			_, err = cast.ToStringE(value)
		case "uint":
			_, err = cast.ToUintE(value)
		case "uint8":
			_, err = cast.ToUint8E(value)
		case "uint16":
			_, err = cast.ToUint16E(value)
		case "uint32":
			_, err = cast.ToUint32E(value)
		case "uint64":
			_, err = cast.ToUint64E(value)
		default:
			fmt.Printf("Unable to validate option %s value of type %s", key, t)
		}

		if err != nil {
			return fmt.Errorf("option %s: %w", key, err)
		}
	}

	return nil
}

func ReplaceDeprecatedFields(m map[string]interface{}) {
	deprecatedFields := map[string]string{
		"monitor-aggregation-level":   MonitorAggregationName,
		"ct-global-max-entries-tcp":   CTMapEntriesGlobalTCPName,
		"ct-global-max-entries-other": CTMapEntriesGlobalAnyName,
	}
	for deprecatedOption, newOption := range deprecatedFields {
		if deprecatedValue, ok := m[deprecatedOption]; ok {
			if _, ok := m[newOption]; !ok {
				m[newOption] = deprecatedValue
			}
		}
	}
}

func MergeConfig(vp *viper.Viper, m map[string]interface{}) error {
	err := vp.MergeConfigMap(m)
	if err != nil {
		return fmt.Errorf("unable to read merge directory configuration: %w", err)
	}
	return nil
}

// read the file content and map the filename with its content
func ReadDirConfig(dirName string) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	files, err := os.ReadDir(dirName)
	if err != nil {
		return nil, fmt.Errorf("ERROR reading dir %v", err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fName := filepath.Join(dirName, f.Name())

		// handle symbolink
		if f.Type()&os.ModeSymlink == 0 {
			absFileName, err := filepath.EvalSymlinks(fName)
			if err != nil {
				fmt.Printf("could not retrieve the original symblink")
				continue
			}
			fName = absFileName
		}

		fi, err := os.Stat(fName)
		if err != nil {
			fmt.Printf("file satatus could not be retrieved")
			continue
		}
		if fi.Mode().IsDir() {
			fmt.Printf("the retrieved file is one directory. Skip")
			continue
		}

		b, err := os.ReadFile(fName)
		if err != nil {
			fmt.Printf("failed to read file")
			continue
		}
		// the content is a string
		m[f.Name()] = string(bytes.TrimSpace(b))
	}
	return m, nil
}
