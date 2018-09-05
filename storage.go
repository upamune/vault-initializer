package main

import (
	"bytes"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, body []byte) error
}

type S3 struct {
	BucketName string
	session    *session.Session
}

func NewS3(sess *session.Session, bucketName string) *S3 {
	return &S3{
		BucketName: bucketName,
		session:    sess,
	}
}

func (s *S3) Get(key string) ([]byte, error) {
	svc := s3.New(s.session)

	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get an object from S3: key=%s", key)
	}
	defer obj.Body.Close()

	return ioutil.ReadAll(obj.Body) // FIXME(upamune)
}

func (s *S3) Put(key string, body []byte) error {
	svc := s3.New(s.session)

	_, err := svc.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(body),
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	return errors.Wrapf(err, "failed to put an object to S3: key=%s", key)
}

// TODO
type GCS struct{}
