package recommender

import (
	"setupper/internal/manifest"
	"setupper/internal/profiles"
)

type Recommendation struct {
	Profile     profiles.Profile
	TriggeredBy []string
}

type Recommender struct {
	profiles []profiles.Profile
}

func New(p []profiles.Profile) *Recommender {
	return &Recommender{profiles: p}
}

// Recommend returns a list of profiles that match the scan evidence
func (r *Recommender) Recommend(observed *manifest.ObservedManifest) []Recommendation {
	var recs []Recommendation

	for _, p := range r.profiles {
		var matchedTriggers []string
		for _, trigger := range p.Triggers {
			if _, exists := observed.Resources[trigger]; exists {
				matchedTriggers = append(matchedTriggers, trigger)
			}
		}

		if len(matchedTriggers) > 0 {
			recs = append(recs, Recommendation{
				Profile:     p,
				TriggeredBy: matchedTriggers,
			})
		}
	}

	return recs
}
