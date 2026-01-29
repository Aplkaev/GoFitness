package history

import (
	"bytes"
	"fmt"
	"log"
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

	// Начальная дата
	startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Заполняем тестовыми данными (7 точек)
	for i := 0; i < 7; i++ {
		currentDate := startDate.AddDate(0, 0, i) // +1 день
		dates = append(dates, currentDate)

		// Пример значений (можно заменить на реальные)
		weight := 100.0 + float64(i*5)  // 100, 105, 110, ...
		rep := 8.0 + float64(i)         // 8, 9, 10, ...

		weights = append(weights, weight)
		reps = append(reps, rep)

		// Выводим каждую итерацию
		log.Printf("Итерация %d: дата = %s, вес = %.1f, повторения = %.1f",
			i, currentDate.Format("02.01.2006"), weight, rep)
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
	
	if err != nil {
		log.Printf("ошибка рендеринга: %w", err)
		return nil, fmt.Errorf("ошибка рендеринга: %w", err)
	}

	log.Printf("PNG создан, размер: %d байт", buf.Len())
	return buf, nil
}