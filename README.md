# SuperEasyMonitoring

SuperEasyMonitoring [ SEM ]

 1. Super easy monitoring
 v1:
 - hosts (hosts list/IP's in file separated by new line)
 - read hosts and ping them

 v2:
 - write status to json file

 v3:
 - create log file
 - email notification
 - sms notification

 v4:
 - switch to sqlite or mysql

 v5:
 - history in db

 v6:
 - GUI (kind of uchiwa)

 v7:
 - TCP check on port 22
 - TCP check on port 80



///
func newAccessKey(client *iam.IAM, username *string) (*iam.CreateAccessKeyOutput, error) {
	attr := &iam.CreateAccessKeyInput{UserName: username}
	resp, err := client.CreateAccessKey(attr)
	if err != nil {
		return &iam.CreateAccessKeyOutput{}, err
	}
	return resp, err
}

func deleteAccessKey(client *iam.IAM, username, accessKeyID *string) error {
	attr := &iam.DeleteAccessKeyInput{AccessKeyId: accessKeyID, UserName: username}
	_, err := client.DeleteAccessKey(attr)
	return err
}

Change code to return always a value (empty value) and error, instead of log/fmt/os.exit



Also

func (m Metric) GetID() string {
	return m.Name
}

func (m *Metric) SetID(id string) error {
	m.Name = id
	return nil
}

return nil or nothing - depends





