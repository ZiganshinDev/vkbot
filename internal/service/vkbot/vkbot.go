package vkbot

import (
	"log"
	"os"
	"strings"

	"github.com/SevereCloud/vksdk/api"
	"github.com/SevereCloud/vksdk/api/params"
	"github.com/SevereCloud/vksdk/object"
	"github.com/ZiganshinDev/scheduleVKBot/internal/storage"

	longpoll "github.com/SevereCloud/vksdk/longpoll-bot"
)

type VkBot struct {
	vk *api.VK
	lp *longpoll.Longpoll
}

type User struct {
	PeerId  int
	Message string
}

// New создает и запускает бота
func New(storage storage.Storage) {
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
func botHandler(bot *VkBot, storage storage.Storage) {
	bot.lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(obj.Message.PeerID)

		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)

		user := &User{PeerId: obj.Message.PeerID, Message: obj.Message.Text}

		// Обработка команд начала и возвращения
		if user.Message == "Начать" || user.Message == "Вернуться" {
			b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
		} else {
			if !storage.CheckUser(user.PeerId) {
				user.Message = strings.TrimSpace(user.Message)
				text := strings.Split(user.Message, " ")
				if len(text) != 3 || !storage.CheckSchedule(text[0], text[1], text[2]) {
					b.Message("Я не понимаю твоего сообщения")
				} else {
					storage.AddUser(text[0], text[1], text[2], user.PeerId)
					b.Message("Выбери неделю")
					b.Keyboard(getKeyboard("week"))
				}
			} else if !storage.UserCheckWeek(user.PeerId) {
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
					log.Fatal(err)
				}
				b.Message(schedule)
			} else {
				b.Message("Я не понимаю твоего сообщения")
			}
		}

		bot.vk.MessagesSend(b.Params)
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
