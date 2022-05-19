package results

import (
	"html/template"
	"net/http"

	"github.com/go-chi/render"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/xiaoxinpro/speedtest-go-zh/config"
	"github.com/xiaoxinpro/speedtest-go-zh/database"
	"github.com/xiaoxinpro/speedtest-go-zh/database/schema"
)

type StatsData struct {
	NoPassword bool
	LoggedIn   bool
	Data       []schema.TelemetryData
}

var (
	key   = []byte(securecookie.GenerateRandomKey(32))
	store = sessions.NewCookieStore(key)
)

func init() {
	store.Options = &sessions.Options{
		Path:     "/stats",
		MaxAge:   3600 * 1, // 1 hour
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func Stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.New("template").Parse(htmlTemplate)
	if err != nil {
		log.Errorf("Failed to parse template: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conf := config.LoadedConfig()

	if conf.DatabaseType == "none" {
		render.PlainText(w, r, "Statistics are disabled")
		return
	}

	var data StatsData

	if conf.StatsPassword == "PASSWORD" {
		data.NoPassword = true
	}

	if !data.NoPassword {
		op := r.FormValue("op")
		session, _ := store.Get(r, "logged")
		auth, ok := session.Values["authenticated"].(bool)

		if auth && ok {
			if op == "logout" {
				session.Values["authenticated"] = false
				session.Options.MaxAge = -1
				session.Save(r, w)
				http.Redirect(w, r, "/stats", http.StatusTemporaryRedirect)
			} else {
				data.LoggedIn = true

				id := r.FormValue("id")
				switch id {
				case "L100":
					stats, err := database.DB.FetchLast100()
					if err != nil {
						log.Errorf("Error fetching data from database: %s", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					data.Data = stats
				case "":
				default:
					stat, err := database.DB.FetchByUUID(id)
					if err != nil {
						log.Errorf("Error fetching data from database: %s", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					data.Data = append(data.Data, *stat)
				}
			}
		} else {
			if op == "login" {
				session, _ := store.Get(r, "logged")
				password := r.FormValue("password")
				if password == conf.StatsPassword {
					session.Values["authenticated"] = true
					session.Save(r, w)
					http.Redirect(w, r, "/stats", http.StatusTemporaryRedirect)
				} else {
					w.WriteHeader(http.StatusForbidden)
				}
			}
		}
	}

	if err := t.Execute(w, data); err != nil {
		log.Errorf("Error executing template: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
<title>网络速度测试 - 后台</title>
<style type="text/css">
	html,body{
		margin:0;
		padding:0;
		border:none;
		width:100%; min-height:100%;
	}
	html{
		background-color: hsl(198,72%,35%);
		font-family: "Segoe UI","Roboto",sans-serif;
	}
	body{
		background-color:#FFFFFF;
		box-sizing:border-box;
		width:100%;
		max-width:70em;
		margin:4em auto;
		box-shadow:0 1em 6em #00000080;
		padding:1em 1em 4em 1em;
		border-radius:0.4em;
	}
	h1,h2,h3,h4,h5,h6{
		font-weight:300;
		margin-bottom: 0.1em;
	}
	h1{
		text-align:center;
	}
	table{
		margin:2em 0;
		width:100%;
	}
	table, tr, th, td {
		border: 1px solid #AAAAAA;
	}
	th {
		width: 6em;
	}
	td {
		word-break: break-all;
	}
</style>
</head>
<body>
<h1>网络速度测试 - 后台</h1>
{{ if .NoPassword }}
		请在配置文件中修改statistics_password登录密码。
{{ else if .LoggedIn }}
	<form action="stats" method="GET"><input type="hidden" name="op" value="logout" /><input type="submit" value="退出" /></form>
	<form action="stats" method="GET">
		<h3>搜索测试结果</h6>
		<input type="hidden" name="op" value="id" />
		<input type="text" name="id" id="id" placeholder="Test ID" value=""/>
		<input type="submit" value="搜索" />
		<input type="submit" onclick="document.getElementById('id').value='L100'" value="显示最新100个测试结果" />
	</form>

	{{ range $i, $v := .Data }}
	<table>
		<tr><th>测试 ID</th><td>{{ $v.UUID }}</td></tr>
		<tr><th>时间</th><td>{{ $v.Timestamp }}</td></tr>
		<tr><th>数据</th><td>{{ $v.IPAddress }}<br/>{{ $v.ISPInfo }}</td></tr>
		<tr><th>浏览器</th><td>{{ $v.UserAgent }}<br/>{{ $v.Language }}</td></tr>
		<tr><th>下行速度</th><td>{{ $v.Download }} Mbps</td></tr>
		<tr><th>上行速度</th><td>{{ $v.Upload }} Mbps</td></tr>
		<tr><th>Ping</th><td>{{ $v.Ping }} ms</td></tr>
		<tr><th>偏差</th><td>{{ $v.Jitter }} ms</td></tr>
		<tr><th>日志</th><td>{{ $v.Log }}</td></tr>
		<tr><th>其他</th><td>{{ $v.Extra }}</td></tr>
	</table>
	{{ end }}
{{ else }}
	<form action="stats?op=login" method="POST">
		<h3>登录</h3>
		<input type="password" name="password" placeholder="请输入密码" value=""/>
		<input type="submit" value="登录" />
	</form>
{{ end }}
</body>
</html>`
