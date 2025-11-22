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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all the installed ollama manifests",
	Long: `Reads the ollama modeles directory and lists out each model & version currently installed. For example:
	
		./ollama-backup list
`,
	Run: func(cmd *cobra.Command, args []string) {
		ollamaPath := viper.GetString("ollama-models")
		librarypath := filepath.Join(ollamaPath, "manifests/registry.ollama.ai/library/")
		modelMap := make(map[string][]string, 0)
		fmt.Printf("--- %s ---\n", librarypath)

		dir, err := os.ReadDir(librarypath)
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range dir {
			if entry.IsDir() {
				vers, err := os.ReadDir(filepath.Join(librarypath, entry.Name()))
				if err != nil {
					log.Fatal(err)
				}
				for _, ver := range vers {
					if !ver.IsDir() {
						modelMap[entry.Name()] = append(modelMap[entry.Name()], ver.Name())
					}
				}
			}
		}
		for name, versions := range modelMap {
			fmt.Printf("%s: %s\n", name, strings.Join(versions, ","))
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
