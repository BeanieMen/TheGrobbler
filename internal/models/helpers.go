package models

//very madarchod stuff
func MergeGrobbles(existing, new []Grobble) []Grobble {
	merged := make([]Grobble, 0, len(existing)+len(new))
	i, j := 0, 0

	for i < len(existing) && j < len(new) {
		// If they are equal, take one (new one) and advance both
		if existing[i].PlayedAt == new[j].PlayedAt {
			merged = append(merged, new[j])
			i++
			j++
		} else if existing[i].PlayedAt > new[j].PlayedAt {
			// existing[i].PlayedAt is lexically greater, meaning it's newer
			merged = append(merged, existing[i])
			i++
		} else {
			merged = append(merged, new[j])
			j++
		}
	}

	for ; i < len(existing); i++ {
		merged = append(merged, existing[i])
	}
	for ; j < len(new); j++ {
		merged = append(merged, new[j])
	}
	return merged
}
