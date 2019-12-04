package serviceInfo

import "fmt"

var serviceName string
var queueName string
var nextServiceName string

func RegisterServiceName(sn string) {
	serviceName = sn
}

func RegisterNextServiceName(nsn string) {
	nextServiceName = nsn
}

func RegisterQueueName(qn string) {
	queueName = qn
}

func GetServiceName() string {
	return serviceName
}
func GetNextServiceName() string {
	return nextServiceName
}

func GetQueueName() string {
	return queueName
}

func DumpServiceInfo() string {
	return fmt.Sprintf("serviceInfo, serviceName: %q, queueName: %q, nextServiceName: %q",
		serviceName, queueName, nextServiceName)
}
