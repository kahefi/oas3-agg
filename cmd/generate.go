/*
Copyright © 2020 Karl-heinz Fiebig

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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate SEARCH_PATH",
	Short: "Aggregates all OAS3 specifications found in files within the SEARCH_PATH",
	Long: `'generate' automatically aggregates OpenAPI 3 Speification (OAS3) from all files
	in the SEARCH_PATH and compiles a single specification document for your API.

	YAML Specifications in files must be enclosed in a block comment as follows:
	/* OAS3-Specification
	<YOUR API SPECIFICATION IN YAML FORMAT HERE>
	*/
	TODO`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generate called")
		outputFile, err := cmd.Flags().GetString("output")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		rootDir := args[0]
		fmt.Println("Root directory: " + rootDir)
		specs, err := extractSpecs(rootDir)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		globalSpec := make(map[interface{}]interface{})
		for _, spec := range specs {
			globalSpec, err = mergeMaps(globalSpec, spec)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		// Write specs to file
		globalSpecBytes, err := yaml.Marshal(globalSpec)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Writing specification to", outputFile)
		err = ioutil.WriteFile(outputFile, globalSpecBytes, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("-----")
		fmt.Println("Finished.")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	// Flag for the output file
	generateCmd.Flags().StringP("output", "o", "./oas3_spec.yaml", "the output file to store the specification.")
}

func extractSpecs(rootDir string) ([]map[interface{}]interface{}, error) {
	// Gather up specficaitions from individual go files
	finalSpecs := []map[interface{}]interface{}{}
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		// Skip non-go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Attempt to open file
		fmt.Println("Processing " + path + "...")
		file, err := os.Open(path)
		if err != nil {
			fmt.Println("-> Failed to open file: ", err)
			return err
		}
		defer file.Close()
		// Read out content
		contentBytes, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("-> Failed to read file: ", err)
			return err
		}
		content := string(contentBytes)

		// Split content around the beginning marker and cut trail from ending marker
		specCandidates := strings.Split(content, "/*** OAS3-Specification")
		c := 0
		for _, candidate := range specCandidates {
			trimIdx := strings.Index(candidate, "\n***/")
			// Only candidates with ending marker are accepted
			if trimIdx != -1 {
				yamlSpec := strings.TrimSpace(candidate[:trimIdx])
				// Marshal specification part into a YAML map
				specMap := make(map[interface{}]interface{})
				err = yaml.Unmarshal([]byte(yamlSpec), &specMap)
				if err != nil {
					fmt.Println("-> Failed to encode spec into YAML: ", err)
					return err
				}
				// Aggregate spec
				finalSpecs = append(finalSpecs, specMap)
				c++
			}
		}
		// Print out results
		if c > 0 {
			fmt.Println("-> Extracted specs: ", c)
		} else {
			fmt.Println("-> No specs found")
		}
		return nil
	})
	// Print and return final results
	fmt.Println("-----")
	fmt.Println("Aggregated", len(finalSpecs), "specifications")
	return finalSpecs, err
}

func mergeMaps(m1 map[interface{}]interface{}, m2 map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	// Iterate through keys of m1 and insert values into m2
	for k := range m1 {
		// If key is not in m2, we can the key value pair directly
		if _, ok := m2[k]; !ok {
			m2[k] = m1[k]
			continue
		}
		// If value in m1 is a slice, append it to the corresponding slice of m2
		if sliceM1, ok := m1[k].([]interface{}); ok {
			sliceM2, ok := m2[k].([]interface{})
			if !ok {
				return nil, errors.New(fmt.Sprintf("Merging failed, expected both entries for \"%s\" to be lists.", k))
			}
			m2[k] = append(sliceM2, sliceM1...)
			continue
		}

		// If entry value in m1 is a map, recurse into submaps and merge them
		if mapM1, ok := m1[k].(map[interface{}]interface{}); ok {
			mapM2, ok := m2[k].(map[interface{}]interface{})
			if !ok {
				return nil, errors.New(fmt.Sprintf("Merging failed, expected both entries for \"%s\" to be objects.", k))
			}
			m1m2, err := mergeMaps(mapM1, mapM2)
			if err != nil {
				return nil, err
			}
			m2[k] = m1m2
			continue
		}

		// If entry value is not a slice and not a map and already in m2, we have a duplicate
		return nil, errors.New(fmt.Sprintf("Specification invalid, duplicate entry for \"%s\" encountered.", k))
	}

	return m2, nil
}
