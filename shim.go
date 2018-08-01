package shim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Request is the request being published to the shim runner
type Request struct {
	Agent      string       `json:"agent"`
	Action     string       `json:"action"`
	RequestID  string       `json:"requestid"`
	SenderID   string       `json:"senderid"`
	CallerID   string       `json:"callerid"`
	Collective string       `json:"collective"`
	TTL        int          `json:"ttl"`
	Time       int64        `json:"msgtime"`
	Body       *RequestBody `json:"body"`
}

// RequestBody is the body passed to the
type RequestBody struct {
	Agent  string          `json:"agent"`
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
	Caller string          `json:"caller"`
}

func check(shim string, shimcfg string) error {
	if shim == "" {
		return errors.New("ruby compatability shim was not configured")
	}

	if shimcfg == "" {
		return errors.New("ruby compatability shim configuration file not configured")
	}

	if _, err := os.Stat(shim); os.IsNotExist(err) {
		return fmt.Errorf("ruby compatability shim was not found in %s", shim)
	}

	if _, err := os.Stat(shimcfg); os.IsNotExist(err) {
		return fmt.Errorf("ruby compatability shim configuration file was not found in %s", shimcfg)
	}

	return nil
}

func runShim(ctx context.Context, shim string, shimcfg string, timeout int, arg string, input string, output interface{}) error {
	err := check(shim, shimcfg)
	if err != nil {
		return err
	}

	tctx, cancel := context.WithTimeout(ctx, time.Duration(time.Duration(timeout)*time.Second))
	defer cancel()

	execution := exec.CommandContext(tctx, shim, "--config", shimcfg, arg)

	stdin, err := execution.StdinPipe()
	if err != nil {
		return fmt.Errorf("cannot create stdin for ruby compatability shim: %s", err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()

	stdout, err := execution.StdoutPipe()
	if err != nil {
		return fmt.Errorf("cannot open STDOUT for ruby compatability shim: %s", err)
	}

	err = execution.Start()
	if err != nil {
		return fmt.Errorf("cannot start the ruby compatability shim: %s", err)
	}

	err = json.NewDecoder(stdout).Decode(output)
	if err != nil {
		return fmt.Errorf("cannot decode output from the ruby compatability shim: %s", err)
	}

	go func() {
		execution.Wait()
	}()

	return nil
}

// InvokeAction runs a action found within a ruby agent
func InvokeAction(ctx context.Context, req *Request, reply interface{}, timeout int, shim string, shimcfg string) error {
	shimr, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("cannot JSON encode ruby compatability shim request: %s", err)
	}

	err = runShim(ctx, shim, shimcfg, timeout, "", string(shimr), reply)
	if err != nil {
		return err
	}

	return nil
}
