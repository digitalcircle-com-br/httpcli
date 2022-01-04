package httpcli_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/digitalcircle-com-br/httpcli"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var s http.Server

var _ = BeforeSuite(func() {

	httpcli.Cli().BasePath = "http://localhost:8090"
	mx := &http.ServeMux{}
	upgrader := websocket.Upgrader{}
	mx.HandleFunc("/a", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("X-SOME", r.Method)
		json.NewEncoder(rw).Encode(r.Method)
	})

	mx.HandleFunc("/b", func(rw http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		f, _, err := r.FormFile("file")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer f.Close()
		buf := new(bytes.Buffer)
		io.Copy(buf, f)
		rw.Header().Add("X-SOME", r.Method)
		json.NewEncoder(rw).Encode(buf.String())

	})

	mx.HandleFunc("/c", func(rw http.ResponseWriter, r *http.Request) {
		con, err := upgrader.Upgrade(rw, r, nil)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = con.WriteMessage(websocket.TextMessage, []byte("OK"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		con.Close()

	})
	s.Addr = ":8090"
	s.Handler = mx
	go s.ListenAndServe()
})
var _ = AfterSuite(func() {
	s.Close()
	os.Exit(0)
})

var _ = Describe("Lib", func() {

	It("should do a JsonGet", func() {
		out := ""
		err := httpcli.Cli().JsonGet("/a", &out)
		Expect(err).To(BeNil())
		Expect(out).To(Equal("GET"))
	})

	It("should do a JsonDelete", func() {
		out := ""
		err := httpcli.Cli().JsonDelete("/a", &out)
		Expect(err).To(BeNil())
		Expect(out).To(Equal("DELETE"))
	})

	It("should do a jsonpost", func() {
		out := ""
		err := httpcli.Cli().JsonPost("/a", &out, &out)
		Expect(err).To(BeNil())
		Expect(out).To(Equal("POST"))
	})

	It("should do a JsonPut", func() {
		out := ""
		err := httpcli.Cli().JsonPut("/a", &out, &out)
		Expect(err).To(BeNil())
		Expect(out).To(Equal("PUT"))
	})

	It("should do a RawHead", func() {
		res, err := httpcli.Cli().RawHead("/a")
		Expect(err).To(BeNil())
		Expect(res.Header.Get("X-SOME")).To(Equal("HEAD"))
	})

	It("should do a MPF", func() {
		defer func() {
			r := recover()
			Expect(r).To(BeNil())
		}()
		err := os.WriteFile("file.txt", []byte("I am a nice file"), 0600)
		Expect(err).To(BeNil())
		defer os.Remove("file.txt")
		out := ""
		err = httpcli.Cli().MultipartJson("/b",
			map[string]string{"a": "1", "b": "ASD"},
			"file",
			"./file.txt", &out)
		Expect(err).To(BeNil())
		Expect(out).To(Equal("I am a nice file"))
	})

	It("should do a WS", func() {
		defer func() {
			r := recover()
			Expect(r).To(BeNil())
		}()
		ch := make(chan bool)
		con, _, err := httpcli.Cli().WS("/c")
		if err != nil {
			Fail(err.Error())
		}
		go func() {
			_, bs, err := con.ReadMessage()
			if err != nil {
				Fail(err.Error())

			}
			Expect(string(bs)).To(Equal("OK"))
			ch <- true
		}()

		<-ch
	})

})
