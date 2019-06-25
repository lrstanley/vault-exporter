local g = import "grafana-builder/grafana.libsonnet";

local row_settings = {
  height: "100px",
  showTitle: false,
};

local panel_settings = {
  repeat: "instance",
  colorBackground: true,
  thresholds: "0.5,0.5",
};

{
  dashboards+:: {
    "vault.json":
     g.dashboard("Vault")
      .addTemplate("job", "vault_up", "job")
      .addMultiTemplate("instance", "vault_up", "instance")
      .addRow(
        g.row("Up")
         .addPanel(
           g.panel("$instance") +
           g.statPanel('vault_up{job="$job",instance=~"$instance"}', 'none') +
           panel_settings {
             valueMaps: [
               { value: '0', op: '=', text: 'DOWN' },
               { value: '1', op: '=', text: 'UP' },
             ],
             colors: ["#d44a3a", "rgba(237, 129, 40, 0.89)", "#299c46"],
           }
         ) +
         row_settings
      )
      .addRow(
        g.row("Initialised")
          .addPanel(
            g.panel("$instance") +
            g.statPanel('vault_initialized{job="$job",instance=~"$instance"}', 'none') +
            panel_settings {
              valueMaps: [
                { value: '0', op: '=', text: 'UNINITIALISED' },
                { value: '1', op: '=', text: 'INITIALISED' },
              ],
              colors: ["#d44a3a", "rgba(237, 129, 40, 0.89)", "#299c46"],
            }
          ) +
          row_settings
      )
      .addRow(
        g.row("Sealed")
          .addPanel(
            g.panel("$instance") +
            g.statPanel('vault_sealed{job="$job",instance=~"$instance"}', 'none') +
            panel_settings {
              valueMaps: [
                { value: '0', op: '=', text: 'UNSEALED' },
                { value: '1', op: '=', text: 'SEALED' },
              ],
            }
          ) +
          row_settings
      )
      .addRow(
        g.row("Elected")
          .addPanel(
            g.panel("$instance") +
            g.statPanel('vault_standby{job="$job",instance=~"$instance"}', 'none') +
            panel_settings {
              valueMaps: [
                { value: '0', op: '=', text: 'MASTER' },
                { value: '1', op: '=', text: 'STANDBY' },
              ],
              thresholds: "0.5,1.5",
            }
          ) +
          row_settings
      )
      .addRow(
        g.row("Vault Server")
        .addPanel(
          g.panel("QPS") +
          g.queryPanel('sum(rate(vault_route_count{job="$job"}[1m])) by (instance)', "{{instance}}") +
          g.stack
        )
        .addPanel(
          g.panel("Latency") +
          g.queryPanel('max(vault_route{job="$job", quantile="0.99"}) by (instance)', "99th Percentile {{instance}}") +
          g.queryPanel('max(vault_route{job="$job", quantile="0.5"}) by (instance)', "50th Percentile {{instance}}") +
          g.queryPanel('sum(rate(vault_route_sum{job="$job"}[5m])) by (instance) / sum(rate(vault_route_count{job="$job"}[5m])) by (instance)', "Average {{instance}}") +
          { yaxes: g.yaxes("ms") }
        )
      )
      .addRow(
        g.row("Etcd Backend")
        .addPanel(
          g.panel("QPS") +
          g.queryPanel('sum(rate(vault_etcd_count{job="$job"}[1m])) by (instance)', "{{instance}}") +
          g.stack
        )
        .addPanel(
          g.panel("Latency") +
          g.queryPanel('max(vault_etcd{job="$job", quantile="0.99"}) by (instance)', "99th Percentile {{instance}}") +
          g.queryPanel('max(vault_etcd{job="$job", quantile="0.5"}) by (instance)', "50th Percentile {{instance}}") +
          g.queryPanel('sum(rate(vault_etcd_sum{job="$job"}[5m])) by (instance) / sum(rate(vault_etcd_count{job="$job"}[5m])) by (instance)', "Average {{instance}}") +
          { yaxes: g.yaxes("ms") }
        )
      )
  },
}
