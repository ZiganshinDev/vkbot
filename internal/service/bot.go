package service

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/SevereCloud/vksdk/api"
	"github.com/SevereCloud/vksdk/api/params"
	"github.com/SevereCloud/vksdk/object"
	"github.com/ZiganshinDev/scheduleVKBot/internal/database"

	longpoll "github.com/SevereCloud/vksdk/longpoll-bot"
)

// CreateBot создает и запускает бота
func CreateBot() {
	vk := api.NewVK(os.Getenv("VK_TOKEN"))

	group, err := vk.GroupsGetByID(nil)
	if err != nil {
		log.Fatal(err)
	}

	lp, err := longpoll.NewLongpoll(vk, group[0].ID)
	if err != nil {
		log.Fatal(err)
	}

	botHandler(vk, lp)

	log.Println("Start Long Poll")
	if err := lp.Run(); err != nil {
		log.Fatal(err)
	}
}

// botHandler обрабатывает сообщения бота
func botHandler(vk *api.VK, lp *longpoll.Longpoll) {
	lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(obj.Message.PeerID)

		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)

		userPeerID := strconv.Itoa(obj.Message.PeerID)
		userMsg := obj.Message.Text

		// Обработка команд начала и возвращения
		if userMsg == "Начать" || userMsg == "Вернуться" {
			handleStartMessage(userPeerID, b)
		} else {
			handleUserMessage(userPeerID, userMsg, b)
		}

		vk.MessagesSend(b.Params)
	})
}

// handleStartMessage обрабатывает команду начала и возвращения
func handleStartMessage(userPeerID string, b *params.MessagesSendBuilder) {
	if database.CheckUser(userPeerID) {
		database.DeleteUser(userPeerID)
	}
	b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
}

// handleUserMessage обрабатывает сообщения пользователя
func handleUserMessage(userPeerID, userMsg string, b *params.MessagesSendBuilder) {
	if !database.CheckUser(userPeerID) {
		handleFirstMessage(userPeerID, userMsg, b)
	} else if !database.CheckUserWithWeekType(userPeerID) {
		handleWeekTypeMessage(userPeerID, userMsg, b)
	} else if isDayOfWeek(userMsg) {
		handleDayOfWeekMessage(userPeerID, userMsg, b)
	} else {
		b.Message("Я не понимаю твоего сообщения")
	}
}

// handleFirstMessage обрабатывает первое сообщение пользователя
func handleFirstMessage(userPeerID, userMsg string, b *params.MessagesSendBuilder) {
	userMsg = strings.TrimSpace(userMsg)
	text := strings.Split(userMsg, " ")
	if len(text) != 3 || !database.CheckSchedule(text[0], text[1], text[2]) {
		b.Message("Я не понимаю твоего сообщения")
	} else {
		database.AddUser(text[0], text[1], text[2], userPeerID)
		b.Message("Выбери неделю")
		b.Keyboard(getKeyboard("week"))
	}
}

// handleWeekTypeMessage обрабатывает сообщение о выборе типа недели
func handleWeekTypeMessage(userPeerID, userMsg string, b *params.MessagesSendBuilder) {
	if userMsg == "Нечетная неделя" || userMsg == "Четная неделя" {
		weekType := strings.Split(userMsg, " ")[0]
		database.AddWeekToUser(weekType, userPeerID)
		b.Message("Выбери день недели")
		if userMsg == "Нечетная неделя" {
			b.Keyboard(getKeyboard("oddweek"))
		} else {
			b.Keyboard(getKeyboard("evenweek"))
		}
	} else {
		b.Message("Я не понимаю твоего сообщения")
	}
}

// handleDayOfWeekMessage обрабатывает сообщение о выборе дня недели
func handleDayOfWeekMessage(userPeerID, userMsg string, b *params.MessagesSendBuilder) {
	b.Message(database.DBShowSchedule(userMsg, userPeerID))
}

// isDayOfWeek проверяет, является ли строка днем недели
func isDayOfWeek(day string) bool {
	days := map[string]bool{
		"Понедельник": true,
		"Вторник":     true,
		"Среда":       true,
		"Четверг":     true,
		"Пятница":     true,
	}

	return days[day]
}
