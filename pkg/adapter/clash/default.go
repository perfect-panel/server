package clash

const DefaultTemplate = `
mixed-port: 7890
allow-lan: true
bind-address: "*"
mode: rule
log-level: info
external-controller: 127.0.0.1:9090
global-client-fingerprint: chrome
unified-delay: true
geox-url:
  mmdb: "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@release/geoip.metadb"
dns:
  enable: true
  ipv6: true
  enhanced-mode: fake-ip
  fake-ip-range: 198.18.0.1/16
  use-hosts: true
  default-nameserver:
    - 120.53.53.53
    - 1.12.12.12
  nameserver:
    - https://120.53.53.53/dns-query#skip-cert-verify=true
    - tls://1.12.12.12#skip-cert-verify=true
  proxy-server-nameserver:
    - https://120.53.53.53/dns-query#skip-cert-verify=true
    - tls://1.12.12.12#skip-cert-verify=true

proxies:

proxy-groups:

rules:
`
