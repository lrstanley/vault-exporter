package main

import (
	"net/http"
	_ "net/http/pprof"

	vault_api "github.com/hashicorp/vault/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("web.listen-address",
		"Address to listen on for web interface and telemetry.").
		Default(":9410").String()
	metricsPath = kingpin.Flag("web.telemetry-path",
		"Path under which to expose metrics.").
		Default("/metrics").String()
	vaultCACert = kingpin.Flag("vault-tls-cacert",
		"The path to a PEM-encoded CA cert file to use to verify the Vault server SSL certificate.").String()
	vaultClientCert = kingpin.Flag("vault-tls-client-cert",
		"The path to the certificate for Vault communication.").String()
	vaultClientKey = kingpin.Flag("vault-tls-client-key",
		"The path to the private key for Vault communication.").String()
	sslInsecure = kingpin.Flag("insecure-ssl",
		"Set SSL to ignore certificate validation.").
		Default("false").Bool()
)

const (
	namespace = "vault"
)

var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last query of Vault successful.",
		nil, nil,
	)
	initialized = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "initialized"),
		"Is the Vault initialised (according to this node).",
		nil, nil,
	)
	sealed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "sealed"),
		"Is the Vault node sealed.",
		nil, nil,
	)
	standby = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "standby"),
		"Is this Vault node in standby.",
		nil, nil,
	)
	replicationDrPrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "replication_dr_primary"),
		"Is this Vault node a primary disaster recovery replica.",
		nil, nil,
	)
	replicationDrSecondary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "replication_dr_secondary"),
		"Is this Vault node a secondary disaster recovery replica.",
		nil, nil,
	)
	replicationPerformancePrimary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "replication_performance_primary"),
		"Is this Vault node a primary performance replica.",
		nil, nil,
	)
	replicationPerformanceSecondary = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "replication_performance_secondary"),
		"Is this Vault node a secondary performance replica.",
		nil, nil,
	)
	info = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "info"),
		"Version of this Vault node.",
		[]string{"version", "cluster_name", "cluster_id"}, nil,
	)
)

// Exporter collects Vault health from the given server and exports them using
// the Prometheus metrics package.
type Exporter struct {
	client *vault_api.Client
}

// NewExporter returns an initialized Exporter.
func NewExporter() (*Exporter, error) {
	vaultConfig := vault_api.DefaultConfig()

	if *sslInsecure {
		tlsconfig := &vault_api.TLSConfig{
			Insecure: true,
		}
		err := vaultConfig.ConfigureTLS(tlsconfig)
		if err != nil {
			return nil, err
		}
	}

	if *vaultCACert != "" || *vaultClientCert != "" || *vaultClientKey != "" {

		tlsconfig := &vault_api.TLSConfig{
			CACert:     *vaultCACert,
			ClientCert: *vaultClientCert,
			ClientKey:  *vaultClientKey,
			Insecure:   *sslInsecure,
		}
		err := vaultConfig.ConfigureTLS(tlsconfig)
		if err != nil {
			return nil, err
		}
	}

	client, err := vault_api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		client: client,
	}, nil
}

// Describe describes all the metrics ever exported by the Vault exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- initialized
	ch <- sealed
	ch <- standby
	ch <- replicationDrPrimary
	ch <- replicationDrSecondary
	ch <- replicationPerformancePrimary
	ch <- replicationPerformanceSecondary
	ch <- info
}

func bool2float(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// Collect fetches the stats from configured Vault and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	health, err := e.client.Sys().Health()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Errorf("Failed to collect health from Vault server: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1,
	)
	ch <- prometheus.MustNewConstMetric(
		initialized, prometheus.GaugeValue, bool2float(health.Initialized),
	)
	ch <- prometheus.MustNewConstMetric(
		sealed, prometheus.GaugeValue, bool2float(health.Sealed),
	)
	ch <- prometheus.MustNewConstMetric(
		standby, prometheus.GaugeValue, bool2float(health.Standby),
	)

	if health.ReplicationDRMode == "disabled" {
		ch <- prometheus.MustNewConstMetric(
			replicationDrPrimary, prometheus.GaugeValue, 0,
		)
		ch <- prometheus.MustNewConstMetric(
			replicationDrSecondary, prometheus.GaugeValue, 0,
		)
	} else if health.ReplicationDRMode == "primary" {
		ch <- prometheus.MustNewConstMetric(
			replicationDrPrimary, prometheus.GaugeValue, 1,
		)
		ch <- prometheus.MustNewConstMetric(
			replicationDrSecondary, prometheus.GaugeValue, 0,
		)
	} else if health.ReplicationDRMode == "secondary" {
		ch <- prometheus.MustNewConstMetric(
			replicationDrPrimary, prometheus.GaugeValue, 0,
		)
		ch <- prometheus.MustNewConstMetric(
			replicationDrSecondary, prometheus.GaugeValue, 1,
		)
	}

	if health.ReplicationPerformanceMode == "disabled" {
		ch <- prometheus.MustNewConstMetric(
			replicationPerformancePrimary, prometheus.GaugeValue, 0,
		)
		ch <- prometheus.MustNewConstMetric(
			replicationPerformanceSecondary, prometheus.GaugeValue, 0,
		)
	} else if health.ReplicationPerformanceMode == "primary" {
		ch <- prometheus.MustNewConstMetric(
			replicationPerformancePrimary, prometheus.GaugeValue, 1,
		)
		ch <- prometheus.MustNewConstMetric(
			replicationPerformanceSecondary, prometheus.GaugeValue, 0,
		)
	} else if health.ReplicationPerformanceMode == "secondary" {
		ch <- prometheus.MustNewConstMetric(
			replicationPerformancePrimary, prometheus.GaugeValue, 0,
		)
		ch <- prometheus.MustNewConstMetric(
			replicationPerformanceSecondary, prometheus.GaugeValue, 1,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		info, prometheus.GaugeValue, 1, health.Version, health.ClusterName, health.ClusterID,
	)
}

func init() {
	prometheus.MustRegister(version.NewCollector("vault_exporter"))
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("vault_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting vault_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exporter, err := NewExporter()
	if err != nil {
		log.Fatalln(err)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
             <head><title>Vault Exporter</title></head>
             <body>
             <h1>Vault Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             <h2>Build</h2>
             <pre>` + version.Info() + ` ` + version.BuildContext() + `</pre>
             </body>
             </html>`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
