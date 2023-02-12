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

	groupName := tu.Entity(chat.Title).Bold()

	var text string
	var entities []telego.MessageEntity

	if chat.JoinByRequest {
		text, entities = tu.MessageEntities(tu.Entity("✅ Successfully added me to "), groupName,
			tu.Entity(", now I will handle all join requests sent by users"))
	} else {
		text, entities = tu.MessageEntities(tu.Entity("❌ Successfully added me to "), groupName,
			tu.Entity(", but the group should have \""), tu.Entity("Approve new members").Bold().Italic(),
			tu.Entity("\" enabled, either I will not be able to verify new users"))
	}

	_, err = bot.SendMessage(tu.Message(tu.ID(message.From.ID), text).WithEntities(entities...))
	if err != nil {
		bot.Logger().Errorf("Send chat shared msg: %s", err)
	}
}

func (h *Handler) newStatusMember(bot *telego.Bot, chatMember telego.ChatMemberUpdated) {
	groupName := tu.Entity(chatMember.Chat.Title).Bold()

	_, err := bot.SendMessage(tu.MessageWithEntities(tu.ID(chatMember.From.ID),
		tu.Entity("❌ My permissions changed in "), groupName,
		tu.Entity(" and has restricted my rights to manage new comers, "+
			"please make me an administrator with rights to "), tu.Entity("invite new users").Bold().Italic(),
		tu.Entity(", so that I can verify them")))
	if err != nil {
		bot.Logger().Errorf("Send member msg: %s", err)
	}
}

func (h *Handler) newStatusAdministrator(bot *telego.Bot, chatMember telego.ChatMemberUpdated) {
	admin, ok := chatMember.NewChatMember.(*telego.ChatMemberAdministrator)
	if !ok {
		bot.Logger().Errorf("Member not administrator: %v", chatMember)
		return
	}

	if chatMember.OldChatMember.MemberStatus() == telego.MemberStatusAdministrator {
		adminOld, okOld := chatMember.OldChatMember.(*telego.ChatMemberAdministrator)
		if okOld && adminOld.CanInviteUsers == admin.CanInviteUsers {
			return
		}
	}

	groupName := tu.Entity(chatMember.Chat.Title).Bold()

	if !admin.CanInviteUsers {
		_, err := bot.SendMessage(tu.MessageWithEntities(tu.ID(chatMember.From.ID),
			tu.Entity("❌ My permissions changed in "), groupName,
			tu.Entity(" and has restricted my rights to manage new comers, "+
				"please give me rights to "), tu.Entity("invite new users").Bold().Italic(),
			tu.Entity(", so that I can verify them")))
		if err != nil {
			bot.Logger().Errorf("Send member msg: %s", err)
		}

		return
	}

	chat, err := bot.GetChat(&telego.GetChatParams{
		ChatID: tu.ID(chatMember.Chat.ID),
	})
	if err != nil {
		bot.Logger().Errorf("Get chat: %s", err)
		return
	}

	var text string
	var entities []telego.MessageEntity

	if chat.JoinByRequest {
		text, entities = tu.MessageEntities(tu.Entity("✅ My permissions in "), groupName,
			tu.Entity(" all good, now I will handle all join requests sent by users"))
	} else {
		text, entities = tu.MessageEntities(tu.Entity("❌ My permissions in "), groupName,
			tu.Entity(" all good, but the group should have \""), tu.Entity("Approve new members").Bold().Italic(),
			tu.Entity("\" enabled, either I will not be able to verify new users"))
	}

	_, err = bot.SendMessage(tu.Message(tu.ID(chatMember.From.ID), text).WithEntities(entities...))
	if err != nil {
		bot.Logger().Errorf("Send admin status msg: %s", err)
	}
}
