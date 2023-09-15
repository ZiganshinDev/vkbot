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
	AddUser(institute string, course string, groupNumber string, peerId int) error
	UserAddWeek(week string, peerId int) error
	DeleteUser(peerId int) error
	GetSchedule(institute string, peerId int) (string, error)
	CheckSchedule(institute string, course string, groupNumber string) (bool, error)
}

func New() (*VkBot, error) {
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

	bot := &VkBot{vk: vk, lp: lp}

	return bot, nil
}

func (v *VkBot) Start(storage Storage) error {
	const op = "service.vkbot.Start"

	v.registerHandlers(storage)

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

func (v *VkBot) registerHandlers(storage Storage) {
	const op = "service.vkbot.registerHandlers"

	user := NewUserState()

	v.lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		if err := v.handleMessage(user, obj, storage); err != nil {
			log.Printf("%s: %v", op, err)
		}
	})
}

func (v *VkBot) handleMessage(user *User, obj object.MessageNewObject, storage Storage) error {
	const op = "service.op.handleMessage"

	b := params.NewMessagesSendBuilder()
	b.RandomID(0)
	b.PeerID(obj.Message.PeerID)

	peerId := obj.Message.PeerID
	message := obj.Message.Text
	state, exists := user.GetState(peerId)
	if !exists {
		user.SetState(peerId, stateStart)
	}

	log.Printf("%d: %s; %d", peerId, message, state)

	switch message {
	case infoMessage:
		sendInfoMessage(b)
	default:
		switch state {
		case stateStart:
			v.handleStateStart(b, user, peerId, message)
		case stateRegister:
			v.handleStateRegister(b, user, peerId, message, storage)
		case stateWeekSelection:
			v.handleStateWeekSelection(b, user, peerId, message, storage)
		case stateDaySelection:
			v.handleStateDaySelection(b, user, peerId, message, storage)
		}
	}

	_, err := v.vk.MessagesSend(b.Params)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func sendInfoMessage(b *params.MessagesSendBuilder) {
	b.Message("Это Чат-Бот с расписанием занятий НИУ МГСУ. \nЧтобы им воспользоваться напиши свои данные согласно инструкции: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
}

func (v *VkBot) handleStateStart(b *params.MessagesSendBuilder, user *User, peerId int, message string) {
	b.Keyboard(v.getKeyboard("info"))
	b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
	user.SetState(peerId, stateRegister)
}

func (v *VkBot) handleStateRegister(b *params.MessagesSendBuilder, user *User, peerId int, message string, storage Storage) {
	const op = "service.vkbot.handleStateRegister"

	if message == backMessage {
		b.Keyboard(v.getKeyboard("info"))
		b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
		return
	}

	parts, err := validateRegisterMessage(message)
	if err != nil {
		b.Message("Проверь свои данные на соответствие: " + message)
	} else if ok, err := storage.CheckSchedule(parts[0], parts[1], parts[2]); ok && err == nil {
		if err := storage.AddUser(parts[0], parts[1], parts[2], peerId); err != nil {
			log.Printf("%s: %v", op, err)
			str := "Проверь свои данные на соответствие: " + message
			b.Message(str)
		}

		b.Message("Выбери неделю")
		b.Keyboard(v.getKeyboard("week"))

		user.SetState(peerId, stateWeekSelection)
	} else {
		b.Message("Проверь свои данные на соответствие: " + message)
	}
}

func (v *VkBot) handleStateWeekSelection(b *params.MessagesSendBuilder, user *User, peerId int, message string, storage Storage) {
	const op = "service.vkbot.handleStateWeekSelection"

	if message == backMessage {
		user.SetState(peerId, stateRegister)
		if err := storage.DeleteUser(peerId); err != nil {
			log.Printf("%s: %v", op, err)
			b.Message("Я не понимаю твоего сообщения")
			return
		} else {
			b.Keyboard(v.getKeyboard("info"))
			b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
			return
		}
	}

	if validateWeekMessage(message) {
		weekType := strings.Split(message, " ")[0]

		if err := storage.UserAddWeek(weekType, peerId); err != nil {
			log.Printf("%s: %v", op, err)
		}

		b.Message("Выбери день недели")

		if message == "Нечетная неделя" {
			b.Keyboard(v.getKeyboard("oddweek"))
		} else {
			b.Keyboard(v.getKeyboard("evenweek"))
		}

		user.SetState(peerId, stateDaySelection)
	} else {
		b.Message("Проверь свои данные на соответствие: " + message)
	}
}

func (v *VkBot) handleStateDaySelection(b *params.MessagesSendBuilder, user *User, peerId int, message string, storage Storage) {
	const op = "service.vk.handleStateDaySelection"

	if message == backMessage {
		user.SetState(peerId, stateRegister)
		if err := storage.DeleteUser(peerId); err != nil {
			log.Printf("%s: %v", op, err)
			b.Message("Я не понимаю твоего сообщения")
			return
		} else {
			b.Keyboard(v.getKeyboard("info"))
			b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
			return
		}
	}
	if validateDayMessage(message) {
		if schedule, err := storage.GetSchedule(message, peerId); err != nil {
			log.Printf("%s: %v", op, err)
			b.Message("Я не понимаю твоего сообщения")
		} else {
			b.Message(schedule)
		}
	} else {
		b.Message("Проверь свои данные на соответствие: " + message)
	}
}

func validateRegisterMessage(message string) ([]string, error) {
	var result []string

	message = strings.TrimSpace(message)
	result = strings.Split(message, " ")
	if len(result) != 3 {
		return nil, fmt.Errorf("len != 3")
	}

	return result, nil
}

func validateWeekMessage(week string) bool {
	return week == "Нечетная неделя" || week == "Четная неделя"
}

func validateDayMessage(day string) bool {
	days := map[string]bool{
		"Понедельник": true,
		"Вторник":     true,
		"Среда":       true,
		"Четверг":     true,
		"Пятница":     true,
	}

	return days[day]
}
