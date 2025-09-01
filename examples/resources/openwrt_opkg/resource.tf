resource "openwrt_opkg" "wanted_packages" {
  packages = ["curl", "tcpdump"]
}
