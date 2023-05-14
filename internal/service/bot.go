package service

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/SevereCloud/vksdk/api"
	"github.com/SevereCloud/vksdk/api/params"
	"github.com/SevereCloud/vksdk/object"
	"github.com/ZiganshinDev/scheduleVKBot/internal/database"

	longpoll "github.com/SevereCloud/vksdk/longpoll-bot"
)

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
	days := map[string]struct{}{
		"Понедельник": struct{}{},
		"Вторник":     struct{}{},
		"Среда":       struct{}{},
		"Четверг":     struct{}{},
		"Пятница":     struct{}{},
	}

	lp.MessageNew(func(obj object.MessageNewObject, groupID int) {
		b := params.NewMessagesSendBuilder()
		b.RandomID(0)
		b.PeerID(obj.Message.PeerID)

		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)

		if obj.Message.Text == "Начать" || obj.Message.Text == "Вернуться" {
			b.Message("Я бот-дневник МГСУ. С какого ты института?")
			b.Keyboard(getKeyboard("institute"))

			err := vk.Execute(fmt.Sprintf(`return {text: "%v"};`, obj.Message.Text), &response)
			if err != nil {
				log.Fatal(err)
			}

		} else if response.Text == "Начать" || response.Text == "Вернуться" {
			if database.IsInstitute(obj.Message.Text) {
				b.Message("С какого ты курса?")
				b.Keyboard(getKeyboard("course"))

				err := vk.Execute(fmt.Sprintf(`return {institute: "%v", text: ""};`, obj.Message.Text), &response)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				b.Message("Такого института не существует")
			}

		} else if response.Institute != "" {
			if _, err := strconv.Atoi(string(obj.Message.Text[0])); err == nil {
				if database.IsCourse(string(obj.Message.Text[0])) {
					b.Message("Из какой ты группы?")

					query.Institute = response.Institute
					err := vk.Execute(fmt.Sprintf(`return {course: "%v", institute: ""};`, obj.Message.Text), &response)
					if err != nil {
						log.Fatal(err)
					}
				} else {
					b.Message("Такого курса не существует")
				}
			} else {
				b.Message("Такого курса не существует")
			}
		} else if response.Course != "" {
			if _, err := strconv.Atoi(obj.Message.Text); err == nil {
				if database.IsGroup(obj.Message.Text) {
					b.Message("Четная или нечетная неделя?")
					b.Keyboard(getKeyboard("week"))

					query.Course = string(response.Course[0])
					err := vk.Execute(fmt.Sprintf(`return {groupid: "%v", course: ""};`, obj.Message.Text), &response)
					if err != nil {
						log.Fatal(err)
					}
				} else {
					b.Message("Такой группы не существует")
				}
			} else {
				b.Message("Такой группы не существует")
			}

		} else if response.GroupID != "" && obj.Message.Text == "Нечетная неделя" {
			b.Message("Нечетная неделя")
			b.Keyboard(getKeyboard("oddweek"))

			query.GroupID = response.GroupID
			err := vk.Execute(fmt.Sprintf(`return {week: "%v"};`, obj.Message.Text), &response)
			if err != nil {
				log.Fatal(err)
			}

		} else if response.GroupID != "" && obj.Message.Text == "Четная неделя" {
			b.Message("Четная неделя")
			b.Keyboard(getKeyboard("evenweek"))

			query.GroupID = response.GroupID
			err := vk.Execute(fmt.Sprintf(`return {week: "%v"};`, obj.Message.Text), &response)
			if err != nil {
				log.Fatal(err)
			}

		} else if response.Week == "Нечетная неделя" && checkInMap(obj.Message.Text, days) {
			b.Message(database.DBShowSchedule(query.Institute, query.Course, query.GroupID, "Нечетная", obj.Message.Text))

		} else if response.Week == "Четная неделя" && checkInMap(obj.Message.Text, days) {
			b.Message(database.DBShowSchedule(query.Institute, query.Course, query.GroupID, "Четная", obj.Message.Text))

		} else {
			b.Message("Я не понимаю твоего сообщения")
		}

		vk.MessagesSend(b.Params)
	})
}

func checkInMap(obj string, m map[string]struct{}) bool {
	if _, inMap := m[obj]; inMap {
		return true
	}

	return false
}
