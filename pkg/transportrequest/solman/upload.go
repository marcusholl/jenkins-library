package solman

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"reflect"
)

type fileSystem interface {
	FileExists(path string) (bool, error)
}

// SOLMANConnection Everything wee need for connecting to CTS
type SOLMANConnection struct {
	Endpoint string
	User     string
	Password string
}

// SOLMANUploadAction Collects all the properties we need for the deployment
type SOLMANUploadAction struct {
	Connection         SOLMANConnection
	ChangeDocumentId   string
	TransportRequestId string
	ApplicationID      string
	File               string
}

func (a *SOLMANUploadAction) Perform(fs fileSystem, command command.ExecRunner) error {

	notInitialized, err := ContainsEmptyStringValue(a)
	if err != nil {
		return fmt.Errorf("Cannot check everything was initialized for SOLMAN upload: %w", err)
	}
	if notInitialized {
		return fmt.Errorf("")
	}

	exists, err := fs.FileExists(a.File)
	if err != nil {
		return fmt.Errorf("Cannot upload file: %w", err)
	}
	if !exists {
		return fmt.Errorf("File '%s' does not exist.", a.File)
	}
	err = command.RunExecutable("cmclient",
		"--endpoint", a.Connection.Endpoint,
		"--user", a.Connection.User,
		"--password", a.Connection.Password,
		"--backend-type", "SOLMAN",
		"upload-file-to-transport",
		"-cID", a.ChangeDocumentId,
		"-tID", a.TransportRequestId,
		a.ApplicationID, a.File)

	return err
}

// ContainsEmptyStrings if the struct hold any empty strings
// in case the stuct contains another struct, also this struct is checked.
func ContainsEmptyStringValue(v interface{}) (bool, error) {

	if reflect.ValueOf(v).Kind() != reflect.Struct {
		return false, fmt.Errorf("%v (%T) was not a stuct", v, v)
	}
	fields := reflect.TypeOf(v)
	values := reflect.ValueOf(v)
	for i := 0; i < fields.NumField(); i++ {
		switch values.Field(i).Kind() {
		case reflect.String:
			if len(values.Field(i).String()) == 0 {
				return true, nil
			}
		case reflect.Struct:
			containsEmptyStrings, err := ContainsEmptyStringValue(values.Field(i).Interface())
			if err != nil {
				return false, err
			}
			if containsEmptyStrings {
				return true, nil
			}
		case reflect.Int:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Bool:
		default:
			return false, fmt.Errorf("Unexpected field: %v, value: %v", fields.Field(i), values.Field(i))
		}
	}

	return false, nil
}
