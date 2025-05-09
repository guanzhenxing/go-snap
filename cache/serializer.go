package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

// Serializer 定义序列化与反序列化接口
type Serializer interface {
	// Marshal 将对象序列化为字节数组
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal 将字节数组反序列化为对象
	Unmarshal(data []byte, v interface{}) error
}

// JSONSerializer 使用JSON进行序列化与反序列化
type JSONSerializer struct{}

// Marshal 将对象序列化为JSON字节数组
func (j *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal 将JSON字节数组反序列化为对象
func (j *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// GobSerializer 使用Gob进行序列化与反序列化
type GobSerializer struct{}

// Marshal 将对象序列化为Gob字节数组
func (g *GobSerializer) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, fmt.Errorf("gob serialization error: %w", err)
	}
	return buf.Bytes(), nil
}

// Unmarshal 将Gob字节数组反序列化为对象
func (g *GobSerializer) Unmarshal(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(v)
	if err != nil {
		return fmt.Errorf("gob deserialization error: %w", err)
	}
	return nil
}

// DefaultSerializer 返回默认序列化器
func DefaultSerializer() Serializer {
	return &JSONSerializer{}
}
