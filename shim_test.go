package shim

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shim")
}

type output struct {
	Statuscode      int         `json:"statuscode"`
	Statusmsg       string      `json:"statusmsg"`
	Data            interface{} `json:"data"`
	DisableResponse bool        `json:"-"`
}

var _ = Describe("Shim", func() {
	var (
		ctx    context.Context
		cancel func()
		rep    *output
		shimr  *Request
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		shimr = &Request{
			Agent:      "one",
			Action:     "status",
			RequestID:  "",
			SenderID:   "",
			CallerID:   "",
			Collective: "",
			TTL:        0,
			Time:       -62135596800,
			Body: &RequestBody{
				Agent:  "one",
				Action: "status",
				Data:   json.RawMessage(nil),
				Caller: "",
			},
		}
		rep = &output{}
	})

	AfterEach(func() {
		cancel()
	})

	Describe("check", func() {
		It("Should fail when no shim is configured", func() {
			err := check("", "")
			Expect(err).To(MatchError("ruby compatability shim was not configured"))
		})

		It("Should fail when the shim cannot be found", func() {
			err := check("/nonexisting", "testdata/shim.cfg")
			Expect(err).To(MatchError("ruby compatability shim was not found in /nonexisting"))
		})

		It("Should fail without a shim config file", func() {
			err := check("testdata/nonzero_shim.sh", "")
			Expect(err).To(MatchError("ruby compatability shim configuration file not configured"))
		})

		It("Should fail when a shim config file does not exist", func() {
			err := check("testdata/nonzero_shim.sh", "/nonexisting")
			Expect(err).To(MatchError("ruby compatability shim configuration file was not found in /nonexisting"))
		})
	})

	Describe("InvokeAction", func() {
		It("Should unmarshal the result", func() {
			shim := filepath.Join("testdata", "good_shim.sh")
			shimcfg := filepath.Join("testdata", "shim.cfg")

			if runtime.GOOS == "windows" {
				shim = filepath.Join("testdata", "good_shim_windows.bat")
			}

			err := InvokeAction(ctx, shimr, rep, 5, shim, shimcfg)
			Expect(err).To(BeNil())

			Expect(rep.Statusmsg).To(Equal("OK"))
			Expect(rep.Statuscode).To(Equal(0))

			d := rep.Data.(map[string]interface{})
			Expect(d["test"].(string)).To(Equal("ok"))
		})
	})
})
