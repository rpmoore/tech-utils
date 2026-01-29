package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type UnitCategory struct {
	Name  string
	Units []Unit
}

type Unit struct {
	Code string
	Name string
}

var UnitCategories = []UnitCategory{
	{
		Name: "length",
		Units: []Unit{
			{Code: "m", Name: "Meters"},
			{Code: "km", Name: "Kilometers"},
			{Code: "cm", Name: "Centimeters"},
			{Code: "mm", Name: "Millimeters"},
			{Code: "mi", Name: "Miles"},
			{Code: "yd", Name: "Yards"},
			{Code: "ft", Name: "Feet"},
			{Code: "in", Name: "Inches"},
		},
	},
	{
		Name: "weight",
		Units: []Unit{
			{Code: "kg", Name: "Kilograms"},
			{Code: "g", Name: "Grams"},
			{Code: "mg", Name: "Milligrams"},
			{Code: "lb", Name: "Pounds"},
			{Code: "oz", Name: "Ounces"},
		},
	},
	{
		Name: "temperature",
		Units: []Unit{
			{Code: "c", Name: "Celsius"},
			{Code: "f", Name: "Fahrenheit"},
			{Code: "k", Name: "Kelvin"},
		},
	},
	{
		Name: "data",
		Units: []Unit{
			{Code: "b", Name: "Bytes"},
			{Code: "kb", Name: "Kilobytes"},
			{Code: "mb", Name: "Megabytes"},
			{Code: "gb", Name: "Gigabytes"},
			{Code: "tb", Name: "Terabytes"},
		},
	},
}

var lengthToBase = map[string]float64{
	"m": 1, "km": 1000, "cm": 0.01, "mm": 0.001,
	"mi": 1609.344, "yd": 0.9144, "ft": 0.3048, "in": 0.0254,
}

var weightToBase = map[string]float64{
	"kg": 1000, "g": 1, "mg": 0.001, "lb": 453.592, "oz": 28.3495,
}

var dataToBase = map[string]float64{
	"b": 1, "kb": 1024, "mb": 1024 * 1024,
	"gb": 1024 * 1024 * 1024, "tb": 1024 * 1024 * 1024 * 1024,
}

type UnitsHandler struct{}

func NewUnitsHandler() *UnitsHandler {
	return &UnitsHandler{}
}

func (h *UnitsHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "units.html", map[string]any{
		"Title":      "Unit Converter",
		"Categories": UnitCategories,
	})
}

func (h *UnitsHandler) Convert(c echo.Context) error {
	valueStr := c.FormValue("value")
	category := c.FormValue("category")
	from := c.FormValue("from")
	to := c.FormValue("to")

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Invalid number: `+html.EscapeString(valueStr)+`</div>`)
	}

	var result float64
	var formula string

	switch category {
	case "length":
		fromFactor, ok1 := lengthToBase[from]
		toFactor, ok2 := lengthToBase[to]
		if !ok1 || !ok2 {
			return c.HTML(http.StatusOK, `<div class="error">Unknown length unit</div>`)
		}
		result = value * fromFactor / toFactor
		formula = fmt.Sprintf("%.6g %s * %.6g / %.6g = %.6g %s", value, from, fromFactor, toFactor, result, to)

	case "weight":
		fromFactor, ok1 := weightToBase[from]
		toFactor, ok2 := weightToBase[to]
		if !ok1 || !ok2 {
			return c.HTML(http.StatusOK, `<div class="error">Unknown weight unit</div>`)
		}
		result = value * fromFactor / toFactor
		formula = fmt.Sprintf("%.6g %s * %.6g / %.6g = %.6g %s", value, from, fromFactor, toFactor, result, to)

	case "temperature":
		result, formula = convertTemperature(value, from, to)
		if formula == "" {
			return c.HTML(http.StatusOK, `<div class="error">Unknown temperature unit</div>`)
		}

	case "data":
		fromFactor, ok1 := dataToBase[from]
		toFactor, ok2 := dataToBase[to]
		if !ok1 || !ok2 {
			return c.HTML(http.StatusOK, `<div class="error">Unknown data unit</div>`)
		}
		result = value * fromFactor / toFactor
		formula = fmt.Sprintf("%.6g %s * %.0f / %.0f = %.6g %s", value, from, fromFactor, toFactor, result, to)

	default:
		return c.HTML(http.StatusOK, `<div class="error">Unknown category</div>`)
	}

	output := fmt.Sprintf(`
<article>
    <header>Result</header>
    <p><strong>%.6g %s = %.6g %s</strong></p>
    <p class="formula">Formula: %s</p>
</article>`, value, from, result, to, formula)

	return c.HTML(http.StatusOK, output)
}

func convertTemperature(value float64, from, to string) (float64, string) {
	var celsius float64
	switch from {
	case "c":
		celsius = value
	case "f":
		celsius = (value - 32) * 5 / 9
	case "k":
		celsius = value - 273.15
	default:
		return 0, ""
	}

	var result float64
	var formula string
	switch to {
	case "c":
		result = celsius
		if from == "f" {
			formula = fmt.Sprintf("(%.6g - 32) * 5/9 = %.6g", value, result)
		} else if from == "k" {
			formula = fmt.Sprintf("%.6g - 273.15 = %.6g", value, result)
		} else {
			formula = fmt.Sprintf("%.6g C = %.6g C", value, result)
		}
	case "f":
		result = celsius*9/5 + 32
		if from == "c" {
			formula = fmt.Sprintf("%.6g * 9/5 + 32 = %.6g", value, result)
		} else if from == "k" {
			formula = fmt.Sprintf("(%.6g - 273.15) * 9/5 + 32 = %.6g", value, result)
		} else {
			formula = fmt.Sprintf("%.6g F = %.6g F", value, result)
		}
	case "k":
		result = celsius + 273.15
		if from == "c" {
			formula = fmt.Sprintf("%.6g + 273.15 = %.6g", value, result)
		} else if from == "f" {
			formula = fmt.Sprintf("(%.6g - 32) * 5/9 + 273.15 = %.6g", value, result)
		} else {
			formula = fmt.Sprintf("%.6g K = %.6g K", value, result)
		}
	default:
		return 0, ""
	}

	return result, formula
}
