mode: rule
ipv6: true
allow-lan: true
bind-address: "*"
mixed-port: 7890
log-level: error
unified-delay: true
tcp-concurrent: true
external-controller: 0.0.0.0:9090

tun:
  enable: true
  stack: system
  auto-route: true

dns:
  enable: true
  cache-algorithm: arc
  listen: 0.0.0.0:1053
  ipv6: true
  enhanced-mode: fake-ip
  fake-ip-range: 198.18.0.1/16
  fake-ip-filter:
    - "*.lan"
    - "lens.l.google.com"
    - "*.srv.nintendo.net"
    - "*.stun.playstation.net"
    - "xbox.*.*.microsoft.com"
    - "*.xboxlive.com"
    - "*.msftncsi.com"
    - "*.msftconnecttest.com"
  default-nameserver:
    - 119.29.29.29
    - 223.5.5.5
  nameserver:
    - system
    - 119.29.29.29
    - 223.5.5.5
  fallback:
    - 8.8.8.8
    - 1.1.1.1
  fallback-filter:
    geoip: true
    geoip-code: CN

proxies:
{{.Proxies | toYaml | indent 2}}
proxy-groups:
{{.ProxyGroups | toYaml | indent 2}}
rules:
{{.Rules | toYaml | indent 2}}