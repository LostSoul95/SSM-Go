package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func main() {

	// Create an aws session.

	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("us-east-1")},
		SharedConfigState: session.SharedConfigEnable,
	})

	// Error check
	
	if err != nil {
		panic(err)
	}

	// Work around targets struct to make it work.

	key := "InstanceIds"
	values := []string{""} // Pass on the intended instance ID in here.

	// AWS provides this func for getting a pointer to a string
	// which gets consumed by the command struct

	value := aws.StringSlice(values)

	// Create a slice of pointers of type target.
	targets := []*ssm.Target{}
	target1 := new(ssm.Target)
	target1.Key = &key
	target1.Values = value
	targets = append(targets, target1)

	// Create a new client of ssm through the aws session.

	ssmsvc := ssm.New(sess, aws.NewConfig())

	// Create an SSM document to attach to the instance.

	contentForDocument := `{
		"schemaVersion":"2.0",
		"description":"Run a PowerShell script to get the IP address of the target instance.",
		"mainSteps":[
		   {
			  "action":"aws:runPowerShellScript",
			  "name":"runPowerShellWithSecureString",
			  "inputs":{
				 "runCommand":[
					"Get-NetIPConfiguration"
				 ]
			  }
		   }
		]
	 }`

	documentType := "Command"
	name := "RunPowershellForNet"
	targetType := "/AWS::EC2::Instance"
	associationName := "RunPowershellTargets"
	documentFormat := "JSON"

	_, err = ssmsvc.CreateDocument(&ssm.CreateDocumentInput{
		Content:        &contentForDocument,
		DocumentType:   &documentType,
		Name:           &name,
		TargetType:     &targetType,
		DocumentFormat: &documentFormat,
	})

	if err != nil {
		fmt.Println("Error while creating the document.")
		fmt.Println(err)
	}

	// Create an association between the target instances and the document.

	out, err := ssmsvc.CreateAssociation(&ssm.CreateAssociationInput{
		Name:            &name,
		Targets:         targets,
		AssociationName: &associationName,
	})

	if err != nil {

		fmt.Println("Error while creating the association.")
		panic(err)
	}

	fmt.Println(out)

	// Run through Send command.

	output, err := ssmsvc.SendCommand(&ssm.SendCommandInput{
		DocumentName: &name,
		Targets:      targets,
	})

	if err != nil {
		fmt.Println("Error while running the command on the target isntance.")
		panic(err)
	}

	fmt.Println(output)

}
