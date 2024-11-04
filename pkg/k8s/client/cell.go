package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	apiext_clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/connrotation"

	"github.com/ccfish2/infra/pkg/controller"
	"github.com/ccfish2/infra/pkg/hive/cell"
	dolphin_clientset "github.com/ccfish2/infra/pkg/k8s/client/clientset/versioned"
	k8smetrics "github.com/ccfish2/infra/pkg/k8s/metrics"
)

var Cell = cell.Module(
	"k8s-client",
	"Kubernetes Client",

	cell.Config(defaultConfig),
	cell.Provide(newClientset),
)

var k8sHeartbeatControllerGroup = controller.NewGroup("k8s-heartbeat")

func setDialer(cfg Config, restConfig *rest.Config) func() {
	ctx := (&net.Dialer{
		Timeout:   cfg.K8sHeartbeatTimeout,
		KeepAlive: cfg.K8sHeartbeatTimeout,
	}).DialContext
	dialer := connrotation.NewDialer(ctx)
	restConfig.Dial = dialer.DialContext
	return dialer.CloseAll
}

func newClientset(lc cell.Lifecycle, log logrus.FieldLogger, cfg Config) (Clientset, error) {
	if !cfg.isEnabled() {
		return &compositeClientset{disabled: true}, nil
	}

	if cfg.K8sAPIServer != "" &&
		!strings.HasPrefix(cfg.K8sAPIServer, "http") {
		cfg.K8sAPIServer = "http://" + cfg.K8sAPIServer // default to HTTP
	}

	client := compositeClientset{
		log:        log,
		controller: controller.NewManager(),
		config:     cfg,
	}

	restConfig, err := createConfig(cfg.K8sAPIServer, cfg.K8sKubeConfigPath, cfg.K8sClientQPS, cfg.K8sClientBurst)
	if err != nil {
		return nil, fmt.Errorf("unable to create k8s client rest configuration: %w", err)
	}
	client.restConfig = restConfig
	defaultCloseAllConns := setDialer(cfg, restConfig)

	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create k8s REST client: %w", err)
	}

	if s := os.Getenv("DISABLE_HTTP2"); len(s) > 0 {
		client.closeAllConns = defaultCloseAllConns
	} else {
		client.closeAllConns = func() {
			utilnet.CloseIdleConnectionsFor(restConfig.Transport)
		}
	}

	restConfig.ContentConfig.ContentType = `application/vnd.kubernetes.protobuf`

	client.APIExtClientset, err = apiext_clientset.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create apiext k8s client: %w", err)
	}

	client.KubernetesClientset, err = kubernetes.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create k8s client: %w", err)
	}

	restConfig.ContentConfig.ContentType = `application/json`
	client.DolphinClientset, err = dolphin_clientset.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create dolphin k8s client: %w", err)
	}

	lc.Append(cell.Hook{
		OnStart: client.onStart,
		OnStop:  client.onStop,
	})

	return &client, nil
}

type Clientset interface {
	kubernetes.Interface
	apiext_clientset.Interface
	dolphin_clientset.Interface

	IsEnabled() bool
	Disable()
	Config() Config
	RestConfig() *rest.Config
}

type (
	KubernetesClientset = kubernetes.Clientset
	APIExtClientset     = apiext_clientset.Clientset
	DolphinClientset    = dolphin_clientset.Clientset
)

type compositeClientset struct {
	started  bool
	disabled bool

	*KubernetesClientset
	*APIExtClientset
	*DolphinClientset

	controller    *controller.Manager
	config        Config
	log           logrus.FieldLogger
	closeAllConns func()
	restConfig    *rest.Config
}

func (c *compositeClientset) Config() Config {
	return c.config
}

func (c *compositeClientset) Discovery() discovery.DiscoveryInterface {
	return c.KubernetesClientset.Discovery()
}

func (c *compositeClientset) IsEnabled() bool {
	return c != nil && c.config.isEnabled() && !c.disabled
}

func (c *compositeClientset) RestConfig() *rest.Config {
	return rest.CopyConfig(c.restConfig)
}

