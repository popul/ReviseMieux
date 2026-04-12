package app

import "testing"

func TestStripAccents(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"french accents", "éèêë àâä ùûü îï ôö ç", "eeee aaa uuu ii oo c"},
		{"greek rho", "ρ = m / V", "rho = m / V"},
		{"greek pi", "π × r²", "pi × r2"},
		{"superscripts", "kg/m³ cm²", "kg/m3 cm2"},
		{"no accents", "hello world", "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAccents(tt.input)
			if got != tt.want {
				t.Errorf("stripAccents(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripPunctuation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"apostrophe", "d'identifier", "d identifier"},
		{"curly apostrophe", "l\u2019eau", "l eau"},
		{"trailing period", "inconnu.", "inconnu"},
		{"multiple punctuation", "Bonjour, le monde! (test)", "Bonjour le monde test"},
		{"preserves spaces", "mot1  mot2   mot3", "mot1 mot2 mot3"},
		{"no punctuation", "hello world", "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripPunctuation(tt.input)
			if got != tt.want {
				t.Errorf("stripPunctuation(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestScoreByKeywords(t *testing.T) {
	tests := []struct {
		name           string
		expectedAnswer string
		keywords       string
		userAnswer     string
		wantScore      float64
	}{
		{
			name:           "Q3: rho formula with greek symbol keywords",
			expectedAnswer: "ρ = m / V",
			keywords:       "formule, masse volumique, ρ, m, V",
			userAnswer:     "rho est la formule masse volumique egale masse divise volume",
			wantScore:      1.0,
		},
		{
			name:           "Q8: identifier with apostrophe and trailing period",
			expectedAnswer: "La masse volumique permet d'identifier un matériau inconnu.",
			keywords:       "",
			userAnswer:     "la masse volumique permet identifier un materiau inconnu",
			wantScore:      1.0,
		},
		{
			name:           "Q2: unite SI exact match",
			expectedAnswer: "L'unité SI de la masse volumique est le kilogramme par mètre cube (kg/m³).",
			keywords:       "",
			userAnswer:     "kilogramme par metre cube kg/m3 unite SI masse volumique",
			wantScore:      1.0,
		},
		{
			name:           "Q5: flotte densite",
			expectedAnswer: "Un corps flotte si sa densité est inférieure à 1.",
			keywords:       "",
			userAnswer:     "un corps flotte si sa densite est inferieure a 1",
			wantScore:      1.0,
		},
		{
			name:           "single char keywords kept",
			expectedAnswer: "a = b",
			keywords:       "a, b",
			userAnswer:     "a equals b",
			wantScore:      1.0,
		},
		{
			name:           "completely wrong answer",
			expectedAnswer: "La masse volumique est le rapport de la masse sur le volume.",
			keywords:       "masse volumique, rapport, masse, volume",
			userAnswer:     "je ne sais pas du tout",
			wantScore:      0.0,
		},
		{
			name:           "partial match 40-70%",
			expectedAnswer: "La masse volumique est le rapport de la masse sur le volume.",
			keywords:       "masse volumique, rapport, masse, volume",
			userAnswer:     "la masse volumique est importante",
			wantScore:      0.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreByKeywords(tt.expectedAnswer, tt.keywords, tt.userAnswer)
			if got != tt.wantScore {
				t.Errorf("scoreByKeywords() = %.2f, want %.2f", got, tt.wantScore)
			}
		})
	}
}
