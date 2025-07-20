package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/paginator"
)

var baseURL = "https://ctftime.org/api/v1/"
var teamId = 54706

type EventDuration struct {
	Hours int
	Days  int
}

type Events struct {
	Title    string
	EventUrl string `json:"ctftime_url"`
	Id       int    `json:"ctf_id"`
	Url      string
	Duration EventDuration
	Format   string
	Start    time.Time
	Weight   float64
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)
	if nil != err {
		// panics for the sake of simplicity.
		// you should handle this error properly in your code.
		panic(err)
	}

	log.Println("Who watches the watchmen, dawg?")

	b.RegisterHandler(bot.HandlerTypeMessageText, "upcoming", bot.MatchTypeCommandStartOnly, upcomingHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "now", bot.MatchTypeCommandStartOnly, nowHandler)

	b.Start(ctx)
}

func upcomingHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Print("/upcoming")

	eventsFmt := "events/?limit=20&start=%d&finish=%d"

	now := time.Now()
	oneMonthAfter := now.AddDate(0, 3, 0)

	r, err := http.Get(baseURL + fmt.Sprintf(eventsFmt, now.Unix(), oneMonthAfter.Unix()))

	if err != nil {
		return // TODO: report error
	}

	defer r.Body.Close()

	var events []Events
	json.NewDecoder(r.Body).Decode(&events)

	var ctfs []string
	for _, event := range events {
		var msg strings.Builder
		msg.WriteString(fmt.Sprintf("[%s](%s) \\([%d](%s)\\)\n", bot.EscapeMarkdown(event.Title), bot.EscapeMarkdown(event.Url), event.Id, bot.EscapeMarkdown(event.EventUrl)))

		msg.WriteString(bot.EscapeMarkdown(event.Format))
		msg.WriteRune('\n')

		msg.WriteString(bot.EscapeMarkdown(event.Start.Local().Format("Mon Jan 2 15:04 MST")))
		msg.WriteRune('\n')

		msg.WriteString("Duration: ")
		msg.WriteString(bot.EscapeMarkdown(event.Duration.String()))
		msg.WriteRune('\n')

		msg.WriteString(bot.EscapeMarkdown(fmt.Sprintf("Weight: %.2f", event.Weight)))

		ctfs = append(ctfs, msg.String())
	}

	opts := []paginator.Option{
		paginator.PerPage(4),
		paginator.WithoutEmptyButtons(),
	}

	p := paginator.New(b, ctfs, opts...)

	_, err = p.Show(ctx, b, fmt.Sprintf("%d", update.Message.Chat.ID))

	if err != nil {
		log.Println(err.Error())
	}
}

func (d EventDuration) String() string {
	if d.Days == 0 && d.Hours == 0 {
		return "No time"
	}

	if d.Days == 0 {
		s := fmt.Sprintf("%d hour", d.Hours)
		if d.Hours > 1 {
			s += "s"
		}
		return s
	}

	if d.Hours == 0 {
		s := fmt.Sprintf("%d day", d.Days)
		if d.Days > 1 {
			s += "s"
		}
		return s
	}

	plural := ""

	if d.Days > 1 {
		plural = "s"
	}

	return fmt.Sprintf("%d day%s and %d hours", d.Days, plural, d.Hours)
}

func nowHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Now é complexo, n tá implementado ainda n chapa.",
	})
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "Next week CTFs: `/upcoming`\nCTFs happening now: `/now`",
		ParseMode: models.ParseModeMarkdown,
	})
}
