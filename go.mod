module github.com/souhailBektachi/container_runtime_with_go

go 1.23.3

require github.com/spf13/cobra v1.9.1

require (
	github.com/apex/log v1.9.0
	github.com/containers/image v3.0.2+incompatible
	github.com/containers/storage v1.57.2
	github.com/docker/docker v28.0.4+incompatible
	github.com/docker/go-units v0.5.0
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/inconshreveable/mousetrap v1.1.0
	github.com/moby/sys/capability v0.4.0
	github.com/moby/sys/mountinfo v0.7.2
	github.com/moby/sys/user v0.3.0
	github.com/mtrmac/gpgme v0.1.2
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/opencontainers/runtime-spec v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/pflag v1.0.6
	golang.org/x/crypto v0.0.0-20190426145343-a29dc8fdc734
	golang.org/x/sys v0.29.0
)

require (
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
)
