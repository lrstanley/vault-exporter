{
  jobs+:: {
    vault: "kube-system/vault",
  },

  prometheus_alerts+:: {
    groups+: [{
      name: "Vault",
      rules: [
        {
          alert: "VaultUp",
          expr: |||
            vault_up{job="%(vault)s"} != 1
          ||| % $.jobs,
          "for": "5m",
          labels: {
            severity: "critical",
          },
          annotations: {
            message: "Vault exporter for '{{ $labels.instance }}' cannot talk to Vault.",
          },
        },
        {
          alert: "VaultUninitialized",
          expr: |||
            vault_initialized{job="%(vault)s"} != 1
          ||| % $.jobs,
          "for": "5m",
          labels: {
            severity: "critical",
          },
          annotations: {
            message: "Vault '{{ $labels.instance }}' is uninitialized.",
          },
        },
        {
          alert: "VaultSealed",
          expr: |||
            vault_sealed{job="%(vault)s"} != 0
          ||| % $.jobs,
          "for": "5m",
          labels: {
            severity: "critical",
          },
          annotations: {
            message: "Vault '{{ $labels.instance }}' is sealed.",
          },
        },
        {
          alert: "VaultStandby",
          expr: |||
            count(vault_standby{job="%(vault)s"} == 0) != 1
          ||| % $.jobs,
          "for": "5m",
          labels: {
            severity: "critical",
          },
          annotations: {
            message: "There are {{ $value }} active Vault instance(s).",
          },
        },
      ],
    }],
  },
}
