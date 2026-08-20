module github.com/shim1103/daily-it-podcast/apps/generator

go 1.22

replace github.com/shim1103/daily-it-podcast/contracts => ../../contracts

require (
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	github.com/shim1103/daily-it-podcast/contracts v0.0.0
)
