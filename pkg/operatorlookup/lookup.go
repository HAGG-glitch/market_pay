// Package operatorlookup ports the operator data from:
// https://github.com/saidubundukamara/mobile-operator-lookup
// (npm package: mobile-operator-lookup)
//
// Maintainers: when updating prefixes, check the upstream repo first.
package operatorlookup

import "strings"

type Result struct {
	Company     string
	MobileMoney string
	Slug        string
	CountryCode string
	MonimeCode  string
}

type operator struct {
	prefixes   []string
	company    string
	mobileMoney string
	slug       string
	monimeCode string
}

var slOperators = []operator{
	{
		prefixes:   []string{"72", "73", "74", "75", "76", "78", "79"},
		company:    "Orange",
		mobileMoney: "Orange Money",
		slug:       "orange-money",
		monimeCode: "m17",
	},
	{
		prefixes:   []string{"88", "77", "90", "99", "30", "33"},
		company:    "Africell",
		mobileMoney: "Afrimoney",
		slug:       "afrimoney",
		monimeCode: "m18",
	},
	{
		prefixes:   []string{"31", "34"},
		company:    "Qcell",
		mobileMoney: "Qcell Money",
		slug:       "qcell-money",
		monimeCode: "m13",
	},
}

// Lookup identifies the mobile operator for a given phone number.
// It strips the country code (+232), then matches the remaining prefix
// against known operator prefixes for Sierra Leone.
func Lookup(phone string) *Result {
	normalized := phone
	if strings.HasPrefix(normalized, "+232") {
		normalized = normalized[4:]
	} else if strings.HasPrefix(normalized, "232") {
		normalized = normalized[3:]
	}

	normalized = strings.TrimPrefix(normalized, "0")

	if len(normalized) == 0 {
		return nil
	}

	for _, op := range slOperators {
		for _, prefix := range op.prefixes {
			if strings.HasPrefix(normalized, prefix) {
				return &Result{
					Company:     op.company,
					MobileMoney: op.mobileMoney,
					Slug:        op.slug,
					CountryCode: "+232",
					MonimeCode:  op.monimeCode,
				}
			}
		}
	}

	return nil
}

// LookupProviderID returns just the Monime provider ID for a phone number.
func LookupProviderID(phone string) string {
	result := Lookup(phone)
	if result == nil {
		return ""
	}
	return result.MonimeCode
}
