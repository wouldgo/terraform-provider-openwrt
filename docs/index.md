---
page_title: "openwrt Provider"
---

# terraform-provider-incus

Use terraform, or opentofu, to manage an openwrt router.

## Description

This provider connets to an openwrt router through the UCI JSON RPC API.

The JSON RPC API requires a couple of packages to be used. Please see [Using the JSON-RPC API](https://github.com/openwrt/luci/blob/master/docs/JsonRpcHowTo.md) from openwrt.

## Basic Example

```terraform
provider "openwrt" {
  user = "root"
  password = "admin"
  remote = "http://192.168.8.1:8080"
}
```

## Configuration Reference

The following arguments are required.

* `user` - The username of the admin account.
* `password` - The password of the account.
* `remote` - The URL of the JSON RPC API.

## Environment Variables

* `OPENWRT_USER` - The username of the admin account.
* `OPENWRT_PASSWORD` - The password of the admin account.
* `OPENWRT_REMOTE` - The remote of the JSON RPC API.
