provider "openwrt" {
  user     = "root"
  password = "admin"
  remote   = "http://192.168.8.1:8080"
  api_timeouts = {
    auth = "20s"
  }
}

resource "openwrt_opkg" "wanted_packages" {
  packages = ["curl", "tcpdump"]
}

resource "openwrt_file" "etc_test_txt" {
  path    = "/etc"
  name    = "test.txt"
  content = <<-EOT
this is a test
EOT
}

resource "openwrt_configfile" "dhcp" {
  name    = "dhcp"
  content = <<-EOT
config dnsmasq
    ....
EOT
}
