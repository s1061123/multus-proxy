package proxy
import (
	"strings"
	"fmt"
	"time"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/kubernetes/pkg/util/async"
	corev1 "k8s.io/api/core/v1"

	informers "k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"

	"github.com/s1061123/multus-proxy/pkg/utils"
	// "k8s.io/klog"
)

// Options contains everything necessary to create/run a daemon
type Options struct {
	//config *DaemonConfiguration
	test string
	Kubeconfig string
}

func NewOptions() *Options{
	return &Options{
		Kubeconfig: "",
	}
}

//
type ProxyServer struct {
	test string
	kubeClient kubernetes.Interface
	syncRunner      *async.BoundedFrequencyRunner

	customInformerFactory informers.SharedInformerFactory
}

func (p *ProxyServer) OnServiceAdd(service *corev1.Service) {
	fmt.Printf("[Service/add]%v\n", service)
}
func (p *ProxyServer) OnServiceUpdate(oldService, service *corev1.Service) {
	/*
	if proxier.serviceChanges.Update(oldService, service) && proxier.isInitialized() {
		proxier.syncRunner.Run()
	}*/
	//fmt.Printf("[Service/update]\nold:%v\nnew:%v\n", oldService, service)
}
func (p *ProxyServer) OnServiceDelete(service *corev1.Service) {
	fmt.Printf("[Service/delete]%v\n", service)
}
func (p *ProxyServer) OnServiceSynced() {
	fmt.Printf("[Service/synced]\n")
}

func (p *ProxyServer) OnEndpointsAdd(endpoints *corev1.Endpoints) {
	fmt.Printf("[Endpoints/add]%v\n", endpoints)
}
func (p *ProxyServer) OnEndpointsUpdate(oldEndpoints, endpoints *corev1.Endpoints) {
	fmt.Printf("[Endpoints/update]%v\n", endpoints)
}
func (p *ProxyServer) OnEndpointsDelete(endpoints *corev1.Endpoints) {
	fmt.Printf("[Endpoints/delete]%v\n", endpoints)
}
func (p *ProxyServer) OnEndpointsSynced() {
	fmt.Printf("[Endpoints/synced]\n")
}


func (p *ProxyServer) syncProxyRules() {
	fmt.Printf("loop\n")
	time.Sleep(2 * time.Second)

	/*
	svcList, err := p.customClient.MultusV1alpha().Services("default").List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("err: %v", err)
		return
	}
	for _, v := range svcList.Items {
		fmt.Printf("svc:%s\n", v.Name)
	}
	*/

	// list all pods in the node
	podList, err := p.kubeClient.CoreV1().Pods("").List(
		metav1.ListOptions{FieldSelector: "spec.nodeName=kube-node-1" })
	if err != nil {
		fmt.Printf("ERR:%v", err)
	}

	runtimeClient, runtimeConn, err := utils.GetRuntimeClient()
	if err != nil {
		fmt.Printf("ERR:%v", err)
		return
	}
        defer utils.CloseConnection(runtimeConn)
	for _, v := range podList.Items {
		podContainerID := v.Status.ContainerStatuses[0].ContainerID
		//containerType := podContainerID[0:strings.Index(podContainerID, ":")]
		containerID := podContainerID[strings.Index(podContainerID, "://")+3:]

		ns, err := utils.GetCrioContainerNS(runtimeClient, "", containerID, "")
		if err != nil {
			fmt.Printf("pod:%s (???: %v\t%s %s)\n", v.Name, err, podContainerID, containerID)
		} else {
			fmt.Printf("pod:%s (%s)\n", v.Name, ns)
		}
	}
}

func NewProxyServer(o *Options) (*ProxyServer, error) {
	var config *rest.Config
	var err error
	if o.Kubeconfig == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", o.Kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	//XXX: need to tuning parameters (duration)
	customInformerFactory := informers.NewSharedInformerFactory(kubeClient, time.Second*10)

	//XXX: need to tuning parameters
	burstSyncs := 2
	minSyncPeriod := 5 * time.Second
	syncPeriod := 10 * time.Second
	proxyServer := &ProxyServer{
		kubeClient: kubeClient,
		customInformerFactory: customInformerFactory,
	}
	proxyServer.syncRunner = async.NewBoundedFrequencyRunner("sync-runner", proxyServer.syncProxyRules, minSyncPeriod, syncPeriod, burstSyncs)

	//XXX Need to do https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/config/config.go
	// service
	serviceConfig := NewServiceConfig(customInformerFactory.Core().V1().Services(), time.Second*10) // configSyncPeriod
	serviceConfig.RegisterEventHandler(proxyServer)
	go serviceConfig.Run(wait.NeverStop)

	//XXX: endpoints
	endpointsConfig := NewEndpointsConfig(customInformerFactory.Core().V1().Endpoints(), time.Second*10) // configSyncPeriod
	endpointsConfig.RegisterEventHandler(proxyServer)
	go endpointsConfig.Run(wait.NeverStop)

	customInformerFactory.Start(wait.NeverStop)

	return proxyServer, nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Kubeconfig, "config", o.Kubeconfig, "The path to the kubeconfig file.")
}

func (o *Options) Run() error {
	fmt.Printf("run!\n")
	proxyServer, err := NewProxyServer(o)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}
	proxyServer.syncRunner.Loop(wait.NeverStop)

	return nil
}

func NewDaemonCommand() *cobra.Command {
	opts := NewOptions()

	cmd := &cobra.Command{
		Use: "kube-netattach-daemon",
		Long: `TBD`,
		Run: func(cmd *cobra.Command, args []string) {
			opts.Run()
		},
	}

	opts.AddFlags(cmd.Flags())
	//var err error
	return cmd
}

