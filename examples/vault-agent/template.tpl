{{ with secret "terraform/demo" }}
{{ range $k, $v := .Data }}
{{ $k }}: {{ $v }}
{{ end }}
{{ end }}