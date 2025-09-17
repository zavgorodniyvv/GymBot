package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/zavgorodniyvv/GymBot/internal/planner"
	"github.com/zavgorodniyvv/GymBot/internal/storage"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

func New(api *tgbotapi.BotAPI) *Bot {
	return &Bot{api: api}
}

func (b *Bot) Handle(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	text := strings.TrimSpace(update.Message.Text)

	// Загружаем/создаём данные пользователя
	u, _ := storage.LoadUser(userID)

	switch {
	case strings.HasPrefix(text, "/start"):
		msg := "Привет! Я бот для учёта отжиманий.\n" +
			"- Пришли число — я добавлю подход в текущую тренировку.\n" +
			"- /plan — предложу план на следующую тренировку.\n" +
			"- /end — завершить тренировку и сохранить результат.\n" +
			"- /stats — статистика за 7 дней.\n" +
			"- /reset — сброс данных.\n\n" +
			"Начни с контрольного подхода и просто пришли число, например: 20"
		b.api.Send(tgbotapi.NewMessage(chatID, msg))
		return

	case strings.HasPrefix(text, "/reset"):
		u = storage.NewUser(userID)
		_ = storage.SaveUser(u)
		b.api.Send(tgbotapi.NewMessage(chatID, "Данные сброшены."))
		return

	case strings.HasPrefix(text, "/plan"):
		if len(u.LastPlan) == 0 {
			u.LastPlan = planner.MakePlan(u)
			_ = storage.SaveUser(u)
		}
		b.api.Send(tgbotapi.NewMessage(chatID, planner.FormatPlan(u.LastPlan)))
		return

	case strings.HasPrefix(text, "/stats"):
		b.api.Send(tgbotapi.NewMessage(chatID, storage.FormatStats(u)))
		return

	case strings.HasPrefix(text, "/end"):
		if len(u.CurrentWorkout) == 0 {
			b.api.Send(tgbotapi.NewMessage(chatID, "Текущая тренировка пуста. Пришли хотя бы одно число."))
			return
		}
		s, err := storage.FinishWorkout(u)
		if err != nil {
			b.api.Send(tgbotapi.NewMessage(chatID, "Ошибка: "+err.Error()))
			return
		}
		resp := fmt.Sprintf("Тренировка сохранена!\nПодходы: %v\nЛучший подход: %d\nОбъём: %d\n\n%s",
			s.Sets, s.MaxSet, s.TotalReps, planner.FormatPlan(u.LastPlan))
		b.api.Send(tgbotapi.NewMessage(chatID, resp))
		return
	}

	// Пытаемся распарсить число повторений
	if n, err := strconv.Atoi(text); err == nil && n > 0 {
		u.CurrentWorkout = append(u.CurrentWorkout, n)
		_ = storage.SaveUser(u)
		b.api.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"Ок, записал подход: %d. Текущая тренировка: %v\nКогда закончишь — пришли /end",
			n, u.CurrentWorkout)))
		return
	}

	// Непонятный ввод
	b.api.Send(tgbotapi.NewMessage(chatID, "Я понимаю команды (/plan, /stats, /end, /reset) и числа (повторы в подходе)."))
}
