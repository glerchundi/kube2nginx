package pkg

import (
    "encoding/base64"
    "encoding/json"
    "io/ioutil"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/glerchundi/kube2nginx/pkg/core"
    log "github.com/glerchundi/logrus"
    kclient "github.com/glerchundi/kubelistener/pkg/client"
    kapi "github.com/glerchundi/kubelistener/pkg/client/api/v1"
    "fmt"
)

type Config struct {
    IngressesData string
    NginxTemplatePath string
    KubeMasterURL string
    Namespace string
    Selector string
    ResyncInterval time.Duration
}

func NewConfig() *Config {
    return &Config{
        IngressesData: "",
        NginxTemplatePath: "",
        KubeMasterURL: "",
        Namespace: "",
        Selector: "",
        ResyncInterval: 1 * time.Minute,
    }
}

type KubeToNginx struct {
    // configuration
    config *Config
    // instantiated template
    tmpl *core.Template
    // parsed ingresses data
    ingressesData map[string]string
    // runtime upstreams data
    upstreamsData map[string]string
}

func NewKubeToNginx(config *Config) *KubeToNginx {
    return &KubeToNginx{
        config: config,
        tmpl: nil,
        ingressesData: make(map[string]string),
        upstreamsData: make(map[string]string),
    }
}

func (k2n *KubeToNginx) Run() {
    if k2n.config.IngressesData == "" {
        log.Fatal("no ingresses, no way")
    }

    id, err := base64.StdEncoding.DecodeString(k2n.config.IngressesData)
    if err != nil {
        log.Fatal(err)
    }

    k2n.ingressesData = make(map[string]string)
    if err := json.Unmarshal(id, &k2n.ingressesData); err != nil {
        log.Fatal(err)
    }

    // Get service account token
    serviceAccountToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
    if err != nil {
        log.Fatal(err)
    }

    // Get CA certificate data
    caCertificate, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
    if err != nil {
        log.Fatal(err)
    }

    // Create new k8s client
    kubeConfig := &kclient.ClientConfig{
        MasterURL: k2n.config.KubeMasterURL,
        Auth: &kclient.TokenAuth{ string(serviceAccountToken) },
        CaCertificate: caCertificate,
    }
    kubeClient, err := kclient.NewClient(kubeConfig)
    if err != nil {
        log.Fatal(err)
    }

    // Flow control channels
    recvChan := make(chan interface{}, 100)
    stopChan := make(<-chan struct{})
    doneChan := make(chan bool)
    errChan := make(chan error, 10)

    // Create informer from client
    informerConfig := &kclient.InformerConfig{
        Namespace: k2n.config.Namespace,
        Resource: "services",
        Selector: k2n.config.Selector,
        ResyncInterval: k2n.config.ResyncInterval,
    }
    i, err := kubeClient.NewInformer(
        informerConfig, recvChan,
        stopChan, doneChan, errChan,
    )
    if err != nil {
        log.Fatal(err)
    }

    tmplCfg := &core.TemplateConfig{
        SrcData:   core.NginxConf,
        Dest:      "/etc/nginx/nginx.conf",
        Uid:       os.Getuid(),
        Gid:       os.Getgid(),
        Mode:      "0644",
        Prefix:    "/lb",
        CheckCmd:  "/usr/sbin/nginx -t -c {{.}}",
        ReloadCmd: "/usr/sbin/nginx -s reload",
    }

    if k2n.config.NginxTemplatePath != "" {
        tmplCfg.Src = k2n.config.NginxTemplatePath
    }

    k2n.tmpl = core.NewTemplate(tmplCfg, false, false, false)

    go i.Run()

    // Wait for signal
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
    for {
        select {
        case v := <-recvChan:
            k2n.process(v)
        case err := <-errChan:
            log.Error(err)
        case s := <-signalChan:
            log.Infof("Captured %v. Exiting...", s)
            close(doneChan)
        case <-doneChan:
            os.Exit(0)
        }
    }
}

func (k2n *KubeToNginx) process(v interface{}) {
    switch vv := v.(type) {
    case *kapi.ServiceList:
        k2n.upstreamsData = make(map[string]string)
        for _, s := range vv.Items {
            k2n.addService(s)
        }
    case *kapi.WatchEvent:
        s, ok := vv.Object.(*kapi.Service)
        if !ok {
            log.Warnf("unknown k8s api object in a watch event was received: %v", vv.Object)
            return
        }

        switch vv.Type {
        case kapi.Added:
            k2n.addService(*s)
        case kapi.Deleted:
            k2n.deleteService(*s)
        case kapi.Modified:
            k2n.updateService(*s)
        }
    default:
        log.Warnf("unknown k8s api object was received: %v", v)
        return
    }

    // mixup everything
    kvs := make(map[string]string)
    for k, v := range k2n.ingressesData {
        kvs[k] = v
    }
    for k, v := range k2n.upstreamsData {
        kvs[k] = v
    }

    // render template
    err := k2n.tmpl.Render(kvs); if err != nil {
        log.Error(err)
    }
}

func (k2n *KubeToNginx) addService(s kapi.Service) {
    k2n.upstreamsData[getUpstreamKey(s)] = getUpstreamValue(s)
}

func (k2n *KubeToNginx) deleteService(s kapi.Service) {
    key := getUpstreamKey(s)
    _, ok := k2n.upstreamsData[key]
    if ok {
        delete(k2n.upstreamsData, key)
    }
}

func (k2n *KubeToNginx) updateService(s kapi.Service) {
    k2n.upstreamsData[getUpstreamKey(s)] = getUpstreamValue(s)
}

func getUpstreamKey(s kapi.Service) string {
    return fmt.Sprintf("/lb/upstreams/%s/servers/%s", s.Name, s.UID)
}

func getUpstreamValue(s kapi.Service) string {
    return fmt.Sprintf(`{"url": "%s:%s"}`, s.Spec.ClusterIP, s.Spec.Ports[0].TargetPort.String())
}