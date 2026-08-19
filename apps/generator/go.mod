module github.com/shim1103/daily-it-podcast/apps/generator

go 1.22

require github.com/santhosh-tekuri/jsonschema/v6 v6.0.3

require github.com/shim1103/daily-it-podcast/contracts v0.0.0

require golang.org/x/text v0.14.0 // indirect

replace github.com/shim1103/daily-it-podcast/contracts => ../../contracts
