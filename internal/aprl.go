// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/Azure/azqr/internal/scanners"
	"gopkg.in/yaml.v3"
)

//go:embed aprl/azure-resources/**/**/*.yaml
//go:embed aprl/azure-resources/**/**/kql/*.kql
//go:embed aprl/azure-specialized-workloads/**/*.yaml
//go:embed aprl/azure-specialized-workloads/**/kql/*.kql
var embededFiles embed.FS

// GetAprlRecommendations
func GetAprlRecommendations() map[string]map[string]scanners.AprlRecommendation {
	r := map[string]map[string]scanners.AprlRecommendation{}

	fsys, err := fs.Sub(embededFiles, "aprl/azure-resources")
	if err != nil {
		return nil
	}

	q := map[string]string{}
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".kql") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			fileName := strings.TrimSuffix(d.Name(), ".kql")
			q[fileName] = string(content)
		}
		return nil
	})
	if err != nil {
		return nil
	}

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".yaml") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			var recommendations []scanners.AprlRecommendation
			err = yaml.Unmarshal(content, &recommendations)
			if err != nil {
				return err
			}

			for _, recommendation := range recommendations {
				t := strings.ToLower(recommendation.ResourceType)
				if _, ok := r[t]; !ok {
					r[t] = map[string]scanners.AprlRecommendation{}
				}

				if i, ok := q[recommendation.RecommendationID]; ok {
					recommendation.GraphQuery = i
				}

				r[t][recommendation.RecommendationID] = recommendation
			}

		}
		return nil
	})
	if err != nil {
		return nil
	}

	return r
}
