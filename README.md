# stencil-languages

Stencil module for generating and interacting with various programming
language specific files.

## Currently Supported

### Golang

* Merging `go.mod` files with `GolangMergeGoMod`. Example usage:

```go
// go.mod.tpl
{{- define "go.mod" }}
module github.com/rgst-io/my-golang-module

go 1.18

// All Go repos should have min module version and this module.
require github.com/jaredallard/cmdexec v1.2.0
{{- end }}

// We're generating go.mod in this file, so we shouldn't output
// a file  from this template.
{{ file.Skip }}

// Render the go mod file, use it if we don't have an existing go.mod
{{ $newGoMod := (stencil.ApplyTemplate "go.mod") }}
{{ $goModContents := $newGoMod }}
{{ file.Create "go.mod" 0600 now }}

// If the go.mod already exists, merge it with the generated one
// then write it to disk.
{{- if stencil.Exists "go.mod" }}
{{ $goModContents = (extensions.Call "github.com/rgst-io/stencil-languages.GolangMergeGoMod" (stencil.ReadFile "go.mod") $newGoMod) }}
{{- end }}

// Write out the go.mod
{{ file.SetContents $goModContents }}
```

## License

GPL-3.0
