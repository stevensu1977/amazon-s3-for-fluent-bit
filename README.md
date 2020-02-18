## Fluent Bit Plugin for Amazon S3 
A Fluent Bit output plugin for Amazon S3. 

### Usage

Run `make` to build `./bin/s3.so`. Then use with Fluent Bit:
```
./fluent-bit -e ./s3.so -i cpu \
-o s3 \
-p "region=us-east-1" \
-p "bucket=fluent-bit-s3-bucket"
```

You can build your own image with Dockerfile.


### Plugin Options

* `region`: The region which your bucket is in,.
* `bucket`: The name of the bucket that you want log records sent to.
* `prefix`: Specify a custom prefix, default is "logs" .
*  `gzip`: By default, the whole log record should be compressed. If you set up  this option with "false", then all logs not be compressed .

### Permissions

The plugin requires s3 bucket write permissions.

### Credentials

This plugin uses the AWS SDK Go, and uses its [default credential provider chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html). If you are using the plugin on Amazon EC2 or Amazon ECS or Amazon EKS, the plugin will use your EC2 instance role or [ECS Task role permissions](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html) or [EKS IAM Roles for Service Accounts for pods](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html). The plugin can also retrieve credentials from a [shared credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html), or from the standard `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN` environment variables.

### Environment Variables

* `FLB_LOG_LEVEL`: Set the log level for the plugin. Valid values are: `debug`, `info`, and `error` (case insensitive). Default is `info`. **Note**: Setting log level in the Fluent Bit Configuration file using the Service key will not affect the plugin log level (because the plugin is external).



### Example Fluent Bit Config File

This is example for nginx access log.
```
[SERVICE]
    Flush    5
[INPUT]
    Name          tail
    Tag  my_nginx
    Path          /var/log/nginx/*.log
[OUTPUT]
    Name  s3
    Match *
    region us-east-1
    bucket fluent-bit-s3
    prefix logs    
    gzip  true
```

### AWS for Fluent Bit

AWS official distribute a container image with Fluent Bit and these plugins.


* Image: [github.com/aws/aws-for-fluent-bit](https://github.com/aws/aws-for-fluent-bit)

* Plugin:
  * [amazon-kinesis-firehose-for-fluent-bit](https://github.com/aws/amazon-kinesis-firehose-for-fluent-bit)
  * [amazon-cloudwatch-logs-for-fluent-bit](https://github.com/aws/amazon-cloudwatch-logs-for-fluent-bit)
  * [amazon-kinesis-streams-for-fluent-bit](https://github.com/aws/amazon-kinesis-streams-for-fluent-bit)
* Docker hub: [amazon/aws-for-fluent-bit](https://hub.docker.com/r/amazon/aws-for-fluent-bit/tags)

### Official output plugin

* [fuentbit.io](https://docs.fluentbit.io/manual/output)


## License

This library is licensed under the Apache 2.0 License.