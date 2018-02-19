package multichain

import (
	"fmt"
	"time"
	"errors"
	"strconv"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"encoding/base64"
	//
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
	//
	"github.com/dghubble/sling"
	//
	"github.com/golangdaddy/tarantula/web"
)

const (
	CONST_ID = "multichain-client"
)

type Response map[string]interface{}

func (r Response) Result() interface{} {
	return r["result"]
}

type Client struct {
	req web.RequestInterface
	httpClient *http.Client
	chain string
	host string
	port int
	credentials string
	debug bool
}

func NewClient(req web.RequestInterface, chain, username, password string, port int) *Client {

	credentials := username + ":" + password

	return &Client{
		req: req,
		httpClient: &http.Client{},
		chain: chain,
		port: port,
		credentials: base64.StdEncoding.EncodeToString([]byte(credentials)),
	}
}

func (client *Client) Localhost() *Client {
	c := *client
	c.host = fmt.Sprintf(
		"http://%s:%v",
		"localhost",
		client.port,
	)
	return &c
}

func (client *Client) ViaNode(ipv4 string, port int) *Client {
	c := *client
	c.host = fmt.Sprintf(
		"http://%s:%v",
		ipv4,
		port,
	)
	return &c
}

func (client *Client) IsDebugMode() bool {
	return client.debug
}

func (client *Client) DebugMode() *Client {
	client.debug = true
	return client
}

func (client *Client) Urlfetch(ctx context.Context, seconds ...int) {

	if len(seconds) > 0 {
		ctx, _ = context.WithDeadline(
			ctx,
			time.Now().Add(time.Duration(1000000000 * seconds[0]) * time.Second),
		)
	}

	client.httpClient = urlfetch.Client(ctx)
}

func (client *Client) msg(params []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "1.0",
		"id": CONST_ID,
		"params": params,
	}
}

func (client *Client) Command(method string, params []interface{}) map[string]interface{} {

	msg := client.msg(params)
	msg["method"] = fmt.Sprintf("%s", method)

	if client.IsDebugMode() {
		client.req.Log().DebugJSON(msg)
	}

	return msg
}

func (client *Client) Post(msg interface{}) (Response, error) {

	t := time.Now()

	if client.debug {
		client.req.Log().Debug("DEBUG MODE ON...")
		b, _ := json.Marshal(msg)
		client.req.Log().Debug(string(b))
	}

	request, err := sling.New().Post(client.host).BodyJSON(msg).Request()
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Basic " + client.credentials)

	resp, err := client.httpClient.Do(request)
	if err != nil {

		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if client.IsDebugMode() {
		client.req.Log().Debug(string(b))
	}

	obj := make(Response)

	err = json.Unmarshal(b, &obj)
	if err != nil {
		return nil, err
	}

	if obj["error"] != nil {
		e := obj["error"].(map[string]interface{})
		var s string
		m, ok := msg.(map[string]interface{})
		if ok {
			s = fmt.Sprintf("multichaind - '%s': %s", m["method"], e["message"].(string))
		} else {
			s = fmt.Sprintf("multichaind - %s", e["message"].(string))
		}
		return nil, errors.New(s)
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("INVALID RESPONSE STATUS CODE: "+strconv.Itoa(resp.StatusCode))
	}

	if client.debug {
		client.req.Log().Debug(
			fmt.Sprintf("RPC CMD DURATION: %s", time.Since(t).String()),
		)
	}

	return obj, nil
}
