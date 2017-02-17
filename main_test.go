package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"testing"

	"github.com/src-d/babelfish-go-driver/msg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"
)

const (
	driverTestVersion = "beta-test-0.9"
)

var tests = []*myTest{
	0: newMyTest("statusError", &msg.Request{Action: msg.ParseAst},
		msg.Error, []string{"source.go:1:1: expected ';', found 'EOF'", "source.go:1:1: expected 'IDENT', found 'EOF'", "source.go:1:1: expected 'package', found 'EOF'"}),
	1: newMyTest("test1.go", loadFile("testfiles/test1.go"), msg.Ok, nil),
	2: newMyTest("test2.go", loadFile("testfiles/test2.go"), msg.Ok, nil),
	3: newMyTest("test3.go", loadFile("testfiles/test3.go"), msg.Ok, nil),
	4: newMyTest("test4.go", loadFile("testfiles/test4.go"), msg.Ok, nil),
	5: newMyTest("test5.go", loadFile("testfiles/test5.go"), msg.Ok, nil),
}

func TestGetResponse(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getResponse(test.req)
			require.Equal(t, test.res, got, fmt.Sprintf("getResponse() = %v, want %v", got, test.res))
		})
	}
}

func TestStart(t *testing.T) {
	var input []byte
	output := &bytes.Buffer{}
	var handle codec.MsgpackHandle
	dec := codec.NewDecoderBytes(output.Bytes(), &handle)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				output.Reset()
			}()

			// encode request
			sliceLen := 64 + len([]byte(test.req.Content))
			input = make([]byte, 0, sliceLen)
			enc := codec.NewEncoderBytes(&input, &handle)
			err := enc.Encode(test.req)
			require.NoError(t, err)

			// execute start()
			err = start(bytes.NewBuffer(input), output)
			require.NoError(t, err, fmt.Sprintf("start(): error = %v, want nil", err))

			// testing output can be decoded
			var got interface{}
			err = dec.Decode(got)
			if assert.Error(t, err, fmt.Sprint("An error was expected")) {
				require.Equal(t, err, io.EOF)
			}

			// encode desired response
			want := make([]byte, 0, output.Len())
			enc.ResetBytes(&want)
			err = enc.Encode(test.res)
			require.NoError(t, err)

			// Comapare output(encoded generated response) against want(encoded desired response)
			require.Equal(t, want, output.Bytes(), "start(): output != want")
		})
	}
}

func TestCmd(t *testing.T) {
	var input []byte
	output := &bytes.Buffer{}
	var handle codec.MsgpackHandle
	test := tests[4]
	t.Run(test.name, func(t *testing.T) {
		defer func() {
			output.Reset()
		}()

		// encode request
		sliceLen := 64 + len([]byte(test.req.Content))
		input = make([]byte, 0, sliceLen)
		enc := codec.NewEncoderBytes(&input, &handle)
		err := enc.Encode(test.req)
		require.NoError(t, err)

		// run command
		dv := fmt.Sprintf("-X main.driverVersion=%v", driverTestVersion)
		cmd := exec.Command("go", "run", "-ldflags", dv, "main.go", "conf_nodes.go")
		cmd.Stdin = bytes.NewBuffer(input)
		cmd.Stdout = output
		err = cmd.Run()
		require.NoError(t, err, fmt.Sprintf("exit command with errors: %v", err))

		// encode desired response
		want := make([]byte, 0, output.Len())
		enc.ResetBytes(&want)
		test.res.Driver = driverTestVersion
		err = enc.Encode(test.res)
		require.NoError(t, err)

		// Comapare output(encoded generated response) against want(encoded desired response)
		require.Equal(t, want, output.Bytes(), "start(): output != want")
	})
}
