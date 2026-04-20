# Installation Guide

## Prerequisites

- Go >= 1.24
- Proxmox VE >= 8.x
- MySQL >= 8.0
- ZooKeeper >= 3.8

## Build

Compile the binary:

```sh
go build -o hci-vcls ./cmd/hci-vcls
```

## Systemd Setup

```ini
[Unit]
Description=HCI vCLS Service
After=network.target

[Service]
ExecStart=/usr/local/bin/hci-vcls serve --config /etc/hci-vcls.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```
