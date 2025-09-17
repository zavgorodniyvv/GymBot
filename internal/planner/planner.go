package planner

import (
	"fmt"
	"math"
	"strings"

	"github.com/zavgorodniyvv/GymBot/internal/storage"
)

func MakePlan(u *storage.UserData) []int {
	M := 0
	if len(u.Sessions) > 0 {
		for i := len(u.Sessions) - 1; i >= 0; i-- {
			if u.Sessions[i].IsFinished {
				M = u.Sessions[i].MaxSet
				break
			}
		}
	}
	if M == 0 {
		// стартовый учебный план
		return []int{5, 7, 9, 7, 5}
	}

	base := storage.RoundTo5(0.6 * float64(M))
	if base < 1 {
		base = 1
	}
	maxAllowed := int(math.Ceil(0.9 * float64(M)))

	raw := []int{base, base + 2, base + 4, base + 2, base}
	for i := range raw {
		raw[i] = clamp(raw[i], 1, maxAllowed)
	}

	// Адаптация к прошлому плану
	var last *storage.Session
	for i := len(u.Sessions) - 1; i >= 0; i-- {
		if u.Sessions[i].IsFinished {
			last = &u.Sessions[i]
			break
		}
	}
	if last != nil && len(last.Planned) > 0 {
		failures := 0
		for i := 0; i < len(last.Planned) && i < len(last.Sets); i++ {
			if last.Sets[i] < last.Planned[i] {
				failures++
			}
		}
		if failures <= 1 {
			for i := range raw {
				raw[i] = clamp(raw[i]+1, 1, maxAllowed)
			}
		} else if failures >= 2 {
			for i := range raw {
				raw[i] = clamp(raw[i]-1, 1, maxAllowed)
			}
		}
	}
	return raw
}

func FormatPlan(plan []int) string {
	if len(plan) == 0 {
		return "Плана пока нет."
	}
	var b strings.Builder
	b.WriteString("План на следующую тренировку:\n")
	for i, v := range plan {
		b.WriteString(fmt.Sprintf("  Подход %d: %d\n", i+1, v))
	}
	b.WriteString("\nСовет: отдых 60–120 сек между подходами. Если легко — добавь +1 в каждом подходе.")
	return b.String()
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
