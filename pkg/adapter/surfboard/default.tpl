#!MANAGED-CONFIG {{.SubscribeURL}} interval=43200 strict=true

[General]
dns-server = system, 119.29.29.29, 223.5.5.5
skip-proxy = 192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12, 127.0.0.0/8, localhost, *.local
always-real-ip = *.lan, lens.l.google.com, *.srv.nintendo.net, *.stun.playstation.net, *.xboxlive.com, xbox.*.*.microsoft.com, *.msftncsi.com, *.msftconnecttest.com
proxy-test-url = http://www.gstatic.com/generate_204
internet-test-url = http://connectivitycheck.platform.hicloud.com/generate_204
test-timeout = 5
http-listen = 0.0.0.0:6088
socks5-listen = 0.0.0.0:6089

[Panel]
SubscribeInfo = {{.SubscribeInfo}}, style=info

[Proxy]
{{.Proxies}}

[Proxy Group]
🚀 Proxy = select, 🌏 Auto, 🎯 Direct, include-other-group=🇺🇳 Nodes
🍎 Apple = select, 🚀 Proxy, 🎯 Direct, include-other-group=🇺🇳 Nodes
🔍 Google = select, 🚀 Proxy, 🎯 Direct, include-other-group=🇺🇳 Nodes
🪟 Microsoft = select, 🚀 Proxy, 🎯 Direct, include-other-group=🇺🇳 Nodes
📺 GlobalMedia = select, 🚀 Proxy, 🎯 Direct, include-other-group=🇺🇳 Nodes
🤖 AI = select, 🚀 Proxy, 🎯 Direct, include-other-group=🇺🇳 Nodes
🪙 Crypto = select, 🚀 Proxy, 🎯 Direct, include-other-group=🇺🇳 Nodes
🎮 Game = select, 🚀 Proxy, 🎯 Direct, include-other-group=🇺🇳 Nodes
📟 Telegram = select, 🚀 Proxy, 🎯 Direct, include-other-group=🇺🇳 Nodes
🇨🇳 China = select, 🎯 Direct, 🚀 Proxy, include-other-group=🇺🇳 Nodes
🐠 Final = select, 🎯 Direct, 🚀 Proxy, include-other-group=🇺🇳 Nodes
🌏 Auto = fallback, include-other-group=🇺🇳 Nodes, url=http://www.gstatic.com/generate_204, interval=600, timeout=5
🎯 Direct = select, DIRECT, hidden=1
🇺🇳 Nodes = select, {{.ProxyGroup}}, hidden=1

[Rule]
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Apple/Apple_All.list, 🍎 Apple
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Google/Google.list, 🔍 Google
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/GitHub/GitHub.list, 🪟 Microsoft
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Microsoft/Microsoft.list, 🪟 Microsoft
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/HBO/HBO.list, 📺 GlobalMedia
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Disney/Disney.list, 📺 GlobalMedia
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/TikTok/TikTok.list, 📺 GlobalMedia
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Netflix/Netflix.list, 📺 GlobalMedia
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/GlobalMedia/GlobalMedia_All_No_Resolve.list, 📺 GlobalMedia
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Telegram/Telegram.list, 📟 Telegram
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/OpenAI/OpenAI.list, 🤖 AI
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Gemini/Gemini.list, 🤖 AI
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Copilot/Copilot.list, 🤖 AI
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Claude/Claude.list, 🤖 AI
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Crypto/Crypto.list, 🪙 Crypto
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Cryptocurrency/Cryptocurrency.list, 🪙 Crypto
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Game/Game.list, 🎮 Game
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Global/Global_All_No_Resolve.list, 🚀 Proxy
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/ChinaMax/ChinaMax_All_No_Resolve.list, 🇨🇳 China
RULE-SET, https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Surge/Lan/Lan.list, 🎯 Direct

GEOIP, CN, 🇨🇳 China
FINAL, 🐠 Final, dns-failed

[URL Rewrite]
^https?:\/\/(www.)?g\.cn https://www.google.com 302
^https?:\/\/(www.)?google\.cn https://www.google.com 302