//go:build releaser

package main

// This file is a script which generates the .goreleaser.yaml file for all
// supported OpenTelemetry Collector distributions.
//
// Run it with `make generate-goreleaser`.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/nfpm/v2/files"
	"gopkg.in/yaml.v3"
)

var (
	ImagePrefixes = []string{"ghcr.io/grafana/opentelemetry-collector-components"}
	Architectures = []string{"386", "amd64", "arm64", "ppc64le"}

	distFlag = flag.String("d", "", "Single distribution to build")
)

func main() {
	flag.Parse()

	if len(*distFlag) == 0 {
		log.Fatal("no distribution to build")
	}

	project := Generate(ImagePrefixes, *distFlag)
	if err := yaml.NewEncoder(os.Stdout).Encode(&project); err != nil {
		log.Fatal(err)
	}
}

func Generate(imagePrefixes []string, dist string) config.Project {
	return config.Project{
		ProjectName: dist,
		Checksum: config.Checksum{
			NameTemplate: "{{ .ProjectName }}_checksums.txt",
		},

		Builds:          []config.Build{Build(dist)},
		Archives:        []config.Archive{Archive(dist)},
		NFPMs:           []config.NFPM{Package(dist)},
		Dockers:         DockerImages(imagePrefixes, dist),
		DockerManifests: DockerManifest(imagePrefixes, dist),
		SBOMs: []config.SBOM{
			{
				ID:        "archive",
				Artifacts: "archive",
			},
			{
				ID:        "package",
				Artifacts: "package",
			},
		},
	}
}

// Build configures a goreleaser build.
// https://goreleaser.com/customization/build/
func Build(dist string) config.Build {
	return config.Build{
		ID:     dist,
		Dir:    "_build",
		Binary: dist,
		BuildDetails: config.BuildDetails{
			Env:     []string{"CGO_ENABLED=0"},
			Flags:   []string{"-trimpath"},
			Ldflags: []string{"-s", "-w"},
		},
		Goos:   []string{"darwin", "linux", "windows"},
		Goarch: Architectures,
		Ignore: []config.IgnoredBuild{
			{Goos: "darwin", Goarch: "386"},
			{Goos: "windows", Goarch: "arm64"},
		},
	}
}

// Archive configures a goreleaser archive (tarball).
// https://goreleaser.com/customization/archive/
func Archive(dist string) config.Archive {
	return config.Archive{
		ID:           dist,
		NameTemplate: "{{ .Binary }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}{{ if .Mips }}_{{ .Mips }}{{ end }}",
		Builds:       []string{dist},
	}
}

// Package configures goreleaser to build a system package.
// https://goreleaser.com/customization/nfpm/
func Package(dist string) config.NFPM {
	return config.NFPM{
		ID:      dist,
		Builds:  []string{dist},
		Formats: []string{"apk", "deb", "rpm"},

		License:     "Apache 2.0",
		Description: fmt.Sprintf("%s distribution of the OpenTelemetry Collector", dist),
		Maintainer:  "Juraci Paixão Kröhling <distributions@kroehling.de>",

		NFPMOverridables: config.NFPMOverridables{
			PackageName: dist,
			Scripts: config.NFPMScripts{
				PreInstall:  "preinstall.sh",
				PostInstall: "postinstall.sh",
				PreRemove:   "preremove.sh",
			},
			Contents: files.Contents{
				{
					Source:      fmt.Sprintf("%s.service", dist),
					Destination: path.Join("/lib", "systemd", "system", fmt.Sprintf("%s.service", dist)),
				},
				{
					Source:      fmt.Sprintf("%s.conf", dist),
					Destination: path.Join("/etc", dist, fmt.Sprintf("%s.conf", dist)),
					Type:        "config|noreplace",
				},
				{
					Source:      "otelcol.yaml",
					Destination: path.Join("/etc", dist, "config.yaml"),
					Type:        "config",
				},
			},
		},
	}
}

func DockerImages(imagePrefixes []string, dist string) (r []config.Docker) {
	for _, arch := range Architectures {
		r = append(r, DockerImage(imagePrefixes, dist, arch))
	}
	return
}

// DockerImage configures goreleaser to build a container image.
// https://goreleaser.com/customization/docker/
func DockerImage(imagePrefixes []string, dist, arch string) config.Docker {
	var imageTemplates []string
	for _, prefix := range imagePrefixes {
		imageTemplates = append(
			imageTemplates,
			fmt.Sprintf("%s/%s:{{ .Version }}-%s", prefix, dist, arch),
		)
	}

	label := func(name, template string) string {
		return fmt.Sprintf("--label=org.opencontainers.image.%s={{%s}}", name, template)
	}

	return config.Docker{
		ImageTemplates: imageTemplates,
		Dockerfile:     "Dockerfile",

		Use: "buildx",
		BuildFlagTemplates: []string{
			"--pull",
			fmt.Sprintf("--platform=linux/%s", arch),
			label("created", ".Date"),
			label("name", ".ProjectName"),
			label("revision", ".FullCommit"),
			label("version", ".Version"),
			label("source", ".GitURL"),
		},
		Files:  []string{"otelcol.yaml"},
		Goos:   "linux",
		Goarch: arch,
	}
}

// DockerManifest configures goreleaser to build a multi-arch container image manifest.
// https://goreleaser.com/customization/docker_manifest/
func DockerManifest(imagePrefixes []string, dist string) (manifests []config.DockerManifest) {
	for _, prefix := range imagePrefixes {
		var imageTemplates []string
		for _, arch := range Architectures {
			imageTemplates = append(
				imageTemplates,
				fmt.Sprintf("%s/%s:{{ .Version }}-%s", prefix, dist, arch),
			)
		}

		manifests = append(manifests, config.DockerManifest{
			NameTemplate:   fmt.Sprintf("%s/%s:{{ .Version }}", prefix, dist),
			ImageTemplates: imageTemplates,
		})
	}
	return
}
