schema_version = 1

project {
  license          = "MPL-2.0"
  copyright_holder = "https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors"

  # (OPTIONAL) A list of globs that should not have copyright/license headers.
  # Supports doublestar glob patterns for more flexibility in defining which
  # files or folders should be ignored
  header_ignore = [
    "**/types/**",
    "tools/**",
    ".goreleaser.yml",
    "docs/**",
    "examples/**",
  ]
}
