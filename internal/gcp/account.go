package gcp

func GetCurrentAccount() (string, error) {
	out, err := output("gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
	if err != nil {
		return "", err
	}
	return string(out), nil
}
