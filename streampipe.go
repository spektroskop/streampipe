package streampipe

import (
	"bufio"
	"net/http"
	"os/exec"

	"go.yuki.no/eventsource"
)

func Stdout(name string, arg ...string) http.Handler {
	return eventsource.Handler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			command := exec.Command(name, arg...)
			stdout, err := command.StdoutPipe()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err = command.Start(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer command.Process.Kill()

			go func() {
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					eventsource.SendMessage(w, scanner.Bytes())
					w.(http.Flusher).Flush()
				}
			}()

			<-r.Context().Done()
		}),
	)
}
