---
page_title: "openwrt_file"

---

# openwrt_file (Resource)

Write configuration files to an arbitray path on the OpenWRT router.

## Example Usage


```hcl
resource "openwrt_file" "etc_test_txt" {
  path    = "/etc"
  name    = "test.txt"
  content = <<-EOT
  this is a test
  EOT
}
```

## Argument Reference

* `path` - *Required* - Path where file has to be.

* `name` - *Required* - Name of the file.

* `content` - *Required* - The content of the file.

## Attribute Reference

No attributes are exported.
