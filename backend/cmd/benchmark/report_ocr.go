package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/popul/revisemieux/internal/benchmark"
)

func generateOCRReport(resultsDir, runDir, outputPath string) error {
	dir, err := resolveRunDir(resultsDir, runDir)
	if err != nil {
		return err
	}

	summaries, err := loadOCRSummaries(filepath.Join(dir, "summary.json"))
	if err != nil {
		return fmt.Errorf("load OCR summaries from %s: %w", dir, err)
	}

	if len(summaries) == 0 {
		return fmt.Errorf("no OCR results found in %s", dir)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].AvgCompositeScore > summaries[j].AvgCompositeScore
	})

	data := buildOCRReportData(summaries, dir)

	if outputPath == "" {
		outputPath = filepath.Join(dir, "report.html")
	}
	return renderOCRHTML(data, outputPath)
}

func loadOCRSummaries(path string) ([]benchmark.OCRRunSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var summaries []benchmark.OCRRunSummary
	if err := json.Unmarshal(data, &summaries); err != nil {
		return nil, err
	}
	return summaries, nil
}

type ocrReportData struct {
	Title         string
	RunDir        string
	Models        []string
	Summaries     []benchmark.OCRRunSummary
	MetricsRadar  template.JS
	MetricsBar    template.JS
	CostData      template.JS
	LatencyData   template.JS
	CompositeData template.JS
}

func buildOCRReportData(summaries []benchmark.OCRRunSummary, dir string) ocrReportData {
	models := make([]string, len(summaries))
	for i, s := range summaries {
		models[i] = s.Provider + "/" + s.Model
	}

	// Radar chart
	type radarDataset struct {
		Label string    `json:"label"`
		Data  []float64 `json:"data"`
	}
	var radarDatasets []radarDataset
	for _, s := range summaries {
		radarDatasets = append(radarDatasets, radarDataset{
			Label: s.Provider + "/" + s.Model,
			Data:  []float64{s.AvgDetectionScore, s.AvgTextAccuracy, s.AvgTypeAccuracy},
		})
	}
	radarJSON := mustJSON(map[string]interface{}{
		"labels":   []string{"Détection", "Texte", "Types"},
		"datasets": radarDatasets,
	})

	// Bar chart
	type barDataset struct {
		Label string    `json:"label"`
		Data  []float64 `json:"data"`
	}
	var barDatasets []barDataset
	for _, s := range summaries {
		barDatasets = append(barDatasets, barDataset{
			Label: s.Provider + "/" + s.Model,
			Data:  []float64{s.AvgDetectionScore, s.AvgTextAccuracy, s.AvgTypeAccuracy, s.AvgQualityScore},
		})
	}
	barJSON := mustJSON(map[string]interface{}{
		"labels":   []string{"Détection", "Texte", "Types", "Qualité globale"},
		"datasets": barDatasets,
	})

	// Cost
	type costEntry struct {
		Model     string  `json:"model"`
		TotalCost float64 `json:"total_cost"`
	}
	var costEntries []costEntry
	for _, s := range summaries {
		costEntries = append(costEntries, costEntry{
			Model:     s.Provider + "/" + s.Model,
			TotalCost: s.TotalCostUSD,
		})
	}
	costJSON := mustJSON(costEntries)

	// Latency
	type latencyEntry struct {
		Model     string `json:"model"`
		LatencyMs int64  `json:"latency_ms"`
	}
	var latencyEntries []latencyEntry
	for _, s := range summaries {
		latencyEntries = append(latencyEntries, latencyEntry{
			Model:     s.Provider + "/" + s.Model,
			LatencyMs: s.AvgLatencyMs,
		})
	}
	latencyJSON := mustJSON(latencyEntries)

	// Composite
	type compositeEntry struct {
		Model     string  `json:"model"`
		Quality   float64 `json:"quality"`
		Composite float64 `json:"composite"`
	}
	var compositeEntries []compositeEntry
	for _, s := range summaries {
		compositeEntries = append(compositeEntries, compositeEntry{
			Model:     s.Provider + "/" + s.Model,
			Quality:   s.AvgQualityScore,
			Composite: s.AvgCompositeScore,
		})
	}
	compositeJSON := mustJSON(compositeEntries)

	return ocrReportData{
		Title:         "Benchmark OCR — Révise Mieux",
		RunDir:        filepath.Base(dir),
		Models:        models,
		Summaries:     summaries,
		MetricsRadar:  template.JS(radarJSON),
		MetricsBar:    template.JS(barJSON),
		CostData:      template.JS(costJSON),
		LatencyData:   template.JS(latencyJSON),
		CompositeData: template.JS(compositeJSON),
	}
}

