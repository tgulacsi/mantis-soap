module github.com/tgulacsi/mantis-soap

require (
	github.com/UNO-SOFT/zlog v0.8.6
	github.com/dlclark/regexp2 v1.11.5
	github.com/peterbourgon/ff/v4 v4.0.0-beta.1
	github.com/tgulacsi/go v0.28.4
	github.com/titanous/json5 v1.0.0
	github.com/zRedShift/mimemagic v1.2.0
	golang.org/x/net v0.40.0
	golang.org/x/term v0.32.0
)

require (
	github.com/dgryski/go-linebreak v0.0.0-20180812204043-d8f37254e7d3 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/quicktemplate v1.8.0 // indirect
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)

go 1.23.0

toolchain go1.24.0

// replace github.com/tgulacsi/go => ../go

replace github.com/peterbourgon/ff/v4 v4.0.0-beta.1 => github.com/UNO-SOFT/ff/v4 v4.0.0-beta.1.us
