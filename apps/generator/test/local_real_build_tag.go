//go:build local_real

// この file は build-tag による collection 境界を凍結する。behavior case は docs/tasks/todo/generator-narrow-*.md が所有する。
// file 名だけでは go test から除外されない。
package test

const LocalRealBuildTag = "local_real"
