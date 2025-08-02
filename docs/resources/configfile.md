---
page_title: "openwrt_configfile"

---

# openwrt_configfile (Resource)

Write configuration files to `/etc/config` on the OpenWRT router.

## Example Usage

The default system settings on openwrt written as a `configfile` resource

```hcl
resource "openwrt_configfile" "system" {
 name    = "system"
 content = <<-EOT
 config timeserver 'ntp'
     option enabled '1'
     option enable_server '0'
     list server '0.openwrt.pool.ntp.org'
     list server '1.openwrt.pool.ntp.org'
     list server '2.openwrt.pool.ntp.org'
     list server '3.openwrt.pool.ntp.org'

 config system
     option hostname 'OpenWrt'
     option timezone 'UTC'
     option ttylogin '0'
     option log_size '64'
     option urandom_seed '0'
  EOT
}
```

## Argument Reference

* `name` - *Required* - Name of the configuration file.

* `content` - *Required* - The content of the configuration file.

* `commit` - *Optional* - If we should tell `uci` to run `commit` on the configuration file. (Default: true)

## Attribute Reference

No attributes are exported.
