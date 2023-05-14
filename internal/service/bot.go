package service

import (
	"log"
	"os"
	"strconv"

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
		Text      string `json:"text"`
		Institute string `json:"institute"`
		Course    string `json:"course"`
		GroupID   string `json:"groupid"`
		Week      string `json:"week"`
	}

	var query struct {
		Text      string `json:"text"`
		Institute string `json:"institute"`
		Course    string `json:"course"`
		GroupID   string `json:"groupid"`
		Week      string `json:"week"`
	}

	lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(obj.Message.PeerID)

		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)

		switch {
		case obj.Message.Text == "Начать" || obj.Message.Text == "Вернуться":
			response.Text = obj.Message.Text
			b.Message("Я бот-дневник МГСУ. С какого ты института?")
			b.Keyboard(getKeyboard("institute"))

		case response.Text == "Начать" || response.Text == "Вернуться":
			if database.IsInstitute(obj.Message.Text) {
				response.Institute = obj.Message.Text
				query.Institute = obj.Message.Text
				response.Text = ""
				b.Message("С какого ты курса?")
				b.Keyboard(getKeyboard("course"))
			} else {
				b.Message("Такого института не существует")
			}

		case response.Institute != "":
			if _, err := strconv.Atoi(string(obj.Message.Text[0])); err == nil && database.IsCourse(string(obj.Message.Text[0])) {
				response.Course = string(obj.Message.Text[0])
				query.Course = string(obj.Message.Text[0])
				response.Institute = ""
				b.Message("Из какой ты группы?")
			} else {
				b.Message("Такого курса не существует")
			}

		case response.Course != "":
			if _, err := strconv.Atoi(obj.Message.Text); err == nil && database.IsGroup(obj.Message.Text) {
				response.GroupID = obj.Message.Text
				query.GroupID = obj.Message.Text
				response.Course = ""
				b.Message("Четная или нечетная неделя?")
				b.Keyboard(getKeyboard("week"))
			} else {
				b.Message("Такой группы не существует")
			}

		case response.GroupID != "" && (obj.Message.Text == "Нечетная неделя" || obj.Message.Text == "Четная неделя"):
			response.Week = obj.Message.Text
			response.GroupID = ""
			b.Message(response.Week)
			if response.Week == "Нечетная неделя" {
				response.Week = "Нечетная"
				b.Keyboard(getKeyboard("oddweek"))
			} else {
				response.Week = "Четная"
				b.Keyboard(getKeyboard("evenweek"))
			}

		case response.Week != "" && isDayOfWeek(obj.Message.Text):
			b.Message(database.DBShowSchedule(query.Institute, query.Course, query.GroupID, response.Week, obj.Message.Text))

		default:
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
