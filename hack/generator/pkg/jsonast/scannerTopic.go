package jsonast

import "strings"

// ScannerTopic captures the topic of what structure we are currently generating
type ScannerTopic struct {
	objectName    string
	objectVersion string
	propertyName  string
}

// NewObjectScannerTopic creates a new topic for scanning the subtree of an object
func NewObjectScannerTopic(objectName string, objectVersion string) ScannerTopic {
	return ScannerTopic{
		objectName:    objectName,
		objectVersion: objectVersion,
		propertyName:  "",
	}
}

// WithProperty creates a new topic for scanning the subtree of a property
func (topic ScannerTopic) WithProperty(propertyName string) *ScannerTopic {
	topic.propertyName = propertyName
	return &topic
}

// CreateStructName returns the name to use for a new struct in this topic
func (topic ScannerTopic) CreateStructName() string {
	return strings.Title(topic.objectName) + strings.Title(topic.propertyName)
}
