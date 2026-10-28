package main

import "fmt"

type Point struct {
	x float64
	y float64
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

func main() {
	var pointCount uint8
	fmt.Print("Введите количество точек: ")
	fmt.Scanln(&pointCount)

	var points []Point = make([]Point, pointCount)

	var input_x, input_y float64
	for i := 0; i < int(pointCount); i++ {
		fmt.Printf("Введите координаты точки №%d через запятую: ", i+1)
		_, err := fmt.Scanf("%f, %f\n", &input_x, &input_y)

		if err != nil {
			fmt.Println("Ошибка ввода, попробуйте еще раз. Пример: 1.2, 3.4")
			i--
			continue
		} else {
			points[i] = Point{
				x: input_x,
				y: input_y,
			}
		}
	}

	function := Function{
		addends: make([]Addend, pointCount),
	}
	var new_addend Addend
	for i := 0; i < int(pointCount); i++ {
		new_addend = Addend{
			a:       1,
			factors: make([]Factor, pointCount-1),
		}
		for j := 0; j < i; j++ {
			new_addend.factors[j] = Factor{
				x0: points[j].x,
			}
		}
		for j := i; j < int(pointCount)-1; j++ {
			new_addend.factors[j] = Factor{
				x0: points[j+1].x,
			}

		}
		new_addend.a *= points[i].y / new_addend.forward(points[i].x)
		function.addends[i] = new_addend
	}

	fmt.Println(function.forward(10))
	fmt.Println(function.forward(1))
	fmt.Println(function.forward(2))
	fmt.Println(function.forward(3))
}
