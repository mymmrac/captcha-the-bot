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

	groupName := tu.Entity(chat.Title)

	var text string
	var entities []telego.MessageEntity

	if chat.JoinByRequest {
		text, entities = tu.MessageEntities(tu.Entity("Successfully added me to "), groupName,
			tu.Entity(", now I will handle all join requests sent by users"))
	} else {
		text, entities = tu.MessageEntities(tu.Entity("Successfully added me to "), groupName,
			tu.Entity(", but the group should have "), tu.Entity("Approve to join").Bold(),
			tu.Entity(" enabled, either I will not be able to verify new users!"))
	}

	_, err = bot.SendMessage(tu.Message(tu.ID(message.From.ID), text).WithEntities(entities...))
	if err != nil {
		bot.Logger().Errorf("Send chat shared msg: %s", err)
	}
}

func (h *Handler) addedMeToChatAsMember(bot *telego.Bot, chatMember telego.ChatMemberUpdated) {

}
