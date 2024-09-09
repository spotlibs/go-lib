package currency

import (
	"strings"
)

// Define maps for number words in Indonesian
var (
	units  = []string{"", "satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan"}
	teens  = []string{"sepuluh", "sebelas", "dua belas", "tiga belas", "empat belas", "lima belas", "enam belas", "tujuh belas", "delapan belas", "sembilan belas"}
	tens   = []string{"", "", "dua puluh", "tiga puluh", "empat puluh", "lima puluh", "enam puluh", "tujuh puluh", "delapan puluh", "sembilan puluh"}
	scales = []string{"", "ribu", "juta", "miliar"} // Add more scales if needed (e.g., triliun)
)

// Helper function to convert numbers below 1000 in Indonesian
func convertHundreds(n int) string {
	if n == 0 {
		return ""
	}
	result := ""

	if n >= 100 {
		if n/100 == 1 {
			result += "seratus"
		} else {
			result += units[n/100] + " ratus"
		}
		n %= 100
		if n > 0 {
			result += " "
		}
	}
	if n >= 20 {
		result += tens[n/10]
		n %= 10
		if n > 0 {
			result += " " + units[n]
		}
	} else if n >= 10 {
		result += teens[n-10]
	} else if n > 0 {
		if n == 1 && result == "" {
			result += "satu"
		} else {
			result += units[n]
		}
	}
	return result
}

// Main function to convert any integer to words in Indonesian
func doIntToWordsIndonesian(n int) string {
	if n == 0 {
		return "nol"
	}

	if n < 0 {
		return "minus " + doIntToWordsIndonesian(-n)
	}

	result := ""
	scaleIndex := 0

	for n > 0 {
		if n%1000 != 0 {
			part := convertHundreds(n % 1000)

			// Special case handling for "seribu" instead of "satu ribu"
			if scaleIndex == 1 && part == "satu" {
				part = "seribu"
			} else if scaleIndex > 0 && part != "" {
				part += " " + scales[scaleIndex]
			}

			if result != "" {
				result = part + " " + result
			} else {
				result = part
			}
		}
		n /= 1000
		scaleIndex++
	}

	return strings.TrimSpace(result)
}

func IntToWordsIndonesian(n int) string {
	return doIntToWordsIndonesian(n)
}
