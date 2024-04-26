package web

import (
	"calendar-sync/pkg/clients"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"html/template"
	"net/http"
	"strings"
)

type AuthServer struct {
	conf       oauth2.Config
	portNumber int
	results    chan Result

	result *oauth2.Token
}

func New(portNumber int, conf oauth2.Config, results chan Result) AuthServer {
	return AuthServer{
		conf:       conf,
		portNumber: portNumber,
		results:    results,
	}
}

type Result struct {
	SourceCalendarID string
	Token            *oauth2.Token
}

type Item struct {
	Name, Value string
}

type SelectModel struct {
	Items []Item
}

const selectTemplate = `
<html>
<body>
<form method="post">
    <h1>Hello World</h1>
    <label for="calendar" >Calendar:</label>
    <select id="calendar" name="calendar">
    {{ range .Items }}
        <option value="{{ .Value }}">{{ .Name }}</option>
    {{ end }}
    </select>
    <input type="submit">
</form>
</body>
</html>
`

func (s *AuthServer) Run(ctx context.Context, cancel func(), expectedState string) {
	write := func(w http.ResponseWriter, statusCode int, body string) {
		header := w.Header()
		header.Set("Content-Type", "text/plain")
		w.WriteHeader(statusCode)

		_, err := w.Write([]byte(body))
		if err != nil {
			log.Warn().Err(err).Msg("failed to write content")
		}
	}

	writeErr := func(w http.ResponseWriter, statusCode int, err error) {
		write(w, 400, err.Error())
	}

	handleSelectCalendars := func(w http.ResponseWriter, r *http.Request) {
		if s.result == nil {
			write(w, 400, "user is not authorized")
			return
		}

		if err := r.ParseForm(); err != nil {
			writeErr(w, 500, err)
			return
		}

		calendarId := r.Form.Get("calendar")
		if calendarId != "" {
			s.results <- Result{SourceCalendarID: calendarId, Token: s.result}
			write(w, 200, "you may now close the window")
			return
		}

		client, err := clients.GetClient(ctx, s.result)
		if err != nil {
			writeErr(w, 500, err)
			return
		}

		var items []Item
		if err := client.CalendarList.List().Pages(ctx, func(list *calendar.CalendarList) error {
			for _, item := range list.Items {
				items = append(items, Item{Name: item.Summary, Value: item.Id})
			}
			return nil
		}); err != nil {
			writeErr(w, 500, err)
			return
		}

		tpl, err := template.New("webpage").Parse(selectTemplate)
		if err != nil {
			writeErr(w, 500, err)
		}

		header := w.Header()
		header.Set("Content-Type", "text/html")
		w.WriteHeader(500)

		if err = tpl.Execute(w, SelectModel{Items: items}); err != nil {
			writeErr(w, 500, err)
		}
	}

	handleAuthEnd := func(w http.ResponseWriter, r *http.Request) {
		var err error

		if err = r.ParseForm(); err != nil {
			writeErr(w, 500, err)
			return
		}

		state := r.Form.Get("state")
		if state == "" {
			write(w, 400, "state missing")
			return
		} else if state != expectedState {
			write(w, 400, "invalid state")
			return
		}

		authCode := r.Form.Get("code")
		if authCode == "" {
			write(w, 400, "auth code missing")
			return
		}

		s.result, err = s.conf.Exchange(ctx, authCode)
		if err != nil {
			writeErr(w, 500, err)
			return
		} else if s.result == nil {
			write(w, 500, "nil result")
			return
		}

		handleSelectCalendars(w, r)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToLower(r.Method) {
		case "get":
			handleAuthEnd(w, r)
		case "post":
			handleSelectCalendars(w, r)
		default:
			write(w, 405, "method not supported")
		}
	}

	listen := fmt.Sprintf(":%d", s.portNumber)

	server := http.Server{Addr: listen, Handler: http.HandlerFunc(handler)}
	if err := server.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("failed to listen and serve")
		cancel()
	}
}
