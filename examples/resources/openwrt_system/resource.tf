#The default system settings on openwrt.
resource "openwrt_system" "system" {
  hostname     = "OpenWrt"
  timezone     = "UTC"
  ttylogin     = "0"
  log_size     = "64"
  urandom_seed = "0"
}
