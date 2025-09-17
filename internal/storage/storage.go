package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"time"
)

const dataDir = "data"

type Session struct {
	Date       time.Time `json:"date"`
	Sets       []int     `json:"sets"`
	Planned    []int     `json:"planned"`
	MaxSet     int       `json:"max_set"`
	TotalReps  int       `json:"total_reps"`
	IsFinished bool      `json:"is_finished"`
}

type UserData struct {
	UserId         int64     `json:"user_id"`
	Sessions       []Session `json:"sessions"`
	CurrentWorkout []int     `json:"current_workout"`
	LastPlan       []int     `json:"last_plan"`
}

func NewUser(userId int64) *UserData {
	return &UserData{UserId: userId}
}

func ensureDataDir() error {
	return os.MkdirAll(dataDir, 0755)
}

func userFile(UserId int64) string {
	return filepath.Join(dataDir, fmt.Sprintf("%d.json", UserId))
}

func LoadUser(UserId int64) (*UserData, error) {
	_ = ensureDataDir()
	f := userFile(UserId)
	if _, err := os.Stat(f); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewUser(UserId), nil
		}
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var u UserData
	if err := json.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func SaveUser(u *UserData) error {
	_ = ensureDataDir()
	b, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(userFile(u.UserId), b, 0644)
}

func FinishWorkout(u *UserData) (Session, error) {
	if len(u.CurrentWorkout) == 0 {
		return Session{}, errors.New("текущая тренировка пуста")
	}
	maxSet := 0
	total := 0
	for _, r := range u.CurrentWorkout {
		if r > maxSet {
			maxSet = r
		}
		total += r
	}
	s := Session{
		Date:       time.Now(),
		Sets:       append([]int(nil), u.CurrentWorkout...),
		MaxSet:     maxSet,
		TotalReps:  total,
		IsFinished: true,
		Planned:    append([]int(nil), u.LastPlan...),
	}
	u.Sessions = append(u.Sessions, s)
	u.CurrentWorkout = nil
	return s, SaveUser(u)
}

func FormatStats(u *UserData) string {
	if len(u.Sessions) == 0 {
		return "Пока нет завершённых тренировок."
	}
	cut := time.Now().AddDate(0, 0, -7)
	total := 0
	bestSet := 0
	days := map[string]int{}
	for _, s := range u.Sessions {
		if !s.IsFinished {
			continue
		}
		if s.Date.Before(cut) {
			continue
		}
		total += s.TotalReps
		if s.MaxSet > bestSet {
			bestSet = s.MaxSet
		}
		key := s.Date.Format("2006-01-02")
		days[key] += s.TotalReps
	}
	out := "Статистика за последние 7 дней:\n"
	out += fmt.Sprintf("- Общий объём: %d повторений\n", total)
	out += fmt.Sprintf("- Лучший подход: %d\n", bestSet)
	if len(days) > 0 {
		out += "- По дням:\n"
		for d, v := range days {
			out += fmt.Sprintf("  %s: %d\n", d, v)
		}
	}
	return out
}

// Вспомогательная функция может пригодиться позже
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// На будущее — чтобы не тянуть математику в planner.
func RoundTo5(x float64) int {
	return int(math.Round(x/5.0) * 5.0)
}
