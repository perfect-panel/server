[General]
ipv6-vif = auto
ip-mode = dual
skip-proxy = 192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,127.0.0.0/8,localhost,*.local
bypass-tun = 192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,127.0.0.0/8,localhost,*.local
dns-server = system,119.29.29.29,223.5.5.5
hijack-dns = 8.8.8.8:53,8.8.4.4:53,1.1.1.1:53,1.0.0.1:53
allow-wifi-access = true
wifi-access-http-port = 6888
wifi-access-socks5-port = 6889
proxy-test-url = http://bing.com/generate_204
internet-test-url = http://wifi.vivo.com.cn/generate_204
test-timeout = 5
interface-mode = auto

[Proxy]
{{.Proxies}}

[Proxy Group]
🚀 Proxy = select,🌏 Auto,{{.Nodes}} 
🌏 Auto = fallback,{{.Nodes}},interval = 600,max-timeout = 3000
🍎 Apple = select,🚀 Proxy,🎯 Direct,{{.Nodes}}
🔍 Google = select,🚀 Proxy,{{.Nodes}}
🪟 Microsoft = select,🚀 Proxy,🎯 Direct,{{.Nodes}}
📠 X = select,🚀 Proxy,{{.Nodes}}
🤖 AI = select,🚀 Proxy,🎯 Direct,{{.Nodes}}
📟 Telegram = select,🚀 Proxy,{{.Nodes}}
📺 YouTube = select,🚀 Proxy,{{.Nodes}}
🇨🇳 China = select,🎯 Direct,🚀 Proxy,{{.Nodes}}
🐠 Final = select,🚀 Proxy,🎯 Direct,{{.Nodes}}
🎯 Direct = select,DIRECT

[Remote Rule]
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Apple/Apple.list, policy=🍎 Apple, tag=Apple, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Apple/Apple_Domain.list, policy=🍎 Apple, tag=Apple_Domain, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Google/Google.list, policy=🔍 Google, tag=Google, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Microsoft/Microsoft.list, policy=🪟 Microsoft, tag=Microsoft, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Twitter/Twitter.list, policy=📠 X, tag=X, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/OpenAI/OpenAI.list, policy=🤖 AI, tag=OpenAI, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Telegram/Telegram.list, policy=📟 Telegram, tag=Telegram, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/YouTube/YouTube.list, policy=📺 YouTube, tag=YouTube, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/YouTubeMusic/YouTubeMusic.list, policy=📺 YouTube, tag=YouTubeMusic, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Global/Global.list, policy=🚀 Proxy, tag=Global, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Global/Global_Domain.list, policy=🚀 Proxy, tag=Global_Domain, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/ChinaMax/ChinaMax.list, policy=🇨🇳 China, tag=ChinaMax, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/ChinaMax/ChinaMax_Domain.list, policy=🇨🇳 China, tag=ChinaMax_Domain, enabled=true
https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Loon/Lan/Lan.list, policy=🎯 Direct, tag=LAN, enabled=true

[Rule]
GEOIP,CN,🇨🇳 China
FINAL,🐠 Final

[Rewrite]
# Redirect Google Service
^https?:\/\/(www.)?g\.cn 302 https://www.google.com
^https?:\/\/(www.)?google\.cn 302 https://www.google.com
# Redirect Githubusercontent 
^https://.*\.githubusercontent\.com\/ header-replace Accept-Language en-us