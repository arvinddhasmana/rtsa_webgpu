// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"github.com/redpanda-data/redpanda/src/transform-sdk/go/transform"
)

func main() {
	transform.OnRecordWritten(validateAlertMessage)
}
