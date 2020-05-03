package api

import (
	"crypto/tls"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

const iconsDir = "icons"
const faviconApi = "https://iconbin.com/api/"

func (apiHandler APIHandler) HandleFavicon(w http.ResponseWriter, r *http.Request) {
	domain := mux.Vars(r)["domain"]
	icon := mux.Vars(r)["icon"]
	iconFile := iconsDir + "/" + domain + "." + icon

	_, err := os.Stat(iconFile)
	if err != nil {
		apiHandler.logger.Info("Didn't find icon, try to download.")
		// TODO make proxy configurable
		proxyUrl, err := url.Parse("http://127.0.0.1:7890")
		if err != nil {
			apiHandler.logger.Error(err)
			return
		}
		t := &http.Transport{
			Proxy:           http.ProxyURL(proxyUrl),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		// TODO rewrite with http.Get
		client := http.Client{
			Transport: t,
		}
		url := faviconApi + domain

		// TODO delete
		apiHandler.logger.Info("Try to access : " + url)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}

		res, err := client.Do(req)
		if err != nil {
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}

		body, _ := ioutil.ReadAll(res.Body)
		defer func() { _ = res.Body.Close() }()

		var foo struct {
			W             int    `json:w`
			H             int    `json:h`
			Content_type  string `json:content_type`
			Canonical_url string `json:canonical_url`
			Src           string `json:src`
		}

		err = json.Unmarshal(body, &foo)
		if err != nil {
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}

		//--------------------------------------------------------------------

		req, err = http.NewRequest(http.MethodGet, foo.Src, nil)
		if err != nil {
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}

		res, err = client.Do(req)
		if err != nil {
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}

		f, err := os.Create(iconFile)
		if err != nil {
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}
		defer func() { _ = f.Close() }()

		_, err = io.Copy(f, res.Body)

		if err != nil {
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}
	}

	f, err := os.Open(iconFile)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	// TODO delete
	apiHandler.logger.Info("Find icon: " + iconFile)

	_, err = io.Copy(w, f)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	return
}
