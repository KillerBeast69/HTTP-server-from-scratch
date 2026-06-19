package main

import (
	"strings"
)

//var profane_words = []string{"kerfulle", "sharbert", "fornax"}

func censor(input string) string {
	words := strings.Split(input, " ")

	for i, word := range words {
		lower_word := strings.ToLower(word)
		if lower_word == "kerfuffle" || lower_word == "sharbert" || lower_word == "fornax" {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
