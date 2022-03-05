//httpcli - Package with a simplified http client for the most common needs.
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

//Client represents the http client itself, but wraps the original http.Client within it.
//Headers property are standard headers you want to send with every request.
//Basepath is the url prefix, in case this client will do all calls to the same endpoint.
//This is usefull when you may need to switch endpoints, like DEV/QAS/PRD, but all remainder path portion remains the same.
//
type Client struct {
	Cli            *http.Client
	BasePath       string
	Headers        http.Header
	LastResHeaders http.Header
}

//Do - This is the innermost function, and is really what DOES the http request.
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
	c.LastResHeaders = res.Header
	return res, err
}

//DoJson - calls Do, but sends i as json string in the body in case of method not being GET or HEAD or DELETE and serializes
// Body response in o, considering response will be a json object.
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

//JsonGet - Calls DoJson for a GET request.
func (c *Client) JsonGet(strurl string, o interface{}) error {
	return c.DoJson(http.MethodGet, strurl, nil, o)
}

//JsonDelete - Calls DoJson for a DELETE request.
func (c *Client) JsonDelete(strurl string, o interface{}) error {
	return c.DoJson(http.MethodDelete, strurl, nil, o)
}

//JsonPost - Calls DoJson for a POST request.
func (c *Client) JsonPost(strurl string, i interface{}, o interface{}) error {
	return c.DoJson(http.MethodPost, strurl, i, o)
}

//JsonPut - Calls DoJson for a PUT request.
func (c *Client) JsonPut(strurl string, i interface{}, o interface{}) error {
	return c.DoJson(http.MethodPut, strurl, i, o)
}

//JsonHead - Calls DoJson for a HEAD request
func (c *Client) JsonPatch(strurl string, i interface{}, o interface{}) error {
	return c.DoJson(http.MethodPatch, strurl, i, o)
}

//RawGet - Calls Do and returns body as byte slice.
func (c *Client) RawGet(strurl string) ([]byte, *http.Response, error) {
	res, err := c.Do(http.MethodGet, strurl, nil)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	bs, _ := io.ReadAll(res.Body)
	return bs, res, nil
}

//RawDelete - Calls Do and returns body as byte slice.
func (c *Client) RawDelete(strurl string) ([]byte, *http.Response, error) {
	res, err := c.Do(http.MethodDelete, strurl, nil)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	bs, _ := io.ReadAll(res.Body)
	return bs, res, nil
}

//RawPost - Calls Do and returns body as byte slice.
func (c *Client) RawPost(strurl string, i []byte) ([]byte, *http.Response, error) {
	res, err := c.Do(http.MethodPost, strurl, i)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	bs, _ := io.ReadAll(res.Body)
	return bs, res, nil
}

//RawPut - Calls Do and returns body as byte slice.
func (c *Client) RawPut(strurl string, i []byte) ([]byte, *http.Response, error) {
	res, err := c.Do(http.MethodPut, strurl, i)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	bs, _ := io.ReadAll(res.Body)
	return bs, res, nil
}

//RawHead - Calls Head
func (c *Client) RawHead(strurl string) (*http.Response, error) {
	res, err := c.Do(http.MethodHead, strurl, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//Multipart does a multipart request, eventually sending also a file.
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

	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", strurl, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err = c.Cli.Do(req)

	return
}

//MultipartJson does a multipart request, eventually sending also a file, returns body in o.
func (c *Client) MultipartJson(strurl string, params map[string]string, paramName, path string, o interface{}) (err error) {
	res, err := c.Multipart(strurl, params, paramName, path)
	if err != nil {
		return
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(o)
	return
}

//WS - creates a websocket connection. returns the conn, the http response associated or an error.
func (c *Client) WS(strurl string) (*websocket.Conn, *http.Response, error) {
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

//New - creates a new Client. Since we rely on std lib client, this is also thread safe, so no need to create multiple ones.
func New() *Client {
	ret := &Client{
		Cli:     &http.Client{},
		Headers: http.Header{},
	}
	return ret
}

//Singleton instance
var cli *Client

//In case a singleton is enought, you can use this particular instance.
func Cli() *Client {
	if cli == nil {
		cli = New()
	}
	return cli
}
