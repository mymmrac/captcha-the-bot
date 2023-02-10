package main

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

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
		if strings.Contains(err.Error(), "HIDE_REQUESTER_MISSING") {
			// TODO: User was rejected by admin or too long time
		}

		bot.Logger().Errorf("Approve request: %s", err)

		err = bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithShowAlert().
			WithText("TODO: Failed to verify or admin rejected"))
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

	_, err = bot.SendMessage(tu.Message(tu.ID(request.UserChatID), "TODO: Verified"))
	if err != nil {
		bot.Logger().Errorf("Send answer verified: %s", err)
	}
}

func (h *Handler) joinRequestTTLExpired(_ string, request telego.ChatJoinRequest) {
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

		_, err := bot.SendMessage(tu.Message(tu.ID(request.UserChatID),
			"TODO: Accepted",
		))
		if err != nil {
			bot.Logger().Errorf("Send accepted message: %s", err)
		}

		h.requests.Delete(requestID)
	}
}
