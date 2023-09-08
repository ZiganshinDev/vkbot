package vkbot

import (
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
	UserAddWeek(week string, peerId int)
	DeleteUser(peerId int)
}

type VkBot struct {
	vk *api.VK
	lp *longpoll.Longpoll
}

type User struct {
	PeerId  int
	Message string
}

// New создает и запускает бота
func New(storage Storage) {
	vk := api.NewVK(os.Getenv("VK_TOKEN"))

	group, err := vk.GroupsGetByID(nil)
	if err != nil {
		log.Fatal(err)
	}

	lp, err := longpoll.NewLongpoll(vk, group[0].ID)
	if err != nil {
		log.Fatal(err)
	}

	bot := &VkBot{vk: vk, lp: lp}

	botHandler(bot, storage)

	log.Println("Start Long Poll")
	if err := lp.Run(); err != nil {
		log.Fatal(err)
	}
}

// botHandler обрабатывает сообщения
func botHandler(bot *VkBot, storage Storage) {
	const op = "service.vkbot.botHandler"

	bot.lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(obj.Message.PeerID)

		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)

		user := &User{PeerId: obj.Message.PeerID, Message: obj.Message.Text}

		// Обработка команд начала и возвращения
		if user.Message == "Начать" || user.Message == "Вернуться" {
			storage.DeleteUser(user.PeerId)

			b.Keyboard(getKeyboard("info"))
			b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
		} else {
			if user.Message == "Инфо" {
				b.Message("Это Чат-Бот с расписанием занятий НИУ МГСУ. \nЧтобы им воспользоваться напиши свои данные согласно инструкции: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")

			} else if ok, err := storage.CheckUser(user.PeerId); ok && err == nil {
				user.Message = strings.TrimSpace(user.Message)
				text := strings.Split(user.Message, " ")
				if len(text) != 3 {
					b.Message("Я не понимаю твоего сообщения")
				} else if ok, err := storage.CheckSchedule(text[0], text[1], text[2]); ok && err == nil {
					if err := storage.AddUser(text[0], text[1], text[2], user.PeerId); err != nil {
						log.Printf("%s: %s", op, err)
					}

					b.Message("Выбери неделю")
					b.Keyboard(getKeyboard("week"))
				} else {
					b.Message("Я не понимаю твоего сообщения")
				}
			} else if ok, err := storage.UserCheckWeek(user.PeerId); ok && err == nil {
				if user.Message == "Нечетная неделя" || user.Message == "Четная неделя" {
					weekType := strings.Split(user.Message, " ")[0]
					storage.UserAddWeek(weekType, user.PeerId)

					b.Message("Выбери день недели")
					if user.Message == "Нечетная неделя" {
						b.Keyboard(getKeyboard("oddweek"))
					} else {
						b.Keyboard(getKeyboard("evenweek"))
					}
				} else {
					b.Message("Я не понимаю твоего сообщения")
				}
			} else if isDayOfWeek(user.Message) {
				schedule, err := storage.GetSchedule(user.Message, user.PeerId)
				if err != nil {
					b.Message("Я не понимаю твоего сообщения")
				}
				b.Message(schedule)
			} else {
				b.Message("Я не понимаю твоего сообщения")
			}
		}

		if _, err := bot.vk.MessagesSend(b.Params); err != nil {
			log.Printf("%s: %s", op, err)
		}
	})
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
