# The default system settings on openwrt written as a `configfile` resource
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
