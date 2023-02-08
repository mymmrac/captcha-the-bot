package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mymmrac/memkey"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

const joinRequestTTL = time.Hour
const joinRequestTTLCheck = time.Minute * 5

type Handler struct {
	me       *telego.User
	bot      *telego.Bot
	bh       *th.BotHandler
	requests memkey.TypedStore[string, telego.ChatJoinRequest]
}

func NewHandler(bot *telego.Bot, bh *th.BotHandler) *Handler {
	return &Handler{
		bot: bot,
		bh:  bh,
	}
}

func (h *Handler) Init() {
	var err error
	h.me, err = h.bot.GetMe()
	assert(err == nil, "Get me:", err)

	err = h.bot.SetMyCommands(&telego.SetMyCommandsParams{
		Commands: []telego.BotCommand{
			{
				Command:     "start",
				Description: "Welcome massage",
			},
			{
				Command:     "help",
				Description: "Help info",
			},
			{
				Command:     "pending",
				Description: "Pending join requests",
			},
		},
	})
	assert(err == nil, "Set commands:", err)

	// h.bot.SetMyDefaultAdministratorRights() TODO

	h.bh.HandleMessage(h.startCmd, th.CommandEqual("start"))
	h.bh.HandleMessage(h.helpCmd, th.CommandEqual("help"))
	// h.bh.HandleMessage(nil, th.CommandEqual("pending"))

	// h.bh.HandleMessage(nil, func(update telego.Update) bool {
	// 	return len(update.Message.NewChatMembers) != 0
	// })

	// h.bh.HandleMessage(nil, func(update telego.Update) bool {
	// 	return update.Message.ChatShared != nil
	// })

	h.bh.HandleMessage(h.unknownMsg, func(update telego.Update) bool {
		return update.Message.Chat.Type == telego.ChatTypePrivate
	})

	h.bh.HandleChatJoinRequest(h.joinRequest)
	h.bh.HandleCallbackQuery(h.verifyAnswer)

	go h.requests.ExpireTTL(joinRequestTTLCheck, h.ttlExpired)
}

func (h *Handler) startCmd(bot *telego.Bot, message telego.Message) {
	var msg *telego.SendMessageParams

	chatID := tu.ID(message.Chat.ID)
	if message.Chat.Type == telego.ChatTypePrivate {
		msg = tu.MessageWithEntities(chatID,
			tu.Entity("Hi "), tu.Entity(message.From.FirstName).Bold(), tu.Entity(", I am "),
			tu.Entity(h.me.FirstName).Italic(),
			tu.Entity("!\nAdd me to the group and I will handle new comers.\n\nUse /help for more info."),
		).WithReplyMarkup(
			tu.Keyboard(tu.KeyboardRow(tu.KeyboardButton("Add me to the group").
				WithRequestChat(&telego.KeyboardButtonRequestChat{
					RequestID: int(rand.Int31()),
					UserAdministratorRights: &telego.ChatAdministratorRights{
						CanInviteUsers: true,
					},
					BotAdministratorRights: &telego.ChatAdministratorRights{
						CanInviteUsers: true,
					},
					BotIsMember: true,
				}),
			)).WithResizeKeyboard(),
		)
	} else {
		msg = tu.Message(chatID, "TODO: Non private chat msg")
	}

	_, err := bot.SendMessage(msg)
	if err != nil {
		bot.Logger().Errorf("Start cmd: %s", err)
	}
}

func (h *Handler) helpCmd(bot *telego.Bot, message telego.Message) {
	_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID),
		"Once I will be added to the group, when a new users joins, "+
			"they will be asked to verify that they are a real humans by pressing on button. "+
			"If the user doesn't click on verify button for under 1 hour, the user will be rejected from the group.",
	))
	if err != nil {
		bot.Logger().Errorf("Help cmd: %s", err)
	}
}

func (h *Handler) unknownMsg(bot *telego.Bot, message telego.Message) {
	_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID),
		"Hmm, I didn't get you, please try /start or /help.",
	))
	if err != nil {
		bot.Logger().Errorf("Unknown msg: %s", err)
	}
}

func (h *Handler) joinRequest(bot *telego.Bot, request telego.ChatJoinRequest) {
	requestID := fmt.Sprintf("%d:%d", request.Chat.ID, request.From.ID)
	h.requests.SetWithTTL(requestID, request, joinRequestTTL)

	_, err := bot.SendMessage(
		tu.Message(tu.ID(request.UserChatID),
			"TODO: Verify request",
		).WithReplyMarkup(tu.InlineKeyboard(tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("I am real!").WithCallbackData(requestID),
		))),
	)
	if err != nil {
		bot.Logger().Errorf("Verify msg: %s", err)
	}
}

func (h *Handler) verifyAnswer(bot *telego.Bot, query telego.CallbackQuery) {
	request, ok := h.requests.Get(query.Data)
	if !ok {
		err := bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithShowAlert().WithText("TODO: Not found"))
		if err != nil {
			bot.Logger().Errorf("Answer not found query: %s", err)
		}

		return
	}

	err := bot.ApproveChatJoinRequest(&telego.ApproveChatJoinRequestParams{
		ChatID: tu.ID(request.Chat.ID),
		UserID: request.From.ID,
	})
	if err != nil {
		bot.Logger().Errorf("Approve request: %s", err)

		err = bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithShowAlert().WithText("TODO: Failed to verify"))
		if err != nil {
			bot.Logger().Errorf("Answer failed to verify: %s", err)
		}

		return
	}

	h.requests.Delete(query.Data)

	err = bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("TODO: Verified"))
	if err != nil {
		bot.Logger().Errorf("Answer verified: %s", err)
	}
}

func (h *Handler) ttlExpired(_ string, request telego.ChatJoinRequest) {
	if request.UserChatID == 0 {
		return
	}

	_, err := h.bot.SendMessage(tu.Message(tu.ID(request.UserChatID),
		"TODO: Rejected",
	))
	if err != nil {
		h.bot.Logger().Errorf("Send rejected notice: %s", err)
	}

	err = h.bot.DeclineChatJoinRequest(&telego.DeclineChatJoinRequestParams{
		ChatID: tu.ID(request.Chat.ID),
		UserID: request.From.ID,
	})
	if err != nil {
		h.bot.Logger().Errorf("Decline request: %s", err)
	}
}
