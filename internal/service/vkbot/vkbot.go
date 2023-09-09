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

type Storage interface {
	GetSchedule(institute string, peerId int) (string, error)
	AddUser(institute string, course string, groupNumber string, peerId int) error
	CheckSchedule(institute string, course string, groupNumber string) (bool, error)
	CheckUser(peerId int) (bool, error)
	UserCheckWeek(peerId int) (bool, error)
	UserAddWeek(week string, peerId int) error
	DeleteUser(peerId int) error
}

type VkBot struct {
	vk *api.VK
	lp *longpoll.Longpoll
}

// New создает и запускает бота
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

// TODO change handler
func registerHandlers(bot *VkBot, storage Storage) {
	const op = "service.vkbot.botHandler"

	user := NewUserState()

	bot.lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(obj.Message.PeerID)

		peerId := obj.Message.PeerID
		message := obj.Message.Text
		state, exists := user.GetState(peerId)
		if !exists {
			user.SetState(peerId, StateStart)
		}

		log.Printf("%d: %s; %d", obj.Message.PeerID, obj.Message.Text, state)

		switch message {
		case "Инфо":
			b.Message("Это Чат-Бот с расписанием занятий НИУ МГСУ. \nЧтобы им воспользоваться напиши свои данные согласно инструкции: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
		default:
			switch state {
			case StateStart:
				b.Keyboard(getKeyboard("info"))
				b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
				user.SetState(peerId, StateRegister)

			case StateRegister:
				if message == "Вернуться" {
					b.Keyboard(getKeyboard("info"))
					b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
					break
				}

				message = strings.TrimSpace(message)
				text := strings.Split(message, " ")

				if len(text) != 3 {
					str := "Проверь свои данные на соответствие: " + message
					b.Message(str)
				} else if ok, err := storage.CheckSchedule(text[0], text[1], text[2]); ok && err == nil {
					if err := storage.AddUser(text[0], text[1], text[2], peerId); err != nil {
						log.Printf("%s: %s", op, err)
						str := "Проверь свои данные на соответствие: " + message
						b.Message(str)
					}

					b.Message("Выбери неделю")
					b.Keyboard(getKeyboard("week"))

					user.SetState(peerId, StateWeekSelection)
				} else {
					str := "Проверь свои данные на соответствие: " + message
					b.Message(str)
				}

			case StateWeekSelection:
				if message == "Вернуться" {
					user.SetState(peerId, StateRegister)
					if err := storage.DeleteUser(peerId); err != nil {
						log.Printf("%s: %s", op, err)
						b.Message("Я не понимаю твоего сообщения")
						break
					} else {
						b.Keyboard(getKeyboard("info"))
						b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
						break
					}
				}

				if isValidWeek(message) {
					weekType := strings.Split(message, " ")[0]

					if err := storage.UserAddWeek(weekType, peerId); err != nil {
						log.Printf("%s: %s", op, err)
					}

					b.Message("Выбери день недели")

					if message == "Нечетная неделя" {
						b.Keyboard(getKeyboard("oddweek"))
					} else {
						b.Keyboard(getKeyboard("evenweek"))
					}

					user.SetState(peerId, StateDaySelection)
				} else {
					str := "Проверь свои данные на соответствие: " + message
					b.Message(str)
				}

			case StateDaySelection:
				if message == "Вернуться" {
					user.SetState(peerId, StateRegister)
					if err := storage.DeleteUser(peerId); err != nil {
						log.Printf("%s: %s", op, err)
						b.Message("Я не понимаю твоего сообщения")
						break
					} else {
						b.Keyboard(getKeyboard("info"))
						b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
						break
					}
				}

				if isValidDay(message) {
					if schedule, err := storage.GetSchedule(message, peerId); err != nil {
						log.Printf("%s: %s", op, err)
						b.Message("Я не понимаю твоего сообщения")
					} else {
						b.Message(schedule)
					}
				} else {
					str := "Проверь свои данные на соответствие: " + message
					b.Message(str)
				}
			}
		}

		if _, err := bot.vk.MessagesSend(b.Params); err != nil {
			log.Printf("%s: %s", op, err)
		}
	})
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
