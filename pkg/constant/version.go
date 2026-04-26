package constant

// Version PPanel version
//
// Note: Version 默认值仅用于本地/开发构建。release 流程会通过 ldflags 注入
// `git describe --tags`(见 Makefile / .github/workflows/release.yml),
// 所以 CI 构建出来的二进制版本号始终来自 git tag,与这里的默认值无关。
var (
	Version     = "v4.3.0"
	BuildTime   = "unknown time"
	Repository  = "https://github.com/perfect-panel/server"
	ServiceName = "ApiService"
)
