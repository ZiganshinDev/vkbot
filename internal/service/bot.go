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

func botHandler(vk *api.VK, lp *longpoll.Longpoll) {
	lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(obj.Message.PeerID)

		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)

		userPeerId := strconv.Itoa(obj.Message.PeerID)
		userMsg := obj.Message.Text

		if userMsg == "Начать" || userMsg == "Вернуться" {
			if database.CheckUser(userPeerId) {
				database.DeleteUser(userPeerId)
			}
			b.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
		} else if !database.CheckUser(userPeerId) {
			userMsg = strings.TrimSpace(userMsg)
			text := strings.Split(userMsg, " ")
			if len(text) != 3 || !database.CheckSchedule(text[0], text[1], text[2]) {
				b.Message("Я не понимаю твоего сообщения")
			} else {
				database.AddUser(text[0], text[1], text[2], userPeerId)
				b.Message("Выбери неделю")
				b.Keyboard(getKeyboard("week"))
			}
		} else if !database.CheckUserWithWeekType(userPeerId) {
			if userMsg == "Нечетная неделя" || userMsg == "Четная неделя" {
				weekType := strings.Split(userMsg, " ")[0]
				database.AddWeekToUser(weekType, userPeerId)
				b.Message("Выбери день недели")
				if userMsg == "Нечетная неделя" {
					b.Keyboard(getKeyboard("oddweek"))
				} else {
					b.Keyboard(getKeyboard("evenweek"))
				}
			} else {
				b.Message("Я не понимаю твоего сообщения")
			}
		} else if isDayOfWeek(userMsg) {
			b.Message(database.DBShowSchedule(userMsg, userPeerId))
		} else {
			b.Message("Я не понимаю твоего сообщения")
		}

		vk.MessagesSend(b.Params)
	})
}

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