func (c *compositeClientset) Disable() {
	if c.started {
		panic("disabled after it has been started")
	}
	c.disabled = true
}

func (c *compositeClientset) onStart(startCtx cell.HookContext) error {
	if !c.IsEnabled() {
		return nil
	}

	if err := c.waitForConn(startCtx); err != nil {
		return err
	}
	c.startHeartbeat()

	// maybe we need some version capabilities
	c.started = true
	return nil
}

func (c *compositeClientset) startHeartbeat() {
	restClient := c.KubernetesClientset.RESTClient()

	timeout := c.config.K8sHeartbeatTimeout
	if timeout == 0 {
		return
	}

	heartBeat := func(ctx context.Context) error {
		res := restClient.Get().Resource("healthz").Do(ctx)
		return res.Error()
	}

	c.controller.UpdateController("k8s-heartbeat",
		controller.ControllerParams{
			Group: k8sHeartbeatControllerGroup,
			DoFunc: func(ctx context.Context) error {
				runHeartbeat(
					c.log,
					heartBeat,
					timeout,
					c.closeAllConns,
				)
				return nil
			},
			RunInterval: timeout,
		})
}

func runHeartbeat(log logrus.FieldLogger, heartBeat func(context.Context) error, timeout time.Duration, closeAllConns ...func()) {
	expireDate := time.Now().Add(-timeout)

	if k8smetrics.LastSuccessInteraction.Time().After(expireDate) {
		return
	}

	done := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go func() {
		err := heartBeat(ctx)
		if err != nil {
			statusError := &k8sErrors.StatusError{}
			if !errors.As(err, &statusError) ||
				statusError.ErrStatus.Code != http.StatusTooManyRequests {
				done <- err
			}
		}
		close(done)
	}()
}

func (c *compositeClientset) onStop(stopCtx cell.HookContext) error {
	if c.IsEnabled() {
		c.controller.RemoveAllAndWait()
		c.closeAllConns()
	}
	c.started = false
	return nil
}

func isConnReady(c kubernetes.Interface) error {
	_, err := c.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	return err
}

func (c *compositeClientset) waitForConn(ctx context.Context) error {
	stop := make(chan struct{})
	timeout := time.NewTimer(time.Minute)
	defer timeout.Stop()

	var err error
	wait.Until(func() {
		fmt.Printf("host establishing connection to apiserver")
		err = isConnReady(c)
		if err == nil {
			close(stop)
			return
		}

		select {
		case <-ctx.Done():
		case <-timeout.C:
		default:
			return
		}
		fmt.Printf("Unable to contact k8s api-server")
		close(stop)
	}, 5*time.Second, stop)
	if err == nil {
		fmt.Printf("connected to apiserver")
	}
	return err
}

func createConfig(apiServerURL, kubeCfgPath string, qps float32, burst int) (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)
	cmdName := "dolphin"
	if len(os.Args[0]) != 0 {
		cmdName = filepath.Base(os.Args[0])
	}
	userAgent := fmt.Sprintf("%s/%s", cmdName, "v1.16.2")

	switch {
	// If the apiServerURL and the kubeCfgPath are empty then we can try getting
	// the rest.Config from the InClusterConfig
	case apiServerURL == "" && kubeCfgPath == "":
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	case kubeCfgPath != "":
		if config, err = clientcmd.BuildConfigFromFlags("", kubeCfgPath); err != nil {
			return nil, err
		}
	case strings.HasPrefix(apiServerURL, "https://"):
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
		config.Host = apiServerURL
	default:
		config = &rest.Config{Host: apiServerURL, UserAgent: userAgent}
	}

	setConfig(config, userAgent, qps, burst)
	return config, nil
}

func setConfig(config *rest.Config, userAgent string, qps float32, burst int) {
	if userAgent != "" {
		config.UserAgent = userAgent
	}
	if qps != 0.0 {
		config.QPS = qps
	}
	if burst != 0 {
		config.Burst = burst
	}
}
