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
	var response struct {
		text      string
		week_type string
	}

	var schedule struct {
		institute    string
		course       string
		group_number string
	}

	lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(obj.Message.PeerID)

		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)
		log.Println(groupID)

		if database.CheckUser(strconv.Itoa(obj.Message.PeerID)) && obj.Message.Text == "Начать" {
			database.DeleteUser(strconv.Itoa(obj.Message.PeerID))
			response.text = obj.Message.Text
			b.Message("Напиши свои данные вот так \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")

		} else if !database.CheckUser(strconv.Itoa(obj.Message.PeerID)) && obj.Message.Text == "Начать" {
			response.text = obj.Message.Text
			b.Message("Напиши свои данные вот так \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")

		} else if response.text == "Начать" {
			obj.Message.Text = strings.TrimSpace(obj.Message.Text)
			text := strings.Split(obj.Message.Text, " ")
			response.text = ""

			if len(text) != 3 {
				b.Message("Я не понимаю твоего сообщения")

			} else {
				schedule.institute = text[0]
				schedule.course = text[1]
				schedule.group_number = text[2]

				if database.CheckSchedule(schedule.institute, schedule.course, schedule.group_number) {
					database.AddUser(schedule.institute, schedule.course, schedule.group_number, strconv.Itoa(obj.Message.PeerID))

					b.Message("Выбери неделю")
					b.Keyboard(getKeyboard("week"))

				} else {
					b.Message("Я не понимаю твоего сообщения")
				}
			}

		} else if database.CheckUser(strconv.Itoa(obj.Message.PeerID)) && !database.CheckUserWithWeekType(strconv.Itoa(obj.Message.PeerID)) {
			if obj.Message.Text == "Нечетная неделя" {
				response.week_type = "Нечетная"

				database.AddWeekToUser(response.week_type, strconv.Itoa(obj.Message.PeerID))
				b.Message("Выбери день недели")
				b.Keyboard(getKeyboard("oddweek"))

			} else if obj.Message.Text == "Четная неделя" {
				response.week_type = "Четная"

				database.AddWeekToUser(response.week_type, strconv.Itoa(obj.Message.PeerID))
				b.Message("Выбери день недели")
				b.Keyboard(getKeyboard("evenweek"))

			} else {
				b.Message("Я не понимаю твоего сообщения")
			}

		} else if database.CheckUserWithWeekType(strconv.Itoa(obj.Message.PeerID)) {
			if isDayOfWeek(obj.Message.Text) {
				b.Message(database.DBShowSchedule(obj.Message.Text, strconv.Itoa(obj.Message.PeerID)))

			} else {
				b.Message("Я не понимаю твоего сообщения")
			}

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
