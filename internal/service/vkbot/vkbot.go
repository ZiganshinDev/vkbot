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
	vk             *api.VK
	lp             *longpoll.Longpoll
	scheduleReader ScheduleReader
	userHandler    UserHandler
}

type Storage interface {
	ScheduleReader
	UserHandler
}

type ScheduleReader interface {
	GetSchedule(institute string, peerId int) (string, error)
	CheckSchedule(institute string, course string, groupNumber string) (bool, error)
}

type UserHandler interface {
	AddUser(institute string, course string, groupNumber string, peerId int) error
	UserAddWeek(week string, peerId int) error
	DeleteUser(peerId int) error
}

func New(storage Storage) (*VkBot, error) {
	const op = "service.vkbot.New"

	vk := api.NewVK(os.Getenv("VK_TOKEN"))

	group, err := vk.GroupsGetByID(nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	lp, err := longpoll.NewLongpoll(vk, group[0].ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	bot := &VkBot{vk: vk, lp: lp, userHandler: storage, scheduleReader: storage}

	return bot, nil
}

func (v *VkBot) Start() error {
	const op = "service.vkbot.Start"

	registerHandlers(v)

	log.Println("Start Long Poll")
	if err := v.lp.Run(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

const (
	stateStart int = iota
	stateRegister
	stateWeekSelection
	stateDaySelection
)

const (
	infoMessage = "Инфо"
	backMessage = "Вернуться"
)

func registerHandlers(bot *VkBot) {
	const op = "service.vkbot.registerHandlers"

	user := NewUserState()

	bot.lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		if err := handleMessage(bot, user, obj); err != nil {
			log.Printf("%s: %v", op, err)
		}
	})
}

func handleMessage(bot *VkBot, user *User, obj object.MessageNewObject) error {
	const op = "service.op.handleMessage"

	b := params.NewMessagesSendBuilder()
	b.RandomID(0)
	b.PeerID(obj.Message.PeerID)

	peerId := obj.Message.PeerID
	text := obj.Message.Text
	state, exists := user.GetState(peerId)
	if !exists {
		user.SetState(peerId, stateStart)
	}

	log.Printf("%d: %s; %d", peerId, text, state)

	switch text {
	case infoMessage:
		sendInfoMessage(b)
	default:
		switch state {
		case stateStart:
			handleStateStart(b, user, peerId, text)
		case stateRegister:
			handleStateRegister(b, user, peerId, text, bot.scheduleReader, bot.userHandler)
		case stateWeekSelection:
			handleStateWeekSelection(b, user, peerId, text, bot.userHandler)
		case stateDaySelection:
			handleStateDaySelection(b, user, peerId, text, bot.userHandler, bot.scheduleReader)
		}
	}

	_, err := bot.vk.MessagesSend(b.Params)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}

func sendInfoMessage(b *params.MessagesSendBuilder) {
	b.Message("Это Чат-Бот с расписанием занятий НИУ МГСУ. \nЧтобы им воспользоваться напиши свои данные согласно инструкции: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
}

func handleStateStart(b *params.MessagesSendBuilder, user *User, peerId int, text string) {
	b.Keyboard(getKeyboard("info"))
	b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
	user.SetState(peerId, stateRegister)
}

func handleStateRegister(b *params.MessagesSendBuilder, user *User, peerId int, text string, scheduleReader ScheduleReader, userHandler UserHandler) {
	const op = "service.vkbot.handleStateRegister"

	if text == backMessage {
		b.Keyboard(getKeyboard("info"))
		b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
		return
	}

	text = strings.TrimSpace(text)
	parts := strings.Split(text, " ")

	if len(parts) != 3 {
		str := "Проверь свои данные на соответствие: " + text
		b.Message(str)
	} else if ok, err := scheduleReader.CheckSchedule(parts[0], parts[1], parts[2]); ok && err == nil {
		if err := userHandler.AddUser(parts[0], parts[1], parts[2], peerId); err != nil {
			log.Printf("%s: %s", op, err)
			str := "Проверь свои данные на соответствие: " + text
			b.Message(str)
		}

		b.Message("Выбери неделю")
		b.Keyboard(getKeyboard("week"))

		user.SetState(peerId, stateWeekSelection)
	} else {
		str := "Проверь свои данные на соответствие: " + text
		b.Message(str)
	}
}

func handleStateWeekSelection(b *params.MessagesSendBuilder, user *User, peerId int, text string, userHandler UserHandler) {
	const op = "service.vkbot.handleStateWeekSelection"

	if text == backMessage {
		user.SetState(peerId, stateRegister)
		if err := userHandler.DeleteUser(peerId); err != nil {
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

		if err := userHandler.UserAddWeek(weekType, peerId); err != nil {
			log.Printf("%s: %s", op, err)
		}

		b.Message("Выбери день недели")

		if text == "Нечетная неделя" {
			b.Keyboard(getKeyboard("oddweek"))
		} else {
			b.Keyboard(getKeyboard("evenweek"))
		}

		user.SetState(peerId, stateDaySelection)
	} else {
		str := "Проверь свои данные на соответствие: " + text
		b.Message(str)
	}
}

func handleStateDaySelection(b *params.MessagesSendBuilder, user *User, peerId int, text string, userHandler UserHandler, scheduleReader ScheduleReader) {
	const op = "service.vk.handleStateDaySelection"

	if text == backMessage {
		user.SetState(peerId, stateRegister)
		if err := userHandler.DeleteUser(peerId); err != nil {
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
		if schedule, err := scheduleReader.GetSchedule(text, peerId); err != nil {
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
