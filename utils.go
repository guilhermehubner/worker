package worker

import (
	"bytes"
	"encoding/gob"
)

func encode(obj interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decode(bin []byte, obj interface{}) error {
	buf := bytes.NewBuffer(bin)
	dec := gob.NewDecoder(buf)
	return dec.Decode(obj)
}
