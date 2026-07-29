package diff

import (
	"sort"
	
	"setupper/internal/manifest"
)

type Result struct {
	Unmanaged []manifest.Resource
	Missing   []manifest.Resource
	Matching  []manifest.Resource
}

// Compare returns a diff between observed and desired manifests
func Compare(observed *manifest.ObservedManifest, desired *manifest.DesiredManifest) Result {
	var result Result

	for key, desRes := range desired.Resources {
		if obsRes, exists := observed.Resources[key]; exists {
			result.Matching = append(result.Matching, obsRes)
		} else {
			result.Missing = append(result.Missing, desRes)
		}
	}

	for key, obsRes := range observed.Resources {
		if _, exists := desired.Resources[key]; !exists {
			result.Unmanaged = append(result.Unmanaged, obsRes)
		}
	}

	sortResources(result.Unmanaged)
	sortResources(result.Missing)
	sortResources(result.Matching)

	return result
}

func sortResources(res []manifest.Resource) {
	sort.Slice(res, func(i, j int) bool {
		if res[i].Type == res[j].Type {
			return res[i].Name < res[j].Name
		}
		return res[i].Type < res[j].Type
	})
}
