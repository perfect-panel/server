#!MANAGED-CONFIG {{ .SubscribeURL }} interval=43200 strict=true

[General]
loglevel = notify
ipv6 = false
skip-proxy = localhost, *.local, injections.adguard.org, local.adguard.org, 0.0.0.0/8, 10.0.0.0/8, 17.0.0.0/8, 100.64.0.0/10, 127.0.0.0/8, 169.254.0.0/16, 172.16.0.0/12, 192.0.0.0/24, 192.0.2.0/24, 192.168.0.0/16, 192.88.99.0/24, 198.18.0.0/15, 198.51.100.0/24, 203.0.113.0/24, 224.0.0.0/4, 240.0.0.0/4, 255.255.255.255/32
tls-provider = default
show-error-page-for-reject = true
dns-server = 223.6.6.6, 119.29.29.29, 119.28.28.28
test-timeout = 5
internet-test-url = http://bing.com
proxy-test-url = http://bing.com

[Panel]
SubscribeInfo = {{ .SubscribeInfo }}, style=info

# Surfboard 配置文档：https://manual.getsurfboard.com/

[Proxy]
# 代理列表
{{ .Proxies }}

[Proxy Group]
# 代理组列表
{{ .ProxyGroup }}

[Rule]
# 规则列表
{{ .Rules }}
