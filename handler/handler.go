package handler

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

func Start() {
	lambda.Start(handleRequest)
}

func handleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "POST" {
		err := errors.New("Method not allowed")
		res := events.APIGatewayProxyResponse{Body: "", StatusCode: 502}
		return res, err
	}

	token := slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: os.Getenv("SLACK_VERIFICATION_TOKEN")})
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(req.Body), token)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error at ParseEvent", StatusCode: 502}, err
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(req.Body), &r)
		if err != nil {
			return events.APIGatewayProxyResponse{Body: "Error parsing as slack event", StatusCode: 502}, err
		}

		headers := map[string]string{"content-type": "text"}
		return events.APIGatewayProxyResponse{Headers: headers, Body: r.Challenge, StatusCode: 200}, nil
	}

	return handleMessageEvent(eventsAPIEvent)
}

func handleMessageEvent(event slackevents.EventsAPIEvent) (events.APIGatewayProxyResponse, error) {
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	if event.Type == slackevents.CallbackEvent {
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// log.Println(ev.Text)
			msg := respond(ev.Text)

			respChannel, respTimestamp, err := api.PostMessage(ev.Channel, slack.MsgOptionText(msg, false))
			if err != nil {
				resp := "Error respChannel:" + respChannel + ", respTimestamp:" + respTimestamp
				return events.APIGatewayProxyResponse{Body: resp, StatusCode: 502}, err
			}
		}

		resp := "Event:" + string(innerEvent.Type) + ", received!"
		return events.APIGatewayProxyResponse{Body: resp, StatusCode: 200}, nil
	}

	return events.APIGatewayProxyResponse{Body: "nothing here", StatusCode: 200}, nil
}

func extractMessage(text string) string {
	// remove mention
	t := text[strings.Index(text, ">")+1:]
	// sanitize
	t = strings.ToLower(strings.TrimSpace(t))
	return t
}

func respond(text string) string {
	msg := "ぽぽっぽ〜"
	s := extractMessage(text)

	switch s {
	case "ping":
		msg = "ぽん"
	case "hi", "hello", "hey", "やっほー":
		msg = "やっほ〜"
	case "かわいい", "かっこいい":
		msg = "うぴゃぁ :poh:"
	case "君の名は":
		if time.Now().Unix()%2 == 0 {
			msg = "ぽー だよ"
		} else {
			msg = "ぷー だよ\n:pooh: :poh: 「「入れ替わってるーー！？！？」」"
		}
	case "しろくろまっちゃ":
		msg = "あがりコーヒーゆずさくら"
	case "天気":
		msg = "わかったらいいのにね"
	case "買い物リスト":
		msg = "忘れちゃった！"
	}

	if len(s) > 3 && strings.HasSuffix(s, "買う") {
		object := strings.TrimSuffix(s, "買う")
		msg = object + "が買いたいんだね〜覚えとくね！"
	}

	if len(s) > 1 && strings.HasSuffix(s, "た") {
		msg = strings.TrimSuffix(s, "た") + "てえらい〜！"
	}
	if len(s) > 1 && strings.HasSuffix(s, "だ") {
		msg = strings.TrimSuffix(s, "だ") + "でえらい〜！"
	}

	return msg
}
