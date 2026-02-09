package chart

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"gofitness/src/model"

	"github.com/wcharczuk/go-chart/v2"
)

// GenerateProgressChart — строит график прогресса с двумя линиями (PNG)
func GenerateProgressChart(points []model.ProgressPoint, exerciseName string) (*bytes.Buffer, error) {
	if len(points) < 2 {
		log.Printf("недостаточно данных: %d точек", len(points))
		return nil, fmt.Errorf("недостаточно данных")
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Date.Before(points[j].Date)
	})

	var dates []time.Time
	var weights []float64
	var reps []float64

	startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i, p := range points {
		currentDate := startDate.AddDate(0, 0, i) // +1 день
		dates = append(dates, currentDate)

		// dates = append(dates, p.Date)
		weights = append(weights, p.AvgWeight)
		reps = append(reps, p.AvgReps)
	}

	// Выводим весь слайс dates в конце
	log.Println("=== Все dates ===")
	for i, d := range dates {
		log.Printf("dates[%d] = %s", i, d.Format("02.01.2006"))
	}

	log.Println("=== weights ===")
	log.Printf("%v", weights)

	log.Println("=== reps ===")
	log.Printf("%v", reps)

	// Отладка
	log.Println("=== Отладка данных ===")
	log.Printf("Точек: %d", len(dates))
	for i := 0; i < len(dates); i++ {
		log.Printf("[%d] %s | вес=%.1f | повторы=%.1f", i, dates[i].Format("02.01"), weights[i], reps[i])
	}

	minWeight, maxWeight := minMax(weights)
	minReps, maxReps := minMax(reps)
	minY := math.Min(minWeight, minReps) * 0.8
	maxY := math.Max(maxWeight, maxReps) * 1.2

	if maxY-minY < 1 {
		maxY = minY + 10 // чтобы не было 0 delta
	}

	graph := chart.Chart{
		Width:  900,
		Height: 600,
		Canvas: chart.Style{
			Padding: chart.Box{Top: 80, Left: 80, Right: 80, Bottom: 80},
		},
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:  "Значение",
			Range: &chart.ContinuousRange{Min: minY, Max: maxY}, // ← ЯВНО задаём диапазон
			Style: chart.Style{
				FontSize: 12,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dates,
				YValues: reps,
			},
			chart.TimeSeries{
				Name:    "Вес (кг)",
				XValues: dates,
				YValues: weights,
				Style: chart.Style{
					StrokeWidth: 4,
					StrokeColor: chart.ColorBlue,
					FillColor:   chart.ColorBlue.WithAlpha(30),
					DotWidth:    8,
					DotColor:    chart.ColorBlue,
				},
			},
			chart.TimeSeries{
				Name:    "Повторения",
				XValues: dates,
				YValues: reps,
				Style: chart.Style{
					StrokeWidth: 4,
					StrokeColor: chart.ColorOrange,
					FillColor:   chart.ColorOrange.WithAlpha(30),
					DotWidth:    8,
					DotColor:    chart.ColorOrange,
				},
			},
		},
	}

	graph.Elements = []chart.Renderable{chart.Legend(&graph)}

	buf := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buf)
	
	if err != nil {
		log.Printf("ошибка рендеринга: %w", err)
		return nil, fmt.Errorf("ошибка рендеринга: %w", err)
	}

	log.Printf("PNG создан, размер: %d байт", buf.Len())
	return buf, nil
}

func minMax(vals []float64) (min, max float64) {
    if len(vals) == 0 {
        return 0, 0
    }
    min = vals[0]
    max = vals[0]
    for _, v := range vals {
        if v < min {
            min = v
        }
        if v > max {
            max = v
        }
    }
    return min, max
}