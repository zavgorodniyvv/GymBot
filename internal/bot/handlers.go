package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/zavgorodniyvv/GymBot/internal/planner"
	"github.com/zavgorodniyvv/GymBot/internal/storage"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	storage *storage.MongoStorage
	mu      sync.Mutex
	timers  map[int64]*restTimer
}

func New(api *tgbotapi.BotAPI, st *storage.MongoStorage) *Bot {
	return &Bot{api: api, storage: st, timers: make(map[int64]*restTimer)}
}

func (b *Bot) Handle(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	text := strings.TrimSpace(update.Message.Text)

	// Загружаем/создаём данные пользователя
	u, _ := b.storage.LoadUser(userID)

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
		_ = b.storage.SaveUser(u)
		b.api.Send(tgbotapi.NewMessage(chatID, "Данные сброшены."))
		return

	case strings.HasPrefix(text, "/plan"):
		if len(u.LastPlan) == 0 {
			u.LastPlan = planner.MakePlan(u)
			_ = b.storage.SaveUser(u)
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
		s, err := b.storage.FinishWorkout(u)
		if err != nil {
			b.api.Send(tgbotapi.NewMessage(chatID, "Ошибка: "+err.Error()))
			return
		}
		// Генерируем новый план для следующей тренировки
		u.LastPlan = planner.MakePlan(u)
		_ = b.storage.SaveUser(u)

		resp := fmt.Sprintf("Тренировка сохранена!\nПодходы: %v\nЛучший подход: %d\nОбъём: %d\n\nПлан на следующую тренировку:\n%s",
			s.Sets, s.MaxSet, s.TotalReps, planner.FormatPlan(u.LastPlan))
		b.api.Send(tgbotapi.NewMessage(chatID, resp))
		return
	}

	// Пытаемся распарсить число повторений
	if n, err := strconv.Atoi(text); err == nil && n > 0 {
		const restSeconds = 120
		u.CurrentWorkout = append(u.CurrentWorkout, n)
		_ = b.storage.SaveUser(u)
		b.api.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"Ок, записал подход: %d. Текущая тренировка: %v\nКогда закончишь — пришли /end",
			n, u.CurrentWorkout)))

		b.cancelTimer(chatID)

		timerMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Таймер отдыха: %d секунд", restSeconds))
		sentTimerMsg, err := b.api.Send(timerMsg)
		if err == nil {
			b.launchRestTimer(chatID, sentTimerMsg.MessageID, restSeconds)
		}
		return
	}

	// Непонятный ввод
	b.api.Send(tgbotapi.NewMessage(chatID, "Я понимаю команды (/plan, /stats, /end, /reset) и числа (повторы в подходе)."))
}

type restTimer struct {
	cancel context.CancelFunc
}

func (b *Bot) launchRestTimer(chatID int64, messageID int, totalSeconds int) {
	ctx, cancel := context.WithCancel(context.Background())
	timer := &restTimer{cancel: cancel}

	b.mu.Lock()
	if prev, ok := b.timers[chatID]; ok {
		prev.cancel()
	}
	b.timers[chatID] = timer
	b.mu.Unlock()

	go b.runRestTimer(ctx, chatID, messageID, totalSeconds, timer)
}

func (b *Bot) runRestTimer(ctx context.Context, chatID int64, messageID int, totalSeconds int, timer *restTimer) {
	defer b.clearTimer(chatID, timer)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for remaining := totalSeconds - 1; remaining >= 0; remaining-- {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		var text string
		if remaining == 0 {
			text = "Отдых завершён! Можно переходить к следующему подходу."
		} else {
			text = fmt.Sprintf("Таймер отдыха: %d секунд", remaining)
		}

		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		if _, err := b.api.Send(edit); err != nil {
			return
		}
	}
}

func (b *Bot) clearTimer(chatID int64, timer *restTimer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if current, ok := b.timers[chatID]; ok && current == timer {
		delete(b.timers, chatID)
	}
}

func (b *Bot) cancelTimer(chatID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if timer, ok := b.timers[chatID]; ok {
		timer.cancel()
		delete(b.timers, chatID)
	}
}
