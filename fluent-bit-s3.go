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

package main

import (
	"C"
	"unsafe"

	"github.com/aws/amazon-kinesis-firehose-for-fluent-bit/plugins"
	"github.com/fluent/fluent-bit-go/output"
	"github.com/sirupsen/logrus"
	s3plugin "github.com/stevensu1977/amazon-s3-for-fluent-bit/s3"
)
import (
	"errors"
	"strings"
	"time"
)

var (
	pluginInstances []*s3plugin.OutputPlugin
)

func addPluginInstance(ctx unsafe.Pointer) error {
	pluginID := len(pluginInstances)
	output.FLBPluginSetContext(ctx, pluginID)
	instance, err := newS3Output(ctx, pluginID)
	if err != nil {
		return err
	}

	pluginInstances = append(pluginInstances, instance)
	return nil
}

func getPluginInstance(ctx unsafe.Pointer) *s3plugin.OutputPlugin {
	pluginID := output.FLBPluginGetContext(ctx).(int)
	return pluginInstances[pluginID]
}

// The "export" comments have syntactic meaning
// This is how the compiler knows a function should be callable from the C code

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "s3", "Amazon S3 Fluent Bit Plugin.")
}

func newS3Output(ctx unsafe.Pointer, pluginID int) (*s3plugin.OutputPlugin, error) {
	region := output.FLBPluginConfigKey(ctx, "region")
	logrus.Infof("[s3 %d] plugin parameter region = '%s'\n", pluginID, region)
	bucket := output.FLBPluginConfigKey(ctx, "bucket")
	logrus.Infof("[s3 %d] plugin parameter bucket = '%s'\n", pluginID, bucket)
	if region == "" || bucket == "" {
		return nil, errors.New("region , bucket required!")
	}

	prefix := output.FLBPluginConfigKey(ctx, "prefix")
	logrus.Infof("[s3 %d] plugin parameter prefix = '%s'\n", pluginID, prefix)
	gzip := getBoolParam(ctx, "gzip", true)
	logrus.Infof("[s3 %d] plugin parameter gzip = '%t'\n", pluginID, gzip)
	return s3plugin.NewOutputPlugin(region, bucket, prefix, gzip, pluginID)
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	plugins.SetupLogger()

	err := addPluginInstance(ctx)
	if err != nil {
		logrus.Errorf("[s3] Failed to initialize plugin: %v\n", err)
		return output.FLB_ERROR
	}
	return output.FLB_OK
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	var count int
	var ret int
	var ts interface{}
	var record map[interface{}]interface{}

	// Create Fluent Bit decoder
	dec := output.NewDecoder(data, int(length))

	s3Output := getPluginInstance(ctx)
	fluentTag := C.GoString(tag)
	logrus.Debugf("[s3 %d] Found logs with tag: %s\n", s3Output.PluginID, fluentTag)

	for {
		// Extract Record
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}

		var timestamp time.Time
		switch tts := ts.(type) {
		case output.FLBTime:
			timestamp = tts.Time
		case uint64:
			// when ts is of type uint64 it appears to
			// be the amount of seconds since unix epoch.
			timestamp = time.Unix(int64(tts), 0)
		default:
			timestamp = time.Now()
		}

		retCode := s3Output.AddRecord(record, fluentTag, timestamp)
		if retCode != output.FLB_OK {
			return retCode
		}
		count++
	}
	err := s3Output.Flush(fluentTag)
	if err != nil {
		logrus.Errorf("[s3 %d] %v\n", s3Output.PluginID, err)
		return output.FLB_ERROR
	}
	logrus.Debugf("[s3 %d] Processed %d events with tag %s\n", s3Output.PluginID, count, fluentTag)

	return output.FLB_OK
}

func getBoolParam(ctx unsafe.Pointer, param string, defaultVal bool) bool {
	val := strings.ToLower(output.FLBPluginConfigKey(ctx, param))
	if val == "true" {
		return true
	} else if val == "false" {
		return false
	} else {
		return defaultVal
	}
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
