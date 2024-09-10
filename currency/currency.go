package currency

import (
	"fmt"
	"strconv"
	"strings"
)

// Define maps for number words in Indonesian
var (
	units  = [10]string{"", "satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan"}
	teens  = [10]string{"sepuluh", "sebelas", "dua belas", "tiga belas", "empat belas", "lima belas", "enam belas", "tujuh belas", "delapan belas", "sembilan belas"}
	tens   = [13]string{"", "", "dua puluh", "tiga puluh", "empat puluh", "lima puluh", "enam puluh", "tujuh puluh", "delapan puluh", "sembilan puluh"}
	scales = [4]string{"", "ribu", "juta", "miliar"} // Add more scales if needed (e.g., triliun)
)

// convertHundreds is helper function to convert numbers below 1000 in Indonesian
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

// doIntToWordsIndonesian is main helper function to convert any integer to words in Indonesian
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

// IntToWordsIndonesian is function to convert integer form currency to words spelling
// example : 1001 -> seribu satu
func IntToWordsIndonesian(n int) string {
	return doIntToWordsIndonesian(n)
}

// decimalToWordsIndonesian is helper function to convert the decimal part of a float to words in Indonesian
func decimalToWordsIndonesian(decimalPart string) string {
	result := ""
	for _, digit := range decimalPart {
		num, _ := strconv.Atoi(string(digit))
		result += units[num] + " "
	}
	return strings.TrimSpace(result)
}

// doFloatToWordsIndonesian is main function to convert any float to words in Indonesian
func doFloatToWordsIndonesian(f float64) string {
	// Split into integer and decimal parts
	parts := strings.Split(fmt.Sprintf("%.10g", f), ".")
	integerPart, _ := strconv.Atoi(parts[0])
	integerWords := doIntToWordsIndonesian(integerPart)

	// If there is a decimal part, convert it
	if len(parts) > 1 {
		decimalWords := decimalToWordsIndonesian(parts[1])
		return integerWords + " koma " + decimalWords
	}

	// If no decimal part, return just the integer words
	return integerWords
}

// FloatToWordsIndonesian is a function to convert float to words in Indonesian
// example 123.45 -> "seratus dua puluh tiga koma empat lima"
func FloatToWordsIndonesian(f float64) string {
	return doFloatToWordsIndonesian(f)
}
