package main

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) chatShared(bot *telego.Bot, message telego.Message) {
	chat, err := bot.GetChat(&telego.GetChatParams{
		ChatID: tu.ID(message.ChatShared.ChatID),
	})
	if err != nil {
		bot.Logger().Errorf("Get chat: %s", err)
		return
	}

	var msg string
	if chat.JoinByRequest {
		msg = "TODO: All good"
	} else {
		msg = "TODO: Need join by request"
	}

	_, err = bot.SendMessage(tu.Message(tu.ID(message.From.ID), msg))
	if err != nil {
		bot.Logger().Errorf("Send need approves: %s", err)
	}
}

func (h *Handler) addedMeToChatAsMember(bot *telego.Bot, chatMember telego.ChatMemberUpdated) {

}
