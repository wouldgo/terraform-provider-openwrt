resource "openwrt_opkg" "wanted_packages" {
  packages = ["dnsmasq"]
}

resource "openwrt_file" "dnsmasq_conf" {
  path    = "/etc/dnsmasq.conf"
  content = file("${path.module}/dnsmasq.conf")

  depends_on = [openwrt_package.dnsmasq]
}

resource "openwrt_service" "dnsmasq" {
  name    = "dnsmasq"
  enabled = true

  # this ensures restart if the conf file changes
  triggers = {
    conf_sha = filesha256("${path.module}/dnsmasq.conf")
  }

  depends_on = [openwrt_file.dnsmasq_conf]
}
