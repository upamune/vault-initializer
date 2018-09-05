package main

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, body io.Reader) error
}

type S3 struct {
	BucketName string
}

func NewS3(bucketName string) *S3 {
	return &S3{
		BucketName: bucketName,
	}
}

func (s *S3) getService() (*s3.S3, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return s3.New(sess), nil
}

func (s *S3) Get(key string) ([]byte, error) {
	svc, err := s.getService()
	if err != nil {
		return nil, err
	}

	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()

	return ioutil.ReadAll(obj.Body) // FIXME(upamune)
}

func (s *S3) Put(key string, body []byte) error {
	svc, err := s.getService()
	if err != nil {
		return err
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(body),
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	return err
}

// TODO
type GCS struct{}
