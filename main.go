package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Factor struct {
	x0 float64
}

type Addend struct {
	a       float64
	factors []Factor
}

func (addend Addend) forward(x float64) float64 {
	var result float64 = addend.a
	for i := 0; i < len(addend.factors); i++ {
		result *= x - addend.factors[i].x0
	}
	return result
}

type Function struct {
	addends []Addend
}

func (function Function) forward(x float64) float64 {
	var result float64
	for i := 0; i < len(function.addends); i++ {
		result += (function.addends)[i].forward(x)
	}
	return result
}

func (function Function) toPolynomial() []float64 {
	n := len(function.addends[0].factors)
	totalCoeffs := make([]float64, n+1)

	for _, addend := range function.addends {
		currentCoeffs := make([]float64, n+1)
		currentCoeffs[0] = 1

		for idx, factor := range addend.factors {
			for j := idx + 1; j > 0; j-- {
				currentCoeffs[j] = currentCoeffs[j-1] - currentCoeffs[j]*factor.x0
			}
			currentCoeffs[0] = -currentCoeffs[0] * factor.x0
		}

		for j := 0; j <= n; j++ {
			totalCoeffs[j] += currentCoeffs[j] * addend.a
		}
	}

	return totalCoeffs
}

func runConsoleMode() {
	var pointCount uint8
	fmt.Print("Введите количество точек: ")
	fmt.Scanln(&pointCount)

	var points []Point = make([]Point, pointCount)

	var input_x, input_y float64
	fmt.Println("Вводите координаты точки через запятую!")
	for i := 0; i < int(pointCount); i++ {
		fmt.Printf("Точка №%d: ", i+1)
		_, err := fmt.Scanf("%f, %f\n", &input_x, &input_y)

		if err != nil {
			fmt.Println("Ошибка ввода, попробуйте еще раз. Пример: 1.2, 3.4")
			i--
			continue
		} else {
			points[i] = Point{
				X: input_x,
				Y: input_y,
			}
		}
	}

	function := buildLagrangeFunction(points)
	coeffs := function.toPolynomial()
	fmt.Print("Итоговый многочлен: ")
	printPolynomial(coeffs)
}

func buildLagrangeFunction(points []Point) Function {
	pointCount := len(points)
	function := Function{
		addends: make([]Addend, pointCount),
	}
	var new_addend Addend
	for i := 0; i < pointCount; i++ {
		new_addend = Addend{
			a:       1,
			factors: make([]Factor, pointCount-1),
		}
		for j := 0; j < i; j++ {
			new_addend.factors[j] = Factor{x0: points[j].X}
		}
		for j := i; j < pointCount-1; j++ {
			new_addend.factors[j] = Factor{x0: points[j+1].X}
		}
		denom := new_addend.forward(points[i].X)
		if denom != 0 {
			new_addend.a *= points[i].Y / denom
		} else {
			new_addend.a = 0
		}
		function.addends[i] = new_addend
	}
	return function
}

func printPolynomial(coeffs []float64) {
	first := true
	for i := len(coeffs) - 1; i >= 0; i-- {
		if coeffs[i] == 0 {
			continue
		}
		if !first && coeffs[i] > 0 {
			fmt.Print(" + ")
		} else if coeffs[i] < 0 {
			fmt.Print(" - ")
		}

		val := coeffs[i]
		if val < 0 {
			val = -val
		}

		if i == 0 {
			fmt.Printf("%.2f", val)
		} else if i == 1 {
			fmt.Printf("%.2fx", val)
		} else {
			fmt.Printf("%.2fx^%d", val, i)
		}
		first = false
	}
	fmt.Println()
}

type Request struct {
	Points []Point `json:"points"`
}

type Response struct {
	Points     []Point `json:"points"`
	Polynomial string  `json:"polynomial"`
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.Points) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	points := req.Points
	function := buildLagrangeFunction(points)

	minX, maxX := points[0].X, points[0].X
	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
	}

	if minX == maxX {
		minX -= 5
		maxX += 5
	} else {
		padding := (maxX - minX) * 0.2
		minX -= padding
		maxX += padding
	}

	chartPoints := make([]Point, 0, 200)
	steps := 200
	stepSize := (maxX - minX) / float64(steps)
	for i := 0; i <= steps; i++ {
		curX := minX + float64(i)*stepSize
		chartPoints = append(chartPoints, Point{
			X: curX,
			Y: function.forward(curX),
		})
	}

	polyStr := "None"

	resp := Response{
		Points:     chartPoints,
		Polynomial: polyStr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func runWebMode(port string) {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/calculate", calculateHandler)

	fmt.Printf("Веб-сервер запущен на http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Ошибка старта сервера:", err)
	}
}

func main() {
	webMode := flag.Bool("web", false, "Запустить в режиме веб-сервиса")
	port := flag.String("port", "8080", "Порт для веб-сервера")
	flag.Parse()

	if *webMode {
		runWebMode(*port)
	} else {
		runConsoleMode()
	}
}
