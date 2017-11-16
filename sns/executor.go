package sns

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// Execute assumes that the ARN provided is an SQS queue and sends the payload to the queue.
func Execute(arn string, payload string) (resp string, err error) {
	// Create a session object to talk to SNS (also make sure you have your key and secret setup in your .aws/credentials file)
	svc := sns.New(session.New())
	// params will be sent to the publish call included here is the bare minimum params to send a message.
	params := &sns.PublishInput{
		Message:  aws.String(payload),
		TopicArn: aws.String(arn), // e.g. arn:aws:sns:us-east-1:478989820108:MCP_DEV_CATEGORY_EX_TOPIC
	}
	po, err := svc.Publish(params)
	return po.String(), err
}
