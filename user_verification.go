package main

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func groupNameFromRequest(request telego.ChatJoinRequest) tu.MessageEntityCollection {
	groupName := tu.Entity(request.Chat.Title).Bold()
	if request.InviteLink != nil {
		groupName.TextLink(request.InviteLink.InviteLink)
	}

	return groupName
}

func (h *Handler) chatJoinRequest(bot *telego.Bot, request telego.ChatJoinRequest) {
	requestID := fmt.Sprintf("%d:%d", request.Chat.ID, request.From.ID)

	msg, err := bot.SendMessage(
		tu.MessageWithEntities(tu.ID(request.UserChatID),
			tu.Entity("Hi "), tu.Entity(request.From.FirstName).Bold(), tu.Entity(", you sent request to join "),
			groupNameFromRequest(request), tu.Entity("\n\nPlease verify the you are a real human by clicking button below"),
		).WithReplyMarkup(tu.InlineKeyboard(tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("I am real!").WithCallbackData(requestID),
		))),
	)
	if err != nil {
		bot.Logger().Errorf("Verify msg: %s", err)
		return
	}

	h.requests.SetWithTTL(requestID, Request{
		JoinRequest:           request,
		VerificationMessageID: msg.MessageID,
	}, joinRequestTTL)
}

func (h *Handler) verifyAnswer(bot *telego.Bot, query telego.CallbackQuery) {
	answer := func(text string, alert bool) {
		ans := tu.CallbackQuery(query.ID).WithText(text)
		ans.ShowAlert = alert
		err := bot.AnswerCallbackQuery(ans)
		if err != nil {
			bot.Logger().Errorf("Answer query: %s", err)
		}
	}

	removeButton := func() {
		_, err := bot.EditMessageReplyMarkup(&telego.EditMessageReplyMarkupParams{
			ChatID:      tu.ID(query.Message.Chat.ID),
			MessageID:   query.Message.MessageID,
			ReplyMarkup: nil,
		})
		if err != nil {
			bot.Logger().Errorf("Edit button: %s", err)
		}
	}

	updateText := func(text string, entities []telego.MessageEntity) {
		_, err := bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:    tu.ID(query.Message.Chat.ID),
			MessageID: query.Message.MessageID,
			Text:      text,
			Entities:  entities,
		})
		if err != nil {
			bot.Logger().Errorf("Edit text: %s", err)
		}
	}

	request, ok := h.requests.Get(query.Data)
	if !ok {
		answer("Sorry, your join request was not found!", true)
		removeButton()
		updateText("Sorry, could not find your join request, "+
			"try joining the group again if you are not a member already", nil)

		return
	}

	h.requests.Delete(query.Data)

	groupName := groupNameFromRequest(request.JoinRequest)

	err := bot.ApproveChatJoinRequest(&telego.ApproveChatJoinRequestParams{
		ChatID: tu.ID(request.JoinRequest.Chat.ID),
		UserID: request.JoinRequest.From.ID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "HIDE_REQUESTER_MISSING") {
			answer("Failed to approve!", false)
			removeButton()
			text, entities := tu.MessageEntities(
				tu.Entity("Sorry, I could not approve your join request to "), groupName,
				tu.Entity(", because too much time passed, or your join request no longer valid"))
			updateText(text, entities)

			return
		}

		bot.Logger().Errorf("Approve request: %s", err)

		answer("Failed to approve!", true)
		removeButton()
		text, entities := tu.MessageEntities(
			tu.Entity("Sorry, I could not approve your join request to "), groupName,
		)
		updateText(text, entities)

		return
	}

	answer("Verified!", false)
	removeButton()
	text, entities := tu.MessageEntities(tu.Entity("Thanks for verification!\n\nWelcome to "), groupName)
	updateText(text, entities)
}

func (h *Handler) joinRequestTTLExpired(_ string, request Request) {
	if request.JoinRequest.UserChatID == 0 {
		return
	}

	err := h.bot.DeleteMessage(&telego.DeleteMessageParams{
		ChatID:    tu.ID(request.JoinRequest.UserChatID),
		MessageID: request.VerificationMessageID,
	})
	if err != nil {
		h.bot.Logger().Errorf("Delete verification msg: %s", err)
	}

	_, err = h.bot.SendMessage(tu.MessageWithEntities(tu.ID(request.JoinRequest.UserChatID),
		tu.Entity("I didn't get verification from you in time, so your join request to "),
		groupNameFromRequest(request.JoinRequest), tu.Entity(" is rejected, please try again"),
	))
	if err != nil {
		h.bot.Logger().Errorf("Send rejected msg: %s", err)
	}

	err = h.bot.DeclineChatJoinRequest(&telego.DeclineChatJoinRequestParams{
		ChatID: tu.ID(request.JoinRequest.Chat.ID),
		UserID: request.JoinRequest.From.ID,
	})
	if err != nil {
		h.bot.Logger().Errorf("Decline request: %s", err)
	}
}

func (h *Handler) newComer(bot *telego.Bot, message telego.Message) {
	for _, user := range message.NewChatMembers {
		if user.IsBot {
			continue
		}

		requestID := fmt.Sprintf("%d:%d", message.Chat.ID, user.ID)
		request, ok := h.requests.Get(requestID)
		if !ok {
			return
		}

		h.requests.Delete(requestID)

		err := bot.DeleteMessage(&telego.DeleteMessageParams{
			ChatID:    tu.ID(request.JoinRequest.UserChatID),
			MessageID: request.VerificationMessageID,
		})
		if err != nil {
			bot.Logger().Errorf("Delete verification msg: %s", err)
		}

		_, err = bot.SendMessage(tu.MessageWithEntities(tu.ID(request.JoinRequest.UserChatID),
			tu.Entity("Your join request was approved!\n\nWelcome to "),
			groupNameFromRequest(request.JoinRequest),
		))
		if err != nil {
			bot.Logger().Errorf("Send accepted msg: %s", err)
		}
	}
}
