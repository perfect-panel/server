#!/bin/bash

# 检查是否以 root 用户运行
if [ "$(id -u)" -ne 0 ]; then
    echo "请以 root 用户运行此脚本"
    exit 1
fi

# 系统检测，确定使用的包管理工具
if [ -f /etc/debian_version ]; then
    # Ubuntu / Debian 系统
    PKG_MANAGER="apt-get"
elif [ -f /etc/redhat-release ]; then
    # CentOS 系统
    PKG_MANAGER="yum"
else
    echo "不支持的系统"
    exit 1
fi

# 检查 jq 是否已安装，若未安装则自动安装
if ! command -v jq &> /dev/null; then
    echo "jq 未安装，正在安装 jq ..."
    if [ "$PKG_MANAGER" == "apt-get" ]; then
        apt-get update && apt-get install -y jq
    elif [ "$PKG_MANAGER" == "yum" ]; then
        yum install -y jq
    else
        echo "无法安装 jq，未知的包管理器"
        exit 1
    fi
fi

# 获取最新的版本号
VERSION=$(curl -s https://api.github.com/repos/perfect-panel/ppanel/releases/latest | jq -r .tag_name)

if [ "$VERSION" == "null" ]; then
    echo "无法获取最新版本号，请检查网络或 GitHub API 状态"
    exit 1
fi

# 安装路径
INSTALL_DIR="/opt/ppanel-server"
SERVICE_NAME="ppanel"

# 下载并解压二进制文件
echo "开始下载 ppanel 二进制文件，版本：$VERSION ..."
wget https://github.com/perfect-panel/ppanel/releases/download/$VERSION/ppanel-server-linux-amd64.tar.gz -O /tmp/ppanel-server-linux-amd64.tar.gz

# 创建安装目录
if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
fi

# 解压文件到安装目录
echo "解压文件到 $INSTALL_DIR ..."
tar -zxvf /tmp/ppanel-server-linux-amd64.tar.gz -C "$INSTALL_DIR" --strip-components=1

# 给二进制文件赋予执行权限
chmod +x "$INSTALL_DIR/ppanel-server"

# 创建 systemd 服务文件
echo "创建 systemd 服务文件 ..."
cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=PPANEL Server
After=network.target

[Service]
ExecStart=$INSTALL_DIR/ppanel-server
Restart=always
User=root
WorkingDirectory=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

# 重新加载 systemd 服务
echo "重新加载 systemd 配置 ..."
systemctl daemon-reload

# 启动服务
echo "启动 ppanel 服务 ..."
systemctl start $SERVICE_NAME

# 设置开机自启
echo "设置服务开机自启 ..."
systemctl enable $SERVICE_NAME

# 输出服务状态
echo "服务已启动，状态如下："
systemctl status $SERVICE_NAME

# 提示配置文件
echo "请通过 http://服务器地址:8080/init 初始化系统配置"
