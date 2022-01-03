package httpcli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	gopath "path"
	"strings"

	"github.com/gorilla/websocket"
)

type Client struct {
	Cli      *http.Client
	Headers  http.Header
	BasePath string
}

func (c *Client) Do(method string, strurl string, body []byte) (*http.Response, error) {
	if !strings.HasPrefix(strurl, "http") {
		addr := c.BasePath
		if strings.HasSuffix(addr, "/") {
			addr = addr[:len(addr)-1]
		}
		if !strings.HasPrefix(strurl, "/") {
			strurl = "/" + strurl
		}
		strurl = addr + strurl
	}
	var req *http.Request
	var err error

	if body != nil && len(body) > 0 {
		req, err = http.NewRequest(method, strurl, io.NopCloser(bytes.NewReader(body)))

	} else {
		req, err = http.NewRequest(method, strurl, nil)
	}
	if err != nil {
		return nil, err
	}
	if c.Headers != nil {
		for k, v := range c.Headers {
			for _, v1 := range v {
				req.Header.Add(k, v1)
			}
		}
	}
	res, err := c.Cli.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, errors.New(fmt.Sprintf("Http return code - %d: %s", res.StatusCode, res.Status))
	}
	return res, err
}
func (c *Client) DoJson(method string, strurl string, i interface{}, o interface{}) (err error) {
	bs, err := json.Marshal(i)
	if err != nil {
		return err
	}
	res, err := c.Do(method, strurl, bs)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if o != nil {
		err = json.NewDecoder(res.Body).Decode(o)
	}
	return
}
func (c *Client) JsonGet(strurl string, o interface{}) error {
	return c.DoJson(http.MethodGet, strurl, nil, o)
}
func (c *Client) JsonDelete(strurl string, o interface{}) error {
	return c.DoJson(http.MethodDelete, strurl, nil, o)
}
func (c *Client) JsonHead(strurl string, o interface{}) error {
	return c.DoJson(http.MethodHead, strurl, nil, o)
}
func (c *Client) JsonPost(strurl string, i interface{}, o interface{}) error {
	return c.DoJson(http.MethodPost, strurl, i, o)
}
func (c *Client) JsonPut(strurl string, i interface{}, o interface{}) error {
	return c.DoJson(http.MethodPut, strurl, i, o)
}
func (c *Client) JsonPatch(strurl string, i interface{}, o interface{}) error {
	return c.DoJson(http.MethodPatch, strurl, i, o)
}
func (c *Client) RawGet(strurl string) ([]byte, error) {
	res, err := c.Do(http.MethodGet, strurl, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bs, _ := io.ReadAll(res.Body)
	return bs, nil
}
func (c *Client) RawDelete(strurl string) ([]byte, error) {
	res, err := c.Do(http.MethodDelete, strurl, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bs, _ := io.ReadAll(res.Body)
	return bs, nil
}
func (c *Client) RawPost(strurl string, i []byte) ([]byte, error) {
	res, err := c.Do(http.MethodPost, strurl, i)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bs, _ := io.ReadAll(res.Body)
	return bs, nil
}
func (c *Client) RawPut(strurl string, i []byte) ([]byte, error) {
	res, err := c.Do(http.MethodPut, strurl, i)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bs, _ := io.ReadAll(res.Body)
	return bs, nil
}

func (c *Client) Multipart(strurl string, params map[string]string, paramName, path string) (res *http.Response, err error) {

	if !strings.HasPrefix(strurl, "http") {
		addr := c.BasePath
		if strings.HasSuffix(addr, "/") {
			addr = addr[:len(addr)-1]
		}
		if !strings.HasPrefix(strurl, "/") {
			strurl = "/" + strurl
		}
		strurl = addr + strurl
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	req, err := http.NewRequest("POST", strurl, body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err = c.Cli.Do(req)

	if paramName != "" && path != "" {
		var file *os.File
		var part io.Writer
		file, err = os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()
		part, err = writer.CreateFormFile(paramName, gopath.Base(path))
		if err != nil {
			return
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return
		}
	}

	for key, val := range params {
		err = writer.WriteField(key, val)
		if err != nil {
			return
		}
	}
	err = writer.Close()

	return
}
func (c *Client) MultipartJson(strurl string, params map[string]string, paramName, path string, o interface{}) (err error) {
	res, err := c.Multipart(strurl, params, paramName, path)
	if err != nil {
		return
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(o)
	return
}

func (c *Client) WS(strurl string) (*websocket.Conn,*http.Response,error) {
	if !strings.HasPrefix(strurl, "http") {
		addr := c.BasePath
		if strings.HasSuffix(addr, "/") {
			addr = addr[:len(addr)-1]
		}
		if !strings.HasPrefix(strurl, "/") {
			strurl = "/" + strurl
		}
		strurl = addr + strurl
	}
	strurl = strings.Replace(strurl, "http", "ws", 1)
	return websocket.DefaultDialer.Dial(strurl, c.Headers)

}

func New() *Client {
	ret := &Client{
		Cli:     &http.Client{},
		Headers: http.Header{},
	}
	return ret
}

var cli *Client

func Cli() *Client {
	if cli == nil {
		cli = New()
	}
	return cli
}
