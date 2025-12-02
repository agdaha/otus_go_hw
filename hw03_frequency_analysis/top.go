package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(text string) []string {
	tokens := strings.Fields(text)

	dict := make(map[string]int)

	for _, token := range tokens {
		key := strings.Trim(strings.ToLower(token), ".,?;:!")
		if key == "-" {
			continue
		}
		dict[key]++
	}

	sotredWords := make([]string, 0, len(dict))

	for k := range dict {
		sotredWords = append(sotredWords, k)
	}

	sort.Slice(sotredWords, func(i, j int) bool {
		if dict[sotredWords[i]] == dict[sotredWords[j]] {
			return sotredWords[i] < sotredWords[j]
		}
		return dict[sotredWords[i]] > dict[sotredWords[j]]
	})

	if len(sotredWords) < 10 {
		return sotredWords
	}

	return sotredWords[:10]
}
