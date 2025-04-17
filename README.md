terraform-provider-openwrt
==========================

openwrt provider for Terraform.

## Prerequisites

- [Terraform](http://terraform.io)
- [OpenWrt](https://openwrt.org)

## Installation

This provider is published in the [Terraform Registry](https://registry.terraform.io/providers/foxboron/openwrt).

### Quick Example

Add the following to your Terraform configuration:

```hcl
terraform {
  required_providers {
    openwrt = {
      source = "foxboron/openwrt"
    }
  }
}
```

### Development

#### Setup

1. Follow these [instructions](https://golang.org/doc/install) to setup a Golang development environment.
2. Checkout the repository `git clone ...`
3. Compile from sources to a development binary:

```shell
cd terraform-provider-openwrt
go build -v
```

4. Configure Terraform (`~/.terraformrc`) to use the development binary provider:

```shell
$ cat ~/.terraformrc
provider_installation {
  dev_overrides {
    "foxboron/openwrt" = "/home/<REPLACE_ME>/git/terraform-provider-openwrt"
  }
}
```

## Documentation

Full documentation can be found in the [`docs`](docs) directory.
