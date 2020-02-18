// Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package s3 containers the OutputPlugin which sends log records to S3
package s3

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aws/amazon-kinesis-firehose-for-fluent-bit/plugins"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	fluentbit "github.com/fluent/fluent-bit-go/output"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

const (
	//Set max object size = 5 MB
	maximumObjectSize = 1024 * 1024 * 5 // 5 MiB

)

// OutputPlugin sends log records to s3
type OutputPlugin struct {
	region   string
	bucket   string
	prefix   string
	gzip     bool
	PluginID int
	client   *s3.S3
	logs     *bytes.Buffer
}

// NewOutputPlugin creates a OutputPlugin object
func NewOutputPlugin(region, bucket, prefix string, gzip bool, pluginID int) (*OutputPlugin, error) {

	sess := session.Must(session.NewSession())

	client := s3.New(sess, &aws.Config{Region: aws.String(region)})

	return &OutputPlugin{
		region:   region,
		bucket:   bucket,
		prefix:   prefix,
		gzip:     gzip,
		PluginID: pluginID,
		client:   client,
		logs:     bytes.NewBuffer([]byte{}),
	}, nil
}

// AddRecord accepts a record and adds it to the buffer, flushing the buffer if it is full
// the return value is one of: FLB_OK FLB_RETRY
// API Errors lead to an FLB_RETRY, and all other errors are logged, the record is discarded and FLB_OK is returned
func (output *OutputPlugin) AddRecord(record map[interface{}]interface{}, fluentTag string, timestamp time.Time) int {
	data, err := output.processRecord(record)
	if err != nil {
		logrus.Errorf("[s3 %d] %v\n", output.PluginID, err)
		// discard this single bad record instead and let the batch continue
		return fluentbit.FLB_OK
	}

	//check 5MB limit
	if output.logs.Len()+len(data) > maximumObjectSize {
		output.sendCurrentLogs(fluentTag)
	}

	output.logs.Write(data)
	return fluentbit.FLB_OK
}

//newUUID generate UUID func
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func (output *OutputPlugin) sendCurrentLogs(fluentTag string) error {

	var dataLenth int
	var params *s3.PutObjectInput
	var body io.ReadSeeker
	var prefix, keyname, ext string

	logrus.Debugf("Flush putobject to s3 %s\n", output.bucket)

	//generate prefix and keyname
	//etc.  /logs/2020/2/16/6,
	now := time.Now()
	year, month, day := now.Date()
	hr, min, _ := now.Clock()

	prefix = fmt.Sprintf("%s/%d/%d/%d/%d", output.prefix, year, int(month), day, hr)
	keyname = strconv.FormatInt(now.Unix(), 10)

	uuid, err := newUUID()
	if err == nil {
		if fluentTag != "" {
			keyname = fmt.Sprintf("%s-%d-%d-%d-%d-%d-%s", fluentTag, year, int(month), day, hr, min, uuid)
		} else {
			keyname = fmt.Sprintf("%d-%d-%d-%d-%d-%s", year, int(month), day, hr, min, uuid)

		}
	}

	if output.gzip {
		ext = ".gz"
	} else {
		ext = ".log"
	}

	bucket := aws.String(output.bucket)
	objectKey := aws.String(prefix + "/" + keyname + ext)

	if output.gzip {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		zw.Name = keyname + ".log"
		zw.Write(output.logs.Bytes())
		zw.Close()

		body = bytes.NewReader(buf.Bytes())
		dataLenth = buf.Len()

	} else {
		body = bytes.NewReader(output.logs.Bytes())
		dataLenth = output.logs.Len()
	}

	params = &s3.PutObjectInput{
		Bucket:        bucket,
		Key:           objectKey,
		ACL:           aws.String("bucket-owner-full-control"),
		Body:          body,
		ContentLength: aws.Int64(int64(dataLenth)),
	}

	_, err = output.client.PutObject(params)
	if err != nil {
		return err
	}

	//reset logs
	output.logs.Reset()
	return nil

}

// Flush sends the current buffer of records
func (output *OutputPlugin) Flush(fluentTag string) error {
	return output.sendCurrentLogs(fluentTag)
}

func (output *OutputPlugin) processRecord(record map[interface{}]interface{}) ([]byte, error) {

	var err error
	record, err = plugins.DecodeMap(record)
	if err != nil {
		logrus.Debugf("[s3 %d] Failed to decode record: %v\n", output.PluginID, record)
		return nil, err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	data, err := json.Marshal(record)
	if err != nil {
		logrus.Debugf("[s3 %d] Failed to marshal record: %v\n", output.PluginID, record)
		return nil, err
	}

	// append newline
	data = append(data, []byte("\n")...)

	return data, nil
}

// Takes the byte slice and returns a string
// Also removes leading and trailing whitespace
func logString(record []byte) string {
	return strings.TrimSpace(string(record))
}
