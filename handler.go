package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/mymmrac/memkey"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

const joinRequestTTL = time.Hour
const joinRequestTTLCheck = time.Minute * 5

type Request struct {
	JoinRequest           telego.ChatJoinRequest
	VerificationMessageID int
}

type Handler struct {
	me       *telego.User
	bot      *telego.Bot
	bh       *th.BotHandler
	requests memkey.TypedStore[string, Request]
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
		},
	})
	assert(err == nil, "Set commands:", err)

	err = h.bot.SetMyDefaultAdministratorRights(&telego.SetMyDefaultAdministratorRightsParams{
		Rights: &telego.ChatAdministratorRights{
			CanInviteUsers: true,
		},
		ForChannels: false,
	})
	assert(err == nil, "Set default administrator rights:", err)
}

func (h *Handler) RegisterHandlers() {
	// ==== USER VERIFICATION ====
	h.bh.HandleChatJoinRequest(h.chatJoinRequest)
	h.bh.HandleCallbackQuery(h.verifyAnswer)
	go h.requests.ExpireTTL(joinRequestTTLCheck, h.joinRequestTTLExpired)
	h.bh.HandleMessage(h.newComer, func(update telego.Update) bool {
		return len(update.Message.NewChatMembers) != 0
	})

	// ==== GENERAL COMMANDS ====
	h.bh.HandleMessage(h.startCmd, privateChat, th.CommandEqual("start"))
	h.bh.HandleMessage(h.helpCmd, privateChat, th.CommandEqual("help"))
	h.bh.HandleMessage(h.closeText, privateChat, th.TextEqual("Close"))

	// ==== GROUP MANAGEMENT ====
	h.bh.HandleMessage(h.chatShared, func(update telego.Update) bool {
		return update.Message.ChatShared != nil
	})
	h.bh.HandleMyChatMemberUpdated(h.newStatusMember, func(update telego.Update) bool {
		return update.MyChatMember.NewChatMember.MemberStatus() == telego.MemberStatusMember
	})
	h.bh.HandleMyChatMemberUpdated(h.newStatusAdministrator, func(update telego.Update) bool {
		return update.MyChatMember.NewChatMember.MemberStatus() == telego.MemberStatusAdministrator
	})

	// ==== FALLBACK ====
	h.bh.HandleMessage(h.unknownMsgInPrivate, privateChat)
	h.bh.HandleMessage(h.unknownMsgAnywhere, func(update telego.Update) bool {
		return strings.Contains(update.Message.Text, "@"+h.me.Username)
	})
}

func privateChat(update telego.Update) bool {
	return update.Message.Chat.Type == telego.ChatTypePrivate
}

func (h *Handler) startCmd(bot *telego.Bot, message telego.Message) {
	_, err := bot.SendMessage(tu.MessageWithEntities(tu.ID(message.Chat.ID),
		tu.Entity("Hi "), tu.Entity(message.From.FirstName).Bold(), tu.Entity(", I am "),
		tu.Entity(h.me.FirstName).Italic(),
		tu.Entity("!\nAdd me to the group with enabled "), tu.Entity("Approve new members").Bold().Italic(),
		tu.Entity(" and I will handle new comers.\n\nUse /help for more info."),
	).WithReplyMarkup(
		tu.Keyboard(
			tu.KeyboardRow(
				tu.KeyboardButton("Add me to the group").
					WithRequestChat(&telego.KeyboardButtonRequestChat{
						RequestID: rand.Int31(),
						UserAdministratorRights: &telego.ChatAdministratorRights{
							CanInviteUsers: true,
						},
						BotAdministratorRights: &telego.ChatAdministratorRights{
							CanInviteUsers: true,
						},
						BotIsMember: true,
					}),
			),
			tu.KeyboardRow(
				tu.KeyboardButton("Close"),
			),
		).WithResizeKeyboard(),
	))
	if err != nil {
		bot.Logger().Errorf("Start cmd: %s", err)
	}
}

func (h *Handler) helpCmd(bot *telego.Bot, message telego.Message) {
	_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID),
		"Once I will be added to the group, when a new users joins, "+
			"they will be asked to verify that they are a real humans by pressing on button. "+
			"If the user doesn't click on verify button for under 1 hour, the user will be rejected from the group."+
			"\n\nPleas keep in mind, that group should have request to join enabled."+
			"\n\nUse /start to add me to the group",
	))
	if err != nil {
		bot.Logger().Errorf("Help cmd: %s", err)
	}
}

func (h *Handler) closeText(bot *telego.Bot, message telego.Message) {
	_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID), "Closed").
		WithReplyMarkup(tu.ReplyKeyboardRemove()))
	if err != nil {
		bot.Logger().Errorf("Close msg: %s", err)
	}
}

func (h *Handler) unknownMsgInPrivate(bot *telego.Bot, message telego.Message) {
	_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID),
		"Hmm, I didn't get you, please try /start or /help",
	))
	if err != nil {
		bot.Logger().Errorf("Unknown msg private: %s", err)
	}
}

func (h *Handler) unknownMsgAnywhere(bot *telego.Bot, message telego.Message) {
	_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID),
		"Sorry, I can be managed only in private chat",
	))
	if err != nil {
		bot.Logger().Errorf("Unknown msg group: %s", err)
	}
}
