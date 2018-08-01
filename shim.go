package shim

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tidwall/gjson"
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

// ValidateReply is the reply when validating a compound filter
type ValidateReply struct {
	Matched bool `json:"matched"`
}

// InvokeAction runs a action found within a ruby agent
func InvokeAction(ctx context.Context, req *Request, reply interface{}, timeout int, shim string, shimcfg string) error {
	shimr, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("cannot JSON encode ruby compatibility shim request: %s", err)
	}

	err = runDecodedShim(ctx, shim, shimcfg, timeout, "", string(shimr), reply)
	if err != nil {
		return err
	}

	return nil
}

// ParseCompoundFilter parses a compound filter and return the JSON representation of it
func ParseCompoundFilter(ctx context.Context, filter string, shim string, shimcfg string) (string, error) {
	stdout, err := runShim(ctx, shim, shimcfg, 2, "--parse-compound", filter)
	if err != nil {
		return "", err
	}

	r := gjson.GetBytes(stdout, "statusmsg")
	if r.Exists() {
		return "", fmt.Errorf("ruby compatibility shim encountered an error: %s", r.String())
	}

	return strings.TrimSpace(string(stdout)), nil
}

// ValidateCompoundCallStack validates a callstack like those produced by ParseCompoundFilter against the current node
func ValidateCompoundCallStack(ctx context.Context, cs string, timeout int, shim string, shimcfg string) (bool, error) {
	rep := &ValidateReply{}

	err := runDecodedShim(ctx, shim, shimcfg, timeout, "--validate-compound", cs, rep)
	if err != nil {
		return false, err
	}

	return rep.Matched, nil
}

// ValidateCompoundFilter validates a compound filter string against the current node
func ValidateCompoundFilter(ctx context.Context, filter string, timeout int, shim string, shimcfg string) (bool, error) {
	stack, err := ParseCompoundFilter(ctx, filter, shim, shimcfg)
	if err != nil {
		return false, fmt.Errorf("could not parse filter: %s", err)
	}

	return ValidateCompoundCallStack(ctx, stack, 10, shim, shimcfg)
}

func check(shim string, shimcfg string) error {
	if shim == "" {
		return errors.New("ruby compatibility shim was not configured")
	}

	if shimcfg == "" {
		return errors.New("ruby compatibility shim configuration file not configured")
	}

	if _, err := os.Stat(shim); os.IsNotExist(err) {
		return fmt.Errorf("ruby compatibility shim was not found in %s", shim)
	}

	if _, err := os.Stat(shimcfg); os.IsNotExist(err) {
		return fmt.Errorf("ruby compatibility shim configuration file was not found in %s", shimcfg)
	}

	return nil
}

func runDecodedShim(ctx context.Context, shim string, shimcfg string, timeout int, arg string, input string, output interface{}) error {
	o, err := runShim(ctx, shim, shimcfg, timeout, arg, input)
	if err != nil {
		return err
	}

	err = json.Unmarshal(o, output)
	if err != nil {
		return fmt.Errorf("cannot decode output from the ruby compatibility shim: %s", err)
	}

	return nil
}

func runShim(ctx context.Context, shim string, shimcfg string, timeout int, arg string, input string) ([]byte, error) {
	err := check(shim, shimcfg)
	if err != nil {
		return nil, err
	}

	tctx, cancel := context.WithTimeout(ctx, time.Duration(time.Duration(timeout)*time.Second))
	defer cancel()

	execution := exec.CommandContext(tctx, shim, "--config", shimcfg, arg)

	stdin, err := execution.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("cannot create stdin for ruby compatibility shim: %s", err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()

	stdout, err := execution.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("cannot open STDOUT for ruby compatibility shim: %s", err)
	}
	defer stdout.Close()

	err = execution.Start()
	if err != nil {
		return nil, fmt.Errorf("cannot start the ruby compatibility shim: %s", err)
	}

	buf := new(bytes.Buffer)

	n, err := buf.ReadFrom(stdout)
	if err != nil {
		return nil, fmt.Errorf("cannot read compatibility shim output: %s", err)
	}
	if n == 0 {
		return nil, fmt.Errorf("cannot read compatibility shim output: zero bytes received")
	}

	go func() {
		execution.Wait()
	}()

	return buf.Bytes(), nil
}
