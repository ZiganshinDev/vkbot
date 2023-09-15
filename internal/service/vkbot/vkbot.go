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
	AddUser(string, string, string, int) error
	UserAddWeek(string, int) error
	DeleteUser(int) error
	GetSchedule(string, int) (string, error)
	CheckSchedule(string, string, string) (bool, error)
}

type RequestContext struct {
	User    *User
	PeerID  int
	Message string
	Builder *params.MessagesSendBuilder
	Storage Storage
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

	context := &RequestContext{
		User:    user,
		PeerID:  obj.Message.PeerID,
		Message: obj.Message.Text,
		Builder: b,
		Storage: storage,
	}

	state, exists := user.GetState(context.PeerID)
	if !exists {
		user.SetState(context.PeerID, stateStart)
	}

	log.Printf("%d: %s; %d", context.PeerID, context.Message, state)

	switch context.Message {
	case infoMessage:
		sendInfoMessage(context.Builder)
	default:
		switch state {
		case stateStart:
			v.handleStateStart(context)
		case stateRegister:
			if err := v.handleStateRegister(context); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		case stateWeekSelection:
			if err := v.handleStateWeekSelection(context); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		case stateDaySelection:
			if err := v.handleStateDaySelection(context); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	_, err := v.vk.MessagesSend(b.Params)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func sendInfoMessage(ctx *params.MessagesSendBuilder) {
	ctx.Message("Это Чат-Бот с расписанием занятий НИУ МГСУ. \nЧтобы им воспользоваться напиши свои данные согласно инструкции: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
}

func (v *VkBot) handleStateStart(ctx *RequestContext) {
	ctx.Builder.Keyboard(v.getKeyboard("info"))
	ctx.Builder.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
	ctx.User.SetState(ctx.PeerID, stateRegister)
}

func (v *VkBot) handleStateRegister(ctx *RequestContext) error {
	const op = "service.vkbot.handleStateRegister"

	if ctx.Message == backMessage {
		ctx.Builder.Keyboard(v.getKeyboard("info"))
		ctx.Builder.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
		return nil
	}

	parts, err := validateRegisterMessage(ctx.Message)
	if err != nil {
		ctx.Builder.Message("Проверь свои данные на соответствие: " + ctx.Message)
	} else if ok, err := ctx.Storage.CheckSchedule(parts[0], parts[1], parts[2]); ok && err == nil {
		if err := ctx.Storage.AddUser(parts[0], parts[1], parts[2], ctx.PeerID); err != nil {
			str := "Проверь свои данные на соответствие: " + ctx.Message
			ctx.Builder.Message(str)
			return fmt.Errorf("%s: %w", op, err)
		}

		ctx.Builder.Message("Выбери неделю")
		ctx.Builder.Keyboard(v.getKeyboard("week"))

		ctx.User.SetState(ctx.PeerID, stateWeekSelection)
	} else {
		ctx.Builder.Message("Проверь свои данные на соответствие: " + ctx.Message)
	}

	return nil
}

func (v *VkBot) handleStateWeekSelection(ctx *RequestContext) error {
	const op = "service.vkbot.handleStateWeekSelection"

	if ctx.Message == backMessage {
		ctx.User.SetState(ctx.PeerID, stateRegister)
		if err := ctx.Storage.DeleteUser(ctx.PeerID); err != nil {
			ctx.Builder.Message("Я не понимаю твоего сообщения")
			return fmt.Errorf("%s: %w", op, err)
		} else {
			ctx.Builder.Keyboard(v.getKeyboard("info"))
			ctx.Builder.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
			return nil
		}
	}

	if validateWeekMessage(ctx.Message) {
		weekType := strings.Split(ctx.Message, " ")[0]

		if err := ctx.Storage.UserAddWeek(weekType, ctx.PeerID); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		ctx.Builder.Message("Выбери день недели")

		if ctx.Message == "Нечетная неделя" {
			ctx.Builder.Keyboard(v.getKeyboard("oddweek"))
		} else {
			ctx.Builder.Keyboard(v.getKeyboard("evenweek"))
		}

		ctx.User.SetState(ctx.PeerID, stateDaySelection)
	} else {
		ctx.Builder.Message("Проверь свои данные на соответствие: " + ctx.Message)
	}

	return nil
}

func (v *VkBot) handleStateDaySelection(ctx *RequestContext) error {
	const op = "service.vk.handleStateDaySelection"

	if ctx.Message == backMessage {
		ctx.User.SetState(ctx.PeerID, stateRegister)
		if err := ctx.Storage.DeleteUser(ctx.PeerID); err != nil {
			ctx.Builder.Message("Я не понимаю твоего сообщения")
			return fmt.Errorf("%s: %w", op, err)
		} else {
			ctx.Builder.Keyboard(v.getKeyboard("info"))
			ctx.Builder.Message("Напиши свои данные вот так: \nИНСТИТУТ КУРС ГРУППА \nНапример: ИГЭС 1 37")
			return nil
		}
	}
	if validateDayMessage(ctx.Message) {
		if schedule, err := ctx.Storage.GetSchedule(ctx.Message, ctx.PeerID); err != nil {
			ctx.Builder.Message("Я не понимаю твоего сообщения")
			return fmt.Errorf("%s: %w", op, err)
		} else {
			ctx.Builder.Message(schedule)
		}
	} else {
		ctx.Builder.Message("Проверь свои данные на соответствие: " + ctx.Message)
	}

	return nil
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
