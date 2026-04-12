package prago

import "math"

// niceNum je pomocná funkce, která najde "pěkné" číslo (1, 2, 5 nebo 10 * 10^n)
// nejblíže k zadanému číslu x.
func niceNum(x float64, round bool) float64 {
	exp := math.Floor(math.Log10(x))
	f := x / math.Pow(10, exp)
	var nf float64

	if round {
		if f < 1.5 {
			nf = 1
		} else if f < 3 {
			nf = 2
		} else if f < 7 {
			nf = 5
		} else {
			nf = 10
		}
	} else {
		if f <= 1 {
			nf = 1
		} else if f <= 2 {
			nf = 2
		} else if f <= 5 {
			nf = 5
		} else {
			nf = 10
		}
	}
	return nf * math.Pow(10, exp)
}

// CalculateGraphTicks přijme minimum, maximum a přibližný počet čar, které v grafu chceš.
// Vrací řezy (slice) float64 hodnot, které bys měl vykreslit jako pomocné čáry.
func CalculateGraphTicks(min, max float64, maxTicks int) []float64 {
	// Ochrana před nesmyslným vstupem
	if max <= min {
		return []float64{min, max}
	}
	if maxTicks < 2 {
		maxTicks = 2
	}

	// 1. Zjistíme "pěkný" rozsah
	rangeNum := niceNum(max-min, false)

	// 2. Vypočítáme ideální krok mezi čarami
	tickSpacing := niceNum(rangeNum/float64(maxTicks-1), true)

	// 3. Spočítáme nové (zaokrouhlené) minimum a maximum grafu
	niceMin := math.Floor(min/tickSpacing) * tickSpacing
	niceMax := math.Ceil(max/tickSpacing) * tickSpacing

	// 4. Vygenerujeme samotné hodnoty
	var ticks []float64
	// Přidáváme 0.5 * tickSpacing kvůli nepřesnostem float64 aritmetiky,
	// abychom nepřišli o poslední hodnotu
	for val := niceMin; val <= niceMax+(0.5*tickSpacing); val += tickSpacing {
		// Zbavíme se případných float zaokrouhlovacích chyb blízko nuly
		if math.Abs(val) < 1e-10 {
			val = 0
		}
		if val > max || val < min {
			continue
		}
		ticks = append(ticks, val)
	}

	return ticks
}
