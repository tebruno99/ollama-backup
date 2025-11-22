/*
Copyright © 2025 Tom Bruno <tbruno@tombruno.dev>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"archive/tar"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const ManifestLibraryPath = "manifests/registry.ollama.ai/library"

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup the manifest and blobs for an ollama model",
	Long: `Collects the manifest and blocks for an ollama model and outputs to tar file or stdout. For example:

	ollama-backup -f codestral-latest.tar codestral:latest
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ollamaPath := viper.GetString("ollama-models")
		librarypath := filepath.Join(ollamaPath, ManifestLibraryPath)
		outFile, err := cmd.Flags().GetString("file")
		if err != nil {
			log.Fatal(err)
		}

		outWriter := io.Writer(os.Stdout)
		if outFile != "-" {
			fileOutput, err := os.Create(outFile)
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				if closeErr := fileOutput.Close(); closeErr != nil {
					// Log or handle the close error
					log.Printf("Error closing output file: %v", closeErr)
				}
			}()
		}

		modelArgs := strings.Split(args[0], ":")
		modelName := modelArgs[0]
		modelVersion := modelArgs[1]
		modelPath := filepath.Join(librarypath, modelName, modelVersion)
		if modelName == "" || modelVersion == "" {
			log.Fatalf("model name and version must be specified")
		}

		jsonbt, err := os.ReadFile(modelPath)
		if err != nil {
			log.Fatalf("error reading model manifest %s: %v", modelPath, err)
		}
		var manifest OllamaModelManifest
		err = json.Unmarshal(jsonbt, &manifest)
		if err != nil {
			log.Fatalf("error parsing model manifest %s: %v", modelPath, err)
		}

		tw := tar.NewWriter(outWriter)

		err = addToArchive(tw, ollamaPath, filepath.Join(ManifestLibraryPath, modelName, modelVersion))
		if err != nil {
			log.Fatalf("error adding model to archive: %v", err)
		}

		err = addToArchive(tw, ollamaPath, filepath.Join("blobs", shaString(manifest.Config.Digest)))
		if err != nil {
			log.Fatalf("error adding model digest to archive %s: %v", filepath.Join("blobs", shaString(manifest.Config.Digest)), err)
		}

		for _, blob := range manifest.Layers {
			name := shaString(blob.Digest)
			err = addToArchive(tw, ollamaPath, filepath.Join("blobs", shaString(name)))
			if err != nil {
				log.Fatalf("error adding model digest to archive %s: %v", filepath.Join("blobs", shaString(blob.Digest)), err)
			}
		}

		err = tw.Close()
		if err != nil {
			log.Fatalf("error closing tar writer: %v", err)
		}
	},
}

func shaString(s string) string {
	return strings.ReplaceAll(s, ":", "-")
}

func addToArchive(tw *tar.Writer, prefix string, filename string) error {
	f, err := os.Open(filepath.Join(prefix, filename))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// Log or handle the close error
			log.Printf("Error closing file: %v", closeErr)
		}
	}()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	log.Printf("Adding %s to archive. Size: %d", filename, fi.Size())

	h, err := tar.FileInfoHeader(fi, fi.Name())
	if err != nil {
		return err
	}
	h.Name = filename
	if err := tw.WriteHeader(h); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVarP(&outFile, "file", "f", "-", "Read the archive from or write the archive to the specified file.  The filename can be - for standard input or standard output.")
}

type OllamaModelManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
	} `json:"layers"`
}
