### Installation Instructions

#### Prerequisites
- MySQL 5.7+ (recommended: 8.0)
- Redis 6.0+ (recommended: 7.0)

#### Binary Installation

1. Determine your system architecture and download the corresponding binary file.

Download URL: `https://github.com/perfect-panel/ppanel/releases`

Example setup: OS: Linux amd64, User: root, Current directory: `/root`

- Download the binary file:

```shell
$ wget https://github.com/perfect-panel/ppanel/releases/download/v0.1.0/ppanel-server-linux-amd64.tar.gz
```

- Extract the binary file:

```shell
$ tar -zxvf ppanel-server-linux-amd64.tar.gz
```

- Navigate to the extracted directory:

```shell
$ cd ppanel-server-linux-amd64
```

- Grant execution permissions to the binary:

```shell
$ chmod +x ppanel
```

- Create a systemd service file:

```shell
$ cat > /etc/systemd/system/ppanel.service <<EOF
[Unit]
Description=PPANEL Server
After=network.target

[Service]
ExecStart=/root/ppanel-server-linux-amd64/ppanel
Restart=always
User=root
WorkingDirectory=/root/ppanel-server-linux-amd64

[Install]
WantedBy=multi-user.target
EOF
```

- Reload the systemd service configuration:

```shell
$ systemctl daemon-reload
```
- Start the service:

```shell
$ systemctl start ppanel
```

#### Additional Notes

1. Installation Path: The binary files will be extracted to /root/ppanel-server-linux-amd64.

2. systemd Service:
   - Service Name: ppanel
   
   - Service Configuration File: /etc/systemd/system/ppanel.service
   
   - Service Commands:
   
   - Start: systemctl start ppanel
   
   - Stop: systemctl stop ppanel
   
   - Restart: systemctl restart ppanel
   
   - Status: systemctl status ppanel
   
   - Enable on Boot: systemctl enable ppanel

3. Enable Auto-start: Use the following command to enable the service on boot:
    ```shell
    $ systemctl enable ppanel
    ```
4. Service Logs: By default, logs are output to `/root/ppanel-server-linux-amd64/ppanel.log`.

5. You can view service logs using: `journalctl -u ppanel -f`
6. If the configuration file is missing or empty, the service will start with default settings. The configuration file path is `./etc/ppanel.yaml`. Access `http://<server_address>:8080/init` to **initialize the system configuration**.

#### NGINX Reverse Proxy Configuration

Below is an example configuration to proxy the ppanel service to the domain api.ppanel.dev:

```nginx
server {
    listen 80;
    server_name ppanel.dev;

    location / {
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header REMOTE-HOST $remote_addr;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_http_version 1.1;
        
        add_header X-Cache $upstream_cache_status;
        
        # Set Nginx Cache
        set $static_file_cache 0;
        if ($uri ~* "\.(gif|png|jpg|css|js|woff|woff2)$") {
            set $static_file_cache 1;
            expires 1m;
        }
        if ($static_file_cache = 0) {
            add_header Cache-Control no-cache;
        }
    }
}
```

If using Cloudflare as a proxy service, you need to retrieve the user's real IP address. Add the following to the http section of the NGINX configuration file:

- Dependency: `ngx_http_realip_module`. Check if your NGINX build includes this module by running `nginx -V`. If not, you will need to recompile NGINX with this module.

```nginx
# Cloudflare Start
set_real_ip_from 0.0.0.0/0;
real_ip_header X-Forwarded-For;
real_ip_recursive on;
# Cloudflare End
```