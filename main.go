package main

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

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "POST" {
		err := errors.New("Method not allowed")
		res := events.APIGatewayProxyResponse{Body: "", StatusCode: 502}
		return res, err
	}

	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

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

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
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

func respond(text string) string {
	// remove mention <@xxx>
	t := strings.Split(strings.ToLower(text), " ")
	s := t[1]
	for _, r := range t[2:] {
		s += " " + r
	}

	msg := "ぽぽっぽ〜"

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
	}

	runes := []rune(s)
	if len(runes) > 1 {
		orig := string(runes[:len(runes)-1])

		switch string(runes[len(runes)-1]) {
		case "た":
			return orig + "てえらい〜！"
		case "だ":
			return orig + "でえらい〜！"
		}
	}

	return msg
}
