package cmd

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

const (
	dockerfile       = "Dockerfile"
	golangUrl        = "https://golang.org/dl/?mode=json&include=all"
	generatedWarning = `#
# NOTE: THIS DOCKERFILE IS GENERATED VIA "task.go"
#
# PLEASE DO NOT EDIT IT DIRECTLY.
#`
)

type (
	Records struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
		Files   []File `json:"files"`
	}

	File struct {
		Filename string `json:"filename"`
		Os       string `json:"os"`
		Arch     string `json:"arch"`
		Version  string `json:"version"`
		Sha256   string `json:"sha256"`
		Size     int    `json:"size"`
		Kind     string `json:"kind"`
	}
)

func (v *Records) GetStableVersions() []File {
	if v.Stable {
		return v.GetVersions(v.Version)
	}

	return nil
}

func (v *Records) GetVersions(version string) []File {
	var files []File

	for _, f := range v.Files {
		if f.Version == version && f.Kind == "archive" {
			files = append(files, f)
		}
	}

	return files
}

func (v *Records) GetSpecificArch(arch string, os string) *File {
	for _, f := range v.GetStableVersions() {
		if f.Arch == arch && f.Os == os && f.Sha256 != "" {
			return &f
		}
	}

	return &File{}
}

func getGolangVersions() ([]Records, error) {
	resp, err := http.Get(golangUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			panic(err)
		}
	}(resp.Body)

	var versions []Records
	if err = json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// executeCmd represents the execute command
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		versions, err := getGolangVersions()
		cobra.CheckErr(err)

		for _, v := range versions {
			file := v.GetSpecificArch("amd64", "linux")
			if file.Size != 0 {
				cmd.Printf("Version: %s\tFilename: %s\tOS: %s Arch: %s Size: %d\tSHA256: %s\n", v.Version, file.Filename, file.Os, file.Arch, file.Size, file.Sha256)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(executeCmd)
}
