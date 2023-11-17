<h1 align="center">
  <img src="https://github.com/MerlinKodo/clash-rev/raw/main/logo.png" alt="Clash" width="200">
  <br>Clash.Rev<br>
</h1>

<h4 align="center">A rule-based tunnel in Go.</h4>

## Important Notice

**There is a need for time to organize and analyze the code. If the Clash.Meta or Clash-core author restarts development within three months, our project will be halted to avoid unnecessary competition.**

## Work with CFW or Clash Verge

Clash.Rev is a Command Line Interface (CLI) program. You can operate it with `Clash For Windows` (CFW) or `Clash Verge` by adhering to the instructions provided below.

### CFW

Taking Windows platform as an example, navigate to the `resources/static/files/win/x64` directory under the installation directory of CFW. Rename the binary file you downloaded for your system to `clash-win64.exe`, then replace the file with the same name, and restart CFW.

If you need to use TUN mode, please run CFW as an administrator.

### Clash Verge

Taking Windows platform as an example, navigate to the installation directory of Clash Verge. Rename the downloaded binary file to `clash-meta.exe`, then replace the file with the same name, and restart Clash Verge. Switch the kernel to `Meta`, and you're all set.

## CLI Usage

Please refer to the [documentation](https://merlinkodo.github.io/Clash-Rev-Doc/startup/cli/) for detailed instructions.

```bash
Usage:
  clash [flags]

Flags:
      --cfg-header string   specify configuration file url header, env: CLASH_CONFIG_URL_HEADER
      --cfg-url string      specify configuration file url, env: CLASH_CONFIG_URL
  -f, --config string       specify configuration file, env: CLASH_CONFIG_FILE
  -d, --dir string          specify configuration directory, env: CLASH_HOME_DIR
      --ext-ctl string      override external controller address, env: CLASH_OVERRIDE_EXTERNAL_CONTROLLER
      --ext-ui string       override external ui directory, env: CLASH_OVERRIDE_EXTERNAL_UI_DIR
  -m, --geodata             set geodata mode
  -h, --help                help for clash
      --secret string       override secret, env: CLASH_OVERRIDE_SECRET
  -t, --test                test configuration and exit
  -v, --version             show current version of clash
```

## Description

Clash.Rev is a personal successor to the discontinued Clash-core and Clash.Meta, providing enhanced network management capabilities with a focus on user-friendliness and advanced features for modern networking needs.

## Features

This is a general overview of the features that comes with Clash.

- Inbound: HTTP, HTTPS, SOCKS5 server, TUN device
- Outbound: Shadowsocks(R), VMess, Trojan, Snell, SOCKS5, HTTP(S), Wireguard
- Rule-based Routing: dynamic scripting, domain, IP addresses, process name and more
- Fake-IP DNS: minimises impact on DNS pollution and improves network performance
- Transparent Proxy: Redirect TCP and TProxy TCP/UDP with automatic route table/rule management
- Proxy Groups: automatic fallback, load balancing or latency testing
- Remote Providers: load remote proxy lists dynamically
- RESTful API: update configuration in-place via a comprehensive API

## Documentation

You can find the latest documentation at [https://merlinkodo.github.io/Clash-Rev-Doc/](https://merlinkodo.github.io/Clash-Rev-Doc/).

## Credits

- [riobard/go-shadowsocks2](https://github.com/riobard/go-shadowsocks2)
- [v2ray/v2ray-core](https://github.com/v2ray/v2ray-core)
- [WireGuard/wireguard-go](https://github.com/WireGuard/wireguard-go)

## License

This software is released under the GPL-3.0 license. Thanks for the original author [Dreamacro](https://github.com/Dreamacro) and [wwqgtxx](https://github.com/wwqgtxx)
