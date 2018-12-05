package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
)

const (
	region               = "ap-northeast-1"
	default_duration_min = 60 * 24
)

var (
	duration_min int64
)

func printInstanceName(svc *ec2.EC2, instanceId *string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{instanceId},
	}
	resp, err := svc.DescribeInstances(input)
	if err != nil {
		return "", err
	}
	tags := resp.Reservations[0].Instances[0].Tags
	for _, elem := range tags {
		if aws.StringValue(elem.Key) == "Name" {
			return aws.StringValue(elem.Value), nil
		}
	}
	return "", nil
}

func main() {
	flag.Int64Var(&duration_min, "d", default_duration_min, "The number of minutes worth of events to retrieve.")
	flag.Parse()

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatal(err)
	}

	ec2Svc := ec2.New(sess)
	elasticacheSvc := elasticache.New(sess)

	ec2Params := &ec2.DescribeInstanceStatusInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("event.code"),
				Values: []*string{
					aws.String("instance-reboot"),
					aws.String("instance-stop"),
					aws.String("instance-retirement"),
					aws.String("system-reboot"),
					aws.String("system-maintenance"),
				},
			},
		},
	}
	elasticacheParams := &elasticache.DescribeEventsInput{
		Duration: &duration_min,
	}

	ec2Resp, err := ec2Svc.DescribeInstanceStatus(ec2Params)
	if err != nil {
		log.Fatal(err)
	}

	for _, instance := range ec2Resp.InstanceStatuses {
		name, err := printInstanceName(ec2Svc, instance.InstanceId)
		if err != nil {
			log.Fatal(err)
		}
		for _, event := range instance.Events {
			fmt.Printf("%v", *instance.InstanceId)
			if name != "" {
				fmt.Printf(" (%s)", name)
			}
			fmt.Printf(": %v %v\n", *event.Code, *event.Description)
		}
	}

	elasticacheResp, err := elasticacheSvc.DescribeEvents(elasticacheParams)
	if err != nil {
		log.Fatal(err)
	}
	for _, event := range elasticacheResp.Events {
		fmt.Printf("%v (%v): %v - %v\n", *event.SourceIdentifier, *event.SourceType, *event.Message, *event.Date)
	}
}
