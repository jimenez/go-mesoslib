package records

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func ExampleDecoder() {
	un := Unmarshaler(func(b []byte, v interface{}) error {
		s, ok := v.([]string)
		if !ok {
			return errors.New("unexpected object store type")
		}
		if len(s) == 0 {
			return errors.New("not enough space in object store")
		}
		s[0] = string(b)
		return nil
	})
	d := NewDecoder(bytes.NewBufferString("5\nhello0\n6\nworld!"), un)
	s := []string{""}
	for {
		err := d.Decode(s)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		} else if s[0] == "" {
			fmt.Println("--empty--")
		} else {
			fmt.Println(s[0])
		}
	}
	// Output:
	// hello
	// --empty--
	// world!
}

func ExampleDecoder_json() {
	type Demo struct {
		Hello string
	}

	s := &Demo{}
	un := Unmarshaler(json.Unmarshal)
	d := NewDecoder(bytes.NewBufferString("18\n{\"hello\": \"world\"}"), un)

	err := d.Decode(s)
	if err != nil {
		if err != io.EOF {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("%+v", *s)
	}
	// Output:
	// {Hello:world}
}
