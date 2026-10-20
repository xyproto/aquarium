package main

import "math/rand"

// randColor replaces digit placeholders 1-9 in a color mask with random
// color letters, matching the Perl rand_color() function.
// Digit '4' (eye) is always mapped to 'W' (white/gray) first.
func randColor(colorMask []string) []string {
	availColors := []rune{'c', 'C', 'r', 'R', 'y', 'Y', 'b', 'B', 'g', 'G', 'm', 'M'}

	assignment := make(map[rune]rune)
	for d := '1'; d <= '9'; d++ {
		assignment[d] = availColors[rand.Intn(len(availColors))]
	}
	assignment['4'] = 'W' // eye is always white/gray

	result := make([]string, len(colorMask))
	for i, line := range colorMask {
		runes := []rune(line)
		for j, ch := range runes {
			if c, ok := assignment[ch]; ok {
				runes[j] = c
			}
		}
		result[i] = string(runes)
	}
	return result
}
