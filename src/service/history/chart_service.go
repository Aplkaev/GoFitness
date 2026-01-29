package history

import (
	"bytes"
	"fmt"
	"log"
	"os"
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

	// Сортировка
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date.Before(points[j].Date)
	})

	var dates []time.Time
	var weights []float64
	var reps []float64

	for _, p := range points {
		dates = append(dates, p.Date)
		weights = append(weights, p.AvgWeight)
		reps = append(reps, p.AvgReps)
	}

	// Отладка
	log.Println("=== Отладка данных ===")
	log.Printf("Точек: %d", len(dates))
	for i := 0; i < len(dates); i++ {
		log.Printf("[%d] %s | вес=%.1f | повторы=%.1f", i, dates[i].Format("02.01"), weights[i], reps[i])
	}

	graph := chart.Chart{
		// XAxis: chart.XAxis{
		// 	TickPosition: chart.TickPositionBetweenTicks,
		// 	ValueFormatter: func(v interface{}) string {
		// 		typed := v.(float64)
		// 		typedDate := chart.TimeFromFloat64(typed)
		// 		return fmt.Sprintf("%d-%d\n%d", typedDate.Month(), typedDate.Day(), typedDate.Year())
		// 	},
		// },
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dates,
				YValues: reps,
			},
			chart.TimeSeries{
				YAxis:   chart.YAxisSecondary,
				XValues: dates,
				YValues: weights,
			},
		},
	}

	graph.Elements = []chart.Renderable{chart.Legend(&graph)}

	buf := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buf)

	err = os.WriteFile("debug_chart.png", buf.Bytes(), 0644)
	if err != nil {
		log.Printf("Ошибка сохранения debug_chart.png: %v", err)
		return nil, fmt.Errorf("Ошибка сохранения debug_chart.png: %v", err)
	} else {
		log.Println("График сохранён → debug_chart.png (ОТКРОЙ!)")
	}

	
	if err != nil {
		log.Printf("ошибка рендеринга: %w", err)
		return nil, fmt.Errorf("ошибка рендеринга: %w", err)
	}

	log.Printf("PNG создан, размер: %d байт", buf.Len())
	return buf, nil
}