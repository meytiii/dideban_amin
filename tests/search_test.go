package main

import (
	"strings"
	"testing"
)

func TestNormalizeForSearchAndMatching(t *testing.T) {
	testCases := []struct {
		name        string
		reportText  string
		searchQuery string
		shouldMatch bool
	}{
		{
			name:        "Exact match Persian بوعلی",
			reportText:  "سلام این مورد در مدرسه بوعلی مشاهده شده",
			searchQuery: "بوعلی",
			shouldMatch: true,
		},
		{
			name:        "Arabic Yeh in report vs Persian Yeh in query",
			reportText:  "سلام این مورد در مدرسه بوعلي مشاهده شده",
			searchQuery: "بوعلی",
			shouldMatch: true,
		},
		{
			name:        "Persian Yeh in report vs Arabic Yeh in query",
			reportText:  "سلام این مورد در مدرسه بوعلی مشاهده شده",
			searchQuery: "بوعلي",
			shouldMatch: true,
		},
		{
			name:        "ZWNJ (half-space) in search query",
			reportText:  "سلام این مورد در مدرسه بوعلی مشاهده شده",
			searchQuery: "بو‌علی",
			shouldMatch: true,
		},
		{
			name:        "ZWNJ in report text",
			reportText:  "سلام این مورد در مدرسه بو‌علی مشاهده شده",
			searchQuery: "بوعلی",
			shouldMatch: true,
		},
		{
			name:        "Arabic Kaf vs Persian Keheh",
			reportText:  "مدرسه کمال الملک",
			searchQuery: "كمال",
			shouldMatch: true,
		},
		{
			name:        "Tashkeel and Tanwin diacritics",
			reportText:  "گزارش مَدرَسَه بُوعلی",
			searchQuery: "مدرسه بوعلی",
			shouldMatch: true,
		},
		{
			name:        "Alef with maddah or hamza",
			reportText:  "آموزش و پرورش ناحیه یک",
			searchQuery: "اموزش",
			shouldMatch: true,
		},
		{
			name:        "Unmatched term",
			reportText:  "سلام این مورد در مدرسه بوعلی مشاهده شده",
			searchQuery: "دانشگاه تهران",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normReport := normalizeForSearch(tc.reportText)
			normQuery := normalizeForSearch(tc.searchQuery)
			matched := strings.Contains(normReport, normQuery)
			if matched != tc.shouldMatch {
				t.Errorf("[%s] Expected match=%v, got=%v (normReport=%q, normQuery=%q)",
					tc.name, tc.shouldMatch, matched, normReport, normQuery)
			}
		})
	}
}
