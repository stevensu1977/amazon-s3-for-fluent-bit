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
	"testing"
	"time"

	fluentbit "github.com/fluent/fluent-bit-go/output"
	"github.com/stretchr/testify/assert"
)

func TestAddRecord(t *testing.T) {

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

	timestamp := time.Now()
	retCode := output.AddRecord(record, "mytag", timestamp)

	assert.Equal(t, retCode, fluentbit.FLB_OK, "Expected return code to be FLB_OK")
	assert.True(t, len(output.logs.Bytes()) > 1, "Expected output to contain 1 record")
}
