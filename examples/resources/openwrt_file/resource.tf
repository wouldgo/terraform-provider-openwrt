resource "openwrt_file" "etc_test_txt" {
  path    = "/etc"
  name    = "test.txt"
  content = <<-EOT
this is a test
EOT
}
