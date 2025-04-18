---
page_title: "openwrt_system"
  
---

# openwrt_system (Resource)

Manage the system settings in openwrt

## Example Usage

The default system settings on openwrt.

```hcl
resource "openwrt_system" "system" {
  hostname = 'OpenWrt'
  timezone = 'UTC'
  ttylogin = '0'
  log_size = '64'
  urandom_seed = '0'
}
```

## Argument Reference

* `buffersize` - *Optional* - Size of the kernel message buffer.

* `conloglevel` - *Optional* - The maximum log level for kernel messages to be logged to the console. (Default: 7)

* `cronloglevel` - *Optional* - The minimum level for cron messages to be logged to syslog. 0 will print all debug messages, 8 will log command executions, and 9 or higher will only log error messages. (Default: 5)

* `description` - *Optional* - A short, single-line description for this system. It should be suitable for human consumption in user interfaces, such as LuCI, selector UIs in remote administration applications, or remote UCI (over ubus RPC).

* `hostname` - *Optional* - The hostname for this system. (Default: "OpenWrt")

* `klogconloglevel` - *Optional* - The maximum log level for kernel messages to be logged to the console. Only messages with a level lower than this will be printed to the console. Identical to conloglevel and will override it. (Default: 7)

* `log_buffer_size` - *Optional* - Size of the log buffer of the procd based system log, that is accessible via the logread command. Defaults to the value of log_size if unset.

* `log_file` - *Optional* - File to write log messages to (type file). The default is to not write a log in a file. The most often used location for a system log file is `/var/log/messages`.

* `log_hostname` - *Optional* - Hostname to send to remote syslog. If none is provided, the actual hostname is send. This feature is only present in 17.xx and later versions

* `log_ip` - *Optional* - IP address of a syslog server to which the log messages should be sent in addition to the local destination.

* `log_port` - *Optional* - Port number of the remote syslog server specified with log_ip. (Default: 514)

* `log_prefix` - *Optional* - Adds a prefix to all log messages send over network.

* `log_proto` - *Optional* - Sets the protocol to use for the connection, either tcp or udp. (Default: "udp")

* `log_remote` - *Optional* - Enables remote logging. (Default: 1)

* `log_size` - *Optional* - Size of the file based log buffer in KiB (see log_file). This value is used as the fallback value for log_buffer_size if the latter is not specified. (Default: 64)

* `log_trailer_null` - *Optional* - Use \0 instead of \n as trailer when using TCP. (Default: 0)

* `log_type` - *Optional* - Either circular or file. The circular option is a fixed size queue in memory, while the file is a dynamically sized file, that can be in memory, or written to disk. Note: If log_type is set to file, then at some point when the log fills, the device may encounter an out-of-space condition. This is especially an issue for devices with limited onboard storage: in memory, or on flash. (Default: "circular")

* `notes` - *Optional* - A multi-line, free-form text field about this system that can be used in any way the user wishes, e.g. to hold installation notes, or unit serial number and inventory number, location, etc.

* `timezone` - *Optional* - POSIX.1 time zone string corresponding to the time zone in which date and time should be displayed by default. See [timezone database](https://github.com/openwrt/luci/blob/master/modules/luci-lua-runtime/luasrc/sys/zoneinfo/tzdata.lua) for a mapping between IANA/Olson and POSIX.1 formats. (For London this corresponds to GMT0BST,M3.5.0/1,M10.5.0) (Default: "UTC")

* `ttylogin` - *Optional* - Require authentication for local users to log in the system. Disabled by default. It applies to the access methods listed in /etc/inittab, such as keyboard and serial. (Default: 0)

* `urandom_seed` - *Optional* - Path of the seed. Enables saving a new seed on each boot. (Default: 0)

* `zonename` - *Optional* - IANA/Olson time zone string. If zoneinfo-* packages are present, possible values can be found by running find /usr/share/zoneinfo. See [timezone database](https://github.com/openwrt/luci/blob/master/modules/luci-lua-runtime/luasrc/sys/zoneinfo/tzdata.lua) for a mapping between IANA/Olson and POSIX.1 formats. (For London this corresponds to Europe/London) (Default: UTC)

* `zram_comp_algo` - *Optional* - Compression algorithm to use for ZRAM, can be one of lzo, lzo-rle, lz4, zstd. (Default: "lzo")

* `zram_size_mb` - *Optional* - Size of ZRAM in MB. (Default: ramsize in Kb divided by 2048)

## Attribute Reference

No attributes are exported.

## Importing

Importing can be done by specifying the resource id as `system`. The provider is
going to try guess the correct provider configuration as there is only one
system directive.

```shell
terraform import openwrt_system.system system
```

