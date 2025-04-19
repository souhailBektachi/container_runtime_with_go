module github.com/souhailBektachi/container_runtime_with_go

go 1.23.3

require (
	github.com/google/uuid v1.6.0
	github.com/spf13/cobra v1.9.1
)

// Add replace directive for mergo path change
replace github.com/imdario/mergo => dario.cat/mergo v1.0.1

require (
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
)
