package httpclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/kosha/vmware-aria/pkg/logger"
	"github.com/kosha/vmware-aria/pkg/models"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func makeAPITokenCall(apiKey string, method, serverUrl string, log logger.Logger) ([]byte, int) {

	var req *http.Request
	data := url.Values{}
	data.Set("refresh_token", apiKey)
	encodedData := data.Encode()
	req, _ = http.NewRequest(method, serverUrl, strings.NewReader(encodedData))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Accept-Encoding", "identity")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Error(err)
		return nil, 500
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}
	return bodyBytes, resp.StatusCode
}

func setOauth2Header(newReq *http.Request, tokenMap map[string]string) {
	newReq.Header.Set("Authorization", "Bearer "+tokenMap["access_token"])
	newReq.Header.Set("Content-Type", "application/json")

	newReq.Header.Set("Accept-Encoding", "identity")

	return
}

func Oauth2ApiRequest(headers map[string]string, method, url string, data interface{}, tokenMap map[string]string, log logger.Logger) ([]byte, int) {
	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	var body io.Reader
	if data == nil {
		body = nil
	} else {
		var requestBody []byte
		requestBody, err := json.Marshal(data)
		if err != nil {
			log.Error(err)
			return nil, 500
		}
		body = bytes.NewBuffer(requestBody)
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Error(err)
		return nil, 500
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	setOauth2Header(request, tokenMap)
	response, err := client.Do(request)

	if err != nil {
		log.Error(err)
		return nil, 500
	}
	defer response.Body.Close()
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return nil, 500
	}
	return respBody, response.StatusCode
}

func MakeHttpCall(headers map[string]string, method, url string, body interface{}, token string, log logger.Logger) (interface{}, int, error) {

	var response interface{}
	var payloadRes []byte

	var statusCode int
	tokenMap := make(map[string]string)

	if token != "" {
		tokenMap["access_token"] = token
		payloadRes, statusCode = Oauth2ApiRequest(headers, method, url, body, tokenMap, log)
		if string(payloadRes) == "" {
			return nil, statusCode, fmt.Errorf("nil")
		}
		// Convert response body to target struct
		err := json.Unmarshal(payloadRes, &response)
		if err != nil {
			log.Error("Unable to parse response as json")
			log.Error(err)
			return nil, http.StatusInternalServerError, err
		}
		if statusCode == 200 && response != nil {
			return response, statusCode, nil
		}
	}
	return nil, http.StatusInternalServerError, fmt.Errorf("token invalid")
}

func GenerateToken(apiKey, serverUrl string, log logger.Logger) (string, int, error) {
	// token is not generated, or is invalid so get new token
	token, expiresIn, _ := getToken(apiKey, serverUrl, log)
	if token == "" {
		return "", 0, fmt.Errorf("error generating api token")
	}
	return token, expiresIn, nil
}

func getToken(apiKey, serverUrl string, log logger.Logger) (string, int, int) {

	var tokenResponse models.APIToken

	url := serverUrl
	res, _ := makeAPITokenCall(apiKey, "POST", url, log)
	if string(res) == "" {
		return "", 0, 500
	}
	// Convert response body to target struct
	err := json.Unmarshal(res, &tokenResponse)
	if err != nil {
		log.Error("Unable to parse auth token response as json")
		log.Error(err)
		return "", 0, 500
	}
	return tokenResponse.AccessToken, tokenResponse.ExpiresIn, 200
}
