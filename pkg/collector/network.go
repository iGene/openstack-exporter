package collector

import (
	"log"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
    "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/prometheus/client_golang/prometheus"
)

// networkCollector collects statistic about Neturon in an OpenStack Clusetr.
type networkCollector struct {
	provider *gophercloud.ProviderClient

	region string

	TotalFloatingIPsUsed prometheus.Gauge

	TotalNetworkNumber prometheus.Gauge

    TotalRouterNumber prometheus.Gauge
}

// GetNetworkNumber return the number of all network provisioned.
func GetNetworkNumber(networkClient *gophercloud.ServiceClient) (int, error) {

	opts := networks.ListOpts{}

	allPages, err := networks.List(networkClient, opts).AllPages()
	if err != nil {
		return 0, err
	}

	allNetworks, err := networks.ExtractNetworks(allPages)
	if err != nil {
		return 0, err
	}

	return len(allNetworks), nil
}

// GetIPsNumber returns the number of all floating IPs used.
func GetIPsNumber(networkClient *gophercloud.ServiceClient) (int, error) {

	opts := floatingips.ListOpts{}
	allPages, err := floatingips.List(networkClient, opts).AllPages()
	if err != nil {
		return 0, err
	}

	allFIPs, err := floatingips.ExtractFloatingIPs(allPages)
	if err != nil {
		return 0, err
	}

	return len(allFIPs), nil
}

func GetRouterNumber(networkClient *gophercloud.ServiceClient) (int, error) {

    opts := routers.ListOpts{}
    allPages, err := routers.List(networkClient, opts).AllPages()
    if err != nil {
    	return 0, err
    }

    allRouters, err := routers.ExtractRouters(allPages)
    if err != nil {
    	return 0, err
    }

    return len(allRouters), nil
}

// NewNetworkCollector create an instance of networkCollector.
func NewNetworkCollector(provider *gophercloud.ProviderClient, region string) *networkCollector {
	return &networkCollector{
		provider: provider,
		region:   region,
		TotalFloatingIPsUsed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "openstack_total_floating_ips_used",
				Help: "Total number of floating IPs used.",
			},
		),
		TotalNetworkNumber: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "openstack_total_network_number",
				Help: "Number of total networks.",
			},
		),
        TotalRouterNumber: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "openstack_total_router_number",
                Help: "Number of total routers",
            },
        ),
	}
}

func (n *networkCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{
		n.TotalNetworkNumber,
		n.TotalFloatingIPsUsed,
        n.TotalRouterNumber,
	}
}

func (n *networkCollector) collect() error {
	region := gophercloud.EndpointOpts{Region: n.region}
	networkClient, err := openstack.NewNetworkV2(n.provider, region)
	if err != nil {
		return err
	}

	netNumber, err := GetNetworkNumber(networkClient)
	if err != nil {
		return err
	}

	ipsNumber, err := GetIPsNumber(networkClient)
	if err != nil {
		return err
	}

    routerNumber, err := GetRouterNumber(networkClient)
    if err != nil {
        return err
    }

	n.TotalFloatingIPsUsed.Set(float64(ipsNumber))
	n.TotalNetworkNumber.Set(float64(netNumber))
    n.TotalRouterNumber.Set(float64(routerNumber))

	return nil
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by networkCollector.
func (n *networkCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range n.collectorList() {
		metric.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics.
func (n *networkCollector) Collect(ch chan<- prometheus.Metric) {

	if err := n.collect(); err != nil {
		log.Println("failed collecting network metrics:", err)
	}

	for _, metric := range n.collectorList() {
		metric.Collect(ch)
	}

}
