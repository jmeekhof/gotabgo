package gotabgo

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func NewTabApi(server, version string, useTLS bool, cType ContentType) (*TabApi, error) {
	c := &httpClient{
		acceptType: cType,
	}

	return &TabApi{
		UseTLS:      useTLS,
		Server:      server,
		ApiVersion:  version,
		ContentType: cType,
		c:           c,
	}, nil

}

// Signin authenticates a user and retrieves an auth token
func (t *TabApi) Signin(username, password, contentUrl, impersonateUser string) (err error) {
	url := fmt.Sprintf("%s/api/%s/auth/signin", t.getUrl(), t.ApiVersion)
	credentials := Credentials{
		Name:     username,
		Password: password,
		Site: &Site{
			ContentUrl: contentUrl,
		},
	}

	if impersonateUser != "" {
		credentials.Impersonate = &User{
			Name: impersonateUser,
		}
	}
	signInRequest := SigninRequest{
		Request: credentials,
	}
	var payload []byte
	switch t.ContentType {
	case Xml:
		payload, err = signInRequest.XML()
	case Json:
		payload, err = json.Marshal(signInRequest)
	}
	if err != nil {
		return err
	}
	// Post this to the endpoint
	t.c.Post(url, t.ContentType.String(), bytes.NewBuffer(payload))

	return nil
}

func (t *TabApi) ServerInfo() (si *ServerInfo, err error) {
	//TODO: figure out how to use the apiversion instead of hard coding
	url := fmt.Sprintf("%s/api/%s/serverinfo", t.getUrl(), "2.4")
	r, e := t.c.Get(url)
	if e != nil {
		log.Error(e)
		return nil, e
	}

	log.WithField("method", "ServerInfo").
		Debug("response:\n", r)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	log.WithField("method", "ServerInfo").
		Debug("response:\n", string(body))
	// unmarshal this
	var sir TsResponse
	switch t.ContentType {
	case Xml:
		err = xml.Unmarshal(body, &sir)
	case Json:
		err = json.Unmarshal(body, &sir)
	}
	if err != nil {
		return
	}
	log.WithField("method", "ServerInfo").
		Debug("ServerInfoResponse:\n", sir)

	si = &sir.ServerInfo

	return

}

func (t *TabApi) getUrl() string {
	url := "http"
	if t.UseTLS {
		url += "s"
	}
	url += "://" + t.Server

	return url
}
