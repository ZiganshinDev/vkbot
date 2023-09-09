package vkbot

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SevereCloud/vksdk/api"
	"github.com/SevereCloud/vksdk/api/params"
	"github.com/SevereCloud/vksdk/object"

	longpoll "github.com/SevereCloud/vksdk/longpoll-bot"
)

type VkBot struct {
	vk *api.VK
	lp *longpoll.Longpoll
}

type Storage interface {
	GetSchedule(institute string, peerId int) (string, error)
	AddUser(institute string, course string, groupNumber string, peerId int) error
	CheckSchedule(institute string, course string, groupNumber string) (bool, error)
	UserAddWeek(week string, peerId int) error
	DeleteUser(peerId int) error
}

func New(storage Storage) error {
	const op = "service.vkbot.New"

	vk := api.NewVK(os.Getenv("VK_TOKEN"))

	group, err := vk.GroupsGetByID(nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	lp, err := longpoll.NewLongpoll(vk, group[0].ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	bot := &VkBot{vk: vk, lp: lp}

	registerHandlers(bot, storage)

	log.Println("Start Long Poll")
	if err := lp.Run(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

const (
	StateStart int = iota
	StateRegister
	StateWeekSelection
	StateDaySelection
)

func registerHandlers(bot *VkBot, storage Storage) {
	user := NewUserState()

	bot.lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		handleMessage(bot, storage, user, obj)
	})
}

func handleMessage(bot *VkBot, storage Storage, user *User, obj object.MessageNewObject) {
	const op = "service.op.handleMessage"

	b := params.NewMessagesSendBuilder()
	b.RandomID(0)
	b.PeerID(obj.Message.PeerID)

	peerId := obj.Message.PeerID
	text := obj.Message.Text
	state, exists := user.GetState(peerId)
	if !exists {
		user.SetState(peerId, StateStart)
	}

	log.Printf("%d: %s; %d", peerId, text, state)

	switch text {
	case "Инфо":
		sendInfoMessage(b)
	default:
		switch state {
		case StateStart:
			handleStateStart(b, user, peerId, text)
		case StateRegister:
			handleStateRegister(b, user, peerId, text, storage)
		case StateWeekSelection:
			handleStateWeekSelection(b, user, peerId, text, storage)
		case StateDaySelection:
			handleStateDaySelection(b, user, peerId, text, storage)
		}
	}

	if _, err := bot.vk.MessagesSend(b.Params); err != nil {
		log.Printf("%s: %s", op, err)
	}
}

func sendInfoMessage(b *params.MessagesSendBuilder) {
	b.Message("Это Чат-Бот с расписанием занятий НИУ МГСУ. \nЧтобы им воспользоваться напиши свои данные согласно инструкции: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
}

func handleStateStart(b *params.MessagesSendBuilder, user *User, peerId int, text string) {
	b.Keyboard(getKeyboard("info"))
	b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
	user.SetState(peerId, StateRegister)
}

func handleStateRegister(b *params.MessagesSendBuilder, user *User, peerId int, text string, storage Storage) {
	const op = "service.vkbot.handleStateRegister"

	if text == "Вернуться" {
		b.Keyboard(getKeyboard("info"))
		b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
		return
	}

	text = strings.TrimSpace(text)
	parts := strings.Split(text, " ")

	if len(parts) != 3 {
		str := "Проверь свои данные на соответствие: " + text
		b.Message(str)
	} else if ok, err := storage.CheckSchedule(parts[0], parts[1], parts[2]); ok && err == nil {
		if err := storage.AddUser(parts[0], parts[1], parts[2], peerId); err != nil {
			log.Printf("%s: %s", op, err)
			str := "Проверь свои данные на соответствие: " + text
			b.Message(str)
		}

		b.Message("Выбери неделю")
		b.Keyboard(getKeyboard("week"))

		user.SetState(peerId, StateWeekSelection)
	} else {
		str := "Проверь свои данные на соответствие: " + text
		b.Message(str)
	}
}

func handleStateWeekSelection(b *params.MessagesSendBuilder, user *User, peerId int, text string, storage Storage) {
	const op = "service.vkbot.handleStateWeekSelection"

	if text == "Вернуться" {
		user.SetState(peerId, StateRegister)
		if err := storage.DeleteUser(peerId); err != nil {
			log.Printf("%s: %s", op, err)
			b.Message("Я не понимаю твоего сообщения")
			return
		} else {
			b.Keyboard(getKeyboard("info"))
			b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
			return
		}
	}

	if isValidWeek(text) {
		weekType := strings.Split(text, " ")[0]

		if err := storage.UserAddWeek(weekType, peerId); err != nil {
			log.Printf("%s: %s", op, err)
		}

		b.Message("Выбери день недели")

		if text == "Нечетная неделя" {
			b.Keyboard(getKeyboard("oddweek"))
		} else {
			b.Keyboard(getKeyboard("evenweek"))
		}

		user.SetState(peerId, StateDaySelection)
	} else {
		str := "Проверь свои данные на соответствие: " + text
		b.Message(str)
	}
}

func handleStateDaySelection(b *params.MessagesSendBuilder, user *User, peerId int, text string, storage Storage) {
	const op = "service.vk.handleStateDaySelection"

	if text == "Вернуться" {
		user.SetState(peerId, StateRegister)
		if err := storage.DeleteUser(peerId); err != nil {
			log.Printf("%s: %s", op, err)
			b.Message("Я не понимаю твоего сообщения")
			return
		} else {
			b.Keyboard(getKeyboard("info"))
			b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
			return
		}
	}
	if isValidDay(text) {
		if schedule, err := storage.GetSchedule(text, peerId); err != nil {
			log.Printf("%s: %s", op, err)
			b.Message("Я не понимаю твоего сообщения")
		} else {
			b.Message(schedule)
		}
	} else {
		str := "Проверь свои данные на соответствие: " + text
		b.Message(str)
	}
}

func isValidWeek(week string) bool {
	return week == "Нечетная неделя" || week == "Четная неделя"
}

func isValidDay(day string) bool {
	days := map[string]bool{
		"Понедельник": true,
		"Вторник":     true,
		"Среда":       true,
		"Четверг":     true,
		"Пятница":     true,
	}

	return days[day]
}