func renderOCRHTML(data ocrReportData, outputPath string) error {
	tmpl, err := template.New("ocr-report").Funcs(template.FuncMap{
		"shortModel": func(s string) string {
			parts := strings.Split(s, "/")
			if len(parts) > 1 {
				return parts[1]
			}
			return s
		},
		"pct": func(f float64) string {
			return fmt.Sprintf("%.1f%%", f*100)
		},
		"f4": func(f float64) string {
			return fmt.Sprintf("%.4f", f)
		},
		"f5": func(f float64) string {
			return fmt.Sprintf("%.5f", f)
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"add": func(a, b int) int {
			return a + b
		},
	}).Parse(ocrHTMLTemplate)
	if err != nil {
		return fmt.Errorf("parse OCR template: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute OCR template: %w", err)
	}

	fmt.Printf("Report generated: %s\n", outputPath)
	return nil
}

const ocrHTMLTemplate = `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js"></script>
    <style>
        :root {
            --bg: #0f172a;
            --surface: #1e293b;
            --surface-hover: #334155;
            --border: #334155;
            --text: #e2e8f0;
            --text-muted: #94a3b8;
            --accent: #3b82f6;
            --accent-light: #60a5fa;
            --success: #22c55e;
            --warning: #f59e0b;
            --danger: #ef4444;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg);
            color: var(--text);
            line-height: 1.6;
        }
        .container { max-width: 1400px; margin: 0 auto; padding: 2rem; }
        header {
            text-align: center;
            margin-bottom: 3rem;
            padding: 2rem;
            background: var(--surface);
            border-radius: 12px;
            border: 1px solid var(--border);
        }
        header h1 { font-size: 2rem; margin-bottom: 0.5rem; }
        header .subtitle { color: var(--text-muted); font-size: 0.95rem; }
        .grid { display: grid; gap: 1.5rem; }
        .grid-2 { grid-template-columns: 1fr 1fr; }
        .card {
            background: var(--surface);
            border-radius: 12px;
            padding: 1.5rem;
            border: 1px solid var(--border);
        }
        .card h2 {
            font-size: 1.1rem;
            margin-bottom: 1rem;
            color: var(--accent-light);
        }
        .card-full { grid-column: 1 / -1; }
        .chart-container {
            position: relative;
            width: 100%;
            height: 350px;
        }
        .chart-container-tall {
            position: relative;
            width: 100%;
            height: 450px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.9rem;
        }
        th, td {
            padding: 0.75rem 1rem;
            text-align: right;
            border-bottom: 1px solid var(--border);
        }
        th {
            color: var(--text-muted);
            font-weight: 600;
            font-size: 0.8rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        td:first-child, th:first-child { text-align: left; }
        tr:hover td { background: var(--surface-hover); }
        .rank { display: inline-block; width: 24px; height: 24px; border-radius: 50%;
                text-align: center; line-height: 24px; font-size: 0.75rem; font-weight: 700; margin-right: 0.5rem; }
        .rank-1 { background: #fbbf24; color: #1e293b; }
        .rank-2 { background: #94a3b8; color: #1e293b; }
        .rank-3 { background: #d97706; color: #1e293b; }
        .score-bar {
            display: inline-block;
            height: 6px;
            border-radius: 3px;
            background: var(--accent);
            vertical-align: middle;
            margin-left: 0.5rem;
        }
        .badge {
            display: inline-block;
            padding: 0.15rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 600;
        }
        .badge-good { background: rgba(34,197,94,0.2); color: var(--success); }
        .badge-warn { background: rgba(245,158,11,0.2); color: var(--warning); }
        .badge-bad { background: rgba(239,68,68,0.2); color: var(--danger); }
        @media (max-width: 900px) {
            .grid-2 { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
<div class="container">
    <header>
        <h1>{{.Title}}</h1>
        <div class="subtitle">Run : {{.RunDir}} &mdash; {{len .Summaries}} modèle(s) comparé(s)</div>
    </header>

    <!-- Ranking Table -->
    <div class="card card-full" style="margin-bottom: 1.5rem;">
        <h2>Classement global — OCR (images &rarr; blocs texte)</h2>
        <table>
            <thead>
                <tr>
                    <th>#</th>
                    <th>Modèle</th>
                    <th>Détection</th>
                    <th>Texte</th>
                    <th>Types</th>
                    <th>Qualité</th>
                    <th>Coût ($)</th>
                    <th>Latence (ms)</th>
                    <th>Erreurs</th>
                    <th>Score composite</th>
                </tr>
            </thead>
            <tbody>
            {{range $i, $s := .Summaries}}
                <tr>
                    <td>
                        {{if eq $i 0}}<span class="rank rank-1">1</span>
                        {{else if eq $i 1}}<span class="rank rank-2">2</span>
                        {{else if eq $i 2}}<span class="rank rank-3">3</span>
                        {{else}}{{add $i 1}}{{end}}
                    </td>
                    <td><strong>{{$s.Provider}}/{{$s.Model}}</strong></td>
                    <td>{{pct $s.AvgDetectionScore}}</td>
                    <td>{{pct $s.AvgTextAccuracy}}</td>
                    <td>{{pct $s.AvgTypeAccuracy}}</td>
                    <td>{{pct $s.AvgQualityScore}}</td>
                    <td>${{f5 $s.TotalCostUSD}}</td>
                    <td>{{$s.AvgLatencyMs}}</td>
                    <td>
                        {{pct $s.ErrorRate}}
                        {{if le $s.ErrorRate 0.0}}<span class="badge badge-good">OK</span>
                        {{else if le $s.ErrorRate 0.2}}<span class="badge badge-warn">Partiel</span>
                        {{else}}<span class="badge badge-bad">Élevé</span>{{end}}
                    </td>
                    <td>
                        <strong>{{f4 $s.AvgCompositeScore}}</strong>
                        <span class="score-bar" style="width: {{printf "%.0f" (mul $s.AvgCompositeScore 100)}}px;"></span>
                    </td>
                </tr>
            {{end}}
            </tbody>
        </table>
    </div>

    <div class="grid grid-2">
        <!-- Radar Chart -->
        <div class="card">
            <h2>Profil qualité OCR (radar)</h2>
            <div class="chart-container-tall">
                <canvas id="radarChart"></canvas>
            </div>
        </div>

        <!-- Bar Chart -->
        <div class="card">
            <h2>Métriques de qualité OCR</h2>
            <div class="chart-container-tall">
                <canvas id="barChart"></canvas>
            </div>
        </div>

        <!-- Cost Chart -->
        <div class="card">
            <h2>Coût total ($)</h2>
            <div class="chart-container">
                <canvas id="costChart"></canvas>
            </div>
        </div>

        <!-- Latency Chart -->
        <div class="card">
            <h2>Latence moyenne (ms)</h2>
            <div class="chart-container">
                <canvas id="latencyChart"></canvas>
            </div>
        </div>

        <!-- Composite Score Chart -->
        <div class="card card-full">
            <h2>Score composite vs. Score qualité</h2>
            <div class="chart-container">
                <canvas id="compositeChart"></canvas>
            </div>
        </div>
    </div>
</div>

<script>
const COLORS = [
    'rgba(59, 130, 246, 0.8)',
    'rgba(239, 68, 68, 0.8)',
    'rgba(34, 197, 94, 0.8)',
    'rgba(245, 158, 11, 0.8)',
    'rgba(168, 85, 247, 0.8)',
    'rgba(236, 72, 153, 0.8)',
    'rgba(20, 184, 166, 0.8)',
    'rgba(249, 115, 22, 0.8)',
];
const COLORS_BG = COLORS.map(c => c.replace('0.8', '0.2'));

Chart.defaults.color = '#94a3b8';
Chart.defaults.borderColor = '#334155';
Chart.defaults.font.family = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';

// Radar Chart
const radarData = {{.MetricsRadar}};
new Chart(document.getElementById('radarChart'), {
    type: 'radar',
    data: {
        labels: radarData.labels,
        datasets: radarData.datasets.map((ds, i) => ({
            label: ds.label,
            data: ds.data,
            borderColor: COLORS[i % COLORS.length],
            backgroundColor: COLORS_BG[i % COLORS.length],
            pointBackgroundColor: COLORS[i % COLORS.length],
            borderWidth: 2,
        }))
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            r: {
                beginAtZero: true,
                max: 1,
                ticks: { stepSize: 0.2, backdropColor: 'transparent' },
                grid: { color: '#334155' },
                angleLines: { color: '#334155' },
                pointLabels: { font: { size: 12 } }
            }
        },
        plugins: { legend: { position: 'bottom' } }
    }
});

// Bar Chart
const barData = {{.MetricsBar}};
new Chart(document.getElementById('barChart'), {
    type: 'bar',
    data: {
        labels: barData.labels,
        datasets: barData.datasets.map((ds, i) => ({
            label: ds.label,
            data: ds.data,
            backgroundColor: COLORS[i % COLORS.length],
            borderColor: COLORS[i % COLORS.length],
            borderWidth: 1,
        }))
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            y: { beginAtZero: true, max: 1, grid: { color: '#1e293b' } },
            x: { grid: { display: false } }
        },
        plugins: { legend: { position: 'bottom' } }
    }
});

// Cost Chart
const costData = {{.CostData}};
new Chart(document.getElementById('costChart'), {
    type: 'bar',
    data: {
        labels: costData.map(e => e.model),
        datasets: [{
            label: 'Coût total ($)',
            data: costData.map(e => e.total_cost),
            backgroundColor: costData.map((_, i) => COLORS[i % COLORS.length]),
            borderWidth: 0,
            borderRadius: 6,
        }]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        indexAxis: 'y',
        scales: {
            x: { beginAtZero: true, grid: { color: '#1e293b' } },
            y: { grid: { display: false } }
        },
        plugins: { legend: { display: false } }
    }
});

// Latency Chart
const latencyData = {{.LatencyData}};
new Chart(document.getElementById('latencyChart'), {
    type: 'bar',
    data: {
        labels: latencyData.map(e => e.model),
        datasets: [{
            label: 'Latence (ms)',
            data: latencyData.map(e => e.latency_ms),
            backgroundColor: latencyData.map((_, i) => COLORS[i % COLORS.length]),
            borderWidth: 0,
            borderRadius: 6,
        }]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        indexAxis: 'y',
        scales: {
            x: { beginAtZero: true, grid: { color: '#1e293b' } },
            y: { grid: { display: false } }
        },
        plugins: { legend: { display: false } }
    }
});

// Composite Chart
const compositeData = {{.CompositeData}};
new Chart(document.getElementById('compositeChart'), {
    type: 'bar',
    data: {
        labels: compositeData.map(e => e.model),
        datasets: [
            {
                label: 'Score qualité',
                data: compositeData.map(e => e.quality),
                backgroundColor: 'rgba(59, 130, 246, 0.6)',
                borderColor: 'rgba(59, 130, 246, 1)',
                borderWidth: 1,
                borderRadius: 6,
            },
            {
                label: 'Score composite',
                data: compositeData.map(e => e.composite),
                backgroundColor: 'rgba(34, 197, 94, 0.6)',
                borderColor: 'rgba(34, 197, 94, 1)',
                borderWidth: 1,
                borderRadius: 6,
            }
        ]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            y: { beginAtZero: true, max: 1, grid: { color: '#1e293b' } },
            x: { grid: { display: false } }
        },
        plugins: { legend: { position: 'bottom' } }
    }
});
</script>
</body>
</html>`
