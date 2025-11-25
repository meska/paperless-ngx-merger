package similarity

import (
	"strings"
	"unicode"
)

// SimilarityGroup rappresenta un gruppo di elementi con testo simile
type SimilarityGroup struct {
	Representative string
	Items          []SimilarItem
}

// SimilarItem rappresenta un elemento simile
type SimilarItem struct {
	ID   int
	Name string
}

// levenshteinDistance calcola la distanza di Levenshtein tra due stringhe
func levenshteinDistance(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)
	
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Crea una matrice per la programmazione dinamica
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Calcola la distanza
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // eliminazione
				matrix[i][j-1]+1,      // inserimento
				matrix[i-1][j-1]+cost, // sostituzione
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min restituisce il minimo tra tre interi
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// normalizeString normalizza una stringa per il confronto
func normalizeString(s string) string {
	// Rimuove spazi multipli e normalizza
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	
	// Rimuove punteggiatura
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// CalculateSimilarity calcola la similarità tra due stringhe (0.0 = diversi, 1.0 = uguali)
func CalculateSimilarity(s1, s2 string) float64 {
	norm1 := normalizeString(s1)
	norm2 := normalizeString(s2)
	
	if norm1 == norm2 {
		return 1.0
	}
	
	maxLen := len(norm1)
	if len(norm2) > maxLen {
		maxLen = len(norm2)
	}
	
	if maxLen == 0 {
		return 1.0
	}
	
	distance := levenshteinDistance(norm1, norm2)
	return 1.0 - float64(distance)/float64(maxLen)
}

// FindSimilarGroups trova gruppi di elementi simili
// threshold: soglia di similarità (0.0-1.0), più alta = più simile richiesto
func FindSimilarGroups(items []SimilarItem, threshold float64) []SimilarityGroup {
	if threshold < 0.0 {
		threshold = 0.0
	}
	if threshold > 1.0 {
		threshold = 1.0
	}

	groups := make([]SimilarityGroup, 0)
	used := make(map[int]bool)

	for i := 0; i < len(items); i++ {
		if used[items[i].ID] {
			continue
		}

		group := SimilarityGroup{
			Representative: items[i].Name,
			Items:          []SimilarItem{items[i]},
		}
		used[items[i].ID] = true

		// Trova elementi simili
		for j := i + 1; j < len(items); j++ {
			if used[items[j].ID] {
				continue
			}

			similarity := CalculateSimilarity(items[i].Name, items[j].Name)
			if similarity >= threshold {
				group.Items = append(group.Items, items[j])
				used[items[j].ID] = true
			}
		}

		// Aggiungi il gruppo solo se ha più di un elemento
		if len(group.Items) > 1 {
			groups = append(groups, group)
		}
	}

	return groups
}
