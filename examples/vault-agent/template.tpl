#!/bin/sh
{{ with secret "terraform/demo" }}
{{ range $k, $v := .Data.data }}export {{ $k }}={{ $v }}
{{ end }}
$*
{{ end }}