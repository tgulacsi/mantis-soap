module github.com/tgulacsi/mantis-soap

require (
	github.com/UNO-SOFT/zlog v0.8.2
	github.com/peterbourgon/ff/v3 v3.4.0
	github.com/tgulacsi/go v0.27.7-0.20240917191515-7d0799c9cdb8
	github.com/titanous/json5 v1.0.0
	github.com/zRedShift/mimemagic v1.2.0
	golang.org/x/term v0.23.0
)

require (
	github.com/dgryski/go-linebreak v0.0.0-20180812204043-d8f37254e7d3 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kylewolfe/soaptrip v0.0.0-20160108184655-f6f12afc06a9 // indirect
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.23.0 // indirect
	golang.org/x/text v0.17.0 // indirect
)

go 1.22.0

toolchain go1.23.1

// replace github.com/tgulacsi/go => ../go
