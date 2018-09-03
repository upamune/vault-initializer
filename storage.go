package main

import "io"

type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, body io.Reader) error
}

type S3 struct {

}

type GCS struct {

}
