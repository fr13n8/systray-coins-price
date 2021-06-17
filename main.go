package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/fr13n8/cprice/icon"
	"github.com/getlantern/systray"
	"github.com/robfig/cron/v3"
)

type state struct {
	Cron *cron.Cron
	Menu struct {
		USDmenu *systray.MenuItem
		GBPmenu *systray.MenuItem
		EURmenu *systray.MenuItem
	}
}

func main() {
	s := &state{}
	systray.Run(s.onReady, s.onExit)
}

func (s *state) onReady() {
	s.initMenuItems()
	s.updatePrice()

	s.Cron = cron.New()
	s.Cron.AddFunc("@every 10s", s.updatePrice)
	go s.Cron.Start()
}

func (s *state) onExit() {
	now := time.Now()
	ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
	s.Cron.Stop()
}

func (s *state) updatePrice() {
	systray.SetTemplateIcon(icon.Data, icon.Data)

	url := "https://api.coindesk.com/v1/bpi/currentprice.json"

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := httpClient.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return
	}
	data := &Response{}
	json.NewDecoder(res.Body).Decode(data)

	usdInfo := fmt.Sprintf("%s %.2f", html.UnescapeString(data.Bpi.Usd.Symbol), data.Bpi.Usd.RateFloat)
	gbpInfo := fmt.Sprintf("%s %.2f", html.UnescapeString(data.Bpi.Gbp.Symbol), data.Bpi.Gbp.RateFloat)
	eurInfo := fmt.Sprintf("%s %.2f", html.UnescapeString(data.Bpi.Eur.Symbol), data.Bpi.Eur.RateFloat)

	s.Menu.USDmenu.SetTitle(usdInfo)
	s.Menu.GBPmenu.SetTitle(gbpInfo)
	s.Menu.EURmenu.SetTitle(eurInfo)
}

func (s *state) initMenuItems() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("price")
	systray.SetTooltip("coins price")
	subMenuTop := systray.AddMenuItem("BTC", "price")
	s.Menu.USDmenu = subMenuTop.AddSubMenuItem("downladed", "downladed")
	s.Menu.GBPmenu = subMenuTop.AddSubMenuItem("downladed", "downladed")
	s.Menu.EURmenu = subMenuTop.AddSubMenuItem("downladed", "downladed")

	mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuitOrig.ClickedCh
		fmt.Println("Requesting quit")
		systray.Quit()
		fmt.Println("Finished quitting")
	}()
}
