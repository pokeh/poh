package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

var slackBotToken string
var slackVerificationToken string

type Request struct {
	Name string `json:"name"`
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "POST" {
		err := errors.New("Method not allowed")
		res := events.APIGatewayProxyResponse{Body: "", StatusCode: 502}
		return res, err
	}

	b := []byte(req.Body)
	var reqStruct Request
	err := json.Unmarshal(b, &reqStruct)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error unmarshalling request", StatusCode: 502}, err
	}

	slackBotToken = os.Getenv("SLACK_BOT_TOKEN")
	slackVerificationToken = os.Getenv("SLACK_VERIFICATION_TOKEN")

	api := slack.New(slackBotToken)

	body := string(b)
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: slackVerificationToken}))
	if err != nil {
		log.Println("Error with ParseEvent")
		return events.APIGatewayProxyResponse{Body: "Parse error", StatusCode: 502}, err
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
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
			poh := respond(ev.Text)

			respChannel, respTimestamp, err := api.PostMessage(ev.Channel, slack.MsgOptionText(poh, false))
			if err != nil {
				msg := "Error respChannel:" + respChannel + ", respTimestamp:" + respTimestamp
				log.Println("Error at AppMentionEvent")
				return events.APIGatewayProxyResponse{Body: msg, StatusCode: 502}, err
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

	poh := "ぽぽっぽ〜"

	switch s {
	case "ping":
		poh = "ぽん"
	case "hi", "hello", "hey", "やっほー":
		poh = "やっほ〜"
	case "かわいい", "かっこいい":
		poh = "うぴゃぁ :poh:"
	case "君の名は":
		if time.Now().Unix()%2 == 0 {
			poh = "ぽー だよ"
		} else {
			poh = "ぷー だよ\n:pooh: :poh: 「「入れ替わってるーー！？！？」」"
		}
	case "しろくろまっちゃ":
		poh = "あがりコーヒーゆずさくら"
	case "天気":
		poh = "わかったらいいのにね"
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

	return poh
}
