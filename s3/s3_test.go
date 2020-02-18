// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package s3

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/aws/amazon-kinesis-firehose-for-fluent-bit/plugins"
	fluentbit "github.com/fluent/fluent-bit-go/output"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAddRecord(t *testing.T) {
	timer, _ := plugins.NewTimeout(func(d time.Duration) {
		logrus.Errorf("[s3] timeout threshold reached: Failed to send logs for %v\n", d)
		logrus.Error("[s3] Quitting Fluent Bit")
		os.Exit(1)
	})

	output := OutputPlugin{
		region: "us-east-1",
		bucket: "flent-bit-test",
		prefix: "logs",
		gzip:   true,
		client: nil,
		logs:   bytes.NewBuffer([]byte{}),
	}

	record := map[interface{}]interface{}{
		"somekey": []byte("some value"),
	}

	timestamp = time.Now()
	retCode := output.AddRecord(record, "mytag", timestamp)

	assert.Equal(t, retCode, fluentbit.FLB_OK, "Expected return code to be FLB_OK")
	assert.Len(t, output.records, 1, "Expected output to contain 1 record")
}
