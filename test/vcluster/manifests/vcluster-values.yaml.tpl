serviceCIDR: 10.43.0.0/16

syncer:
  extraArgs:
  - --tls-san={{.Name}}.k3s
