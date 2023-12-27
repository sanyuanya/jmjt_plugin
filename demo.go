package traefik_jmjt_plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type Config struct {
	URL        string `json:"url,omitempty"`
	AllowField string `json:"allow-field,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type Opa struct {
	next       http.Handler
	url        string
	allowField string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	return &Opa{
		next:       next,
		url:        config.URL,
		allowField: config.AllowField,
	}, nil
}

func (o *Opa) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	data := &RequestData{
		Method: r.Method,
		Path:   p,
		User:   r.Header.Get("Authorization"),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(o.url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		http.Error(w, "发生了错误", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "不等于OK", resp.StatusCode)
		return
	}

	var allowed map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&allowed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if allowed == nil {
		http.Error(w, "没有权限", http.StatusForbidden)
		return
	}
	e.next.ServeHTTP(w, r)

}

type RequestData struct {
	Method string   `json:"method"`
	Path   []string `json:"path"`
	User   string   `json:"user"`
}
