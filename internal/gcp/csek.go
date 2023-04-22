package gcp

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
)

/*
Example CSEK JSON file with 2 keys from GCP's docs.
Note, This library will only create a single item CSEK bundle and only supports the 'raw' key-type.
[
  {
    "uri": "https://www.googleapis.com/compute/v1/projects/myproject/zones/us-central1-a/disks/example-disk",
    "key": "acXTX3rxrKAFTF0tYVLvydU1riRZTvUNC4g5I11NY+c=",
    "key-type": "raw"
  },
  {
    "uri": "https://www.googleapis.com/compute/v1/projects/myproject/global/snapshots/my-private-snapshot",
    "key": "ieCx/NcW06PcT7Ep1X6LUTc/hLvUDYyzSZPPVCVPTVEohpeHASqC8uw5TzyO9U+Fka9JFHz0mBibXUInrC/jEk014kCK/NPjYgEMOyssZ4ZINPKxlUh2zn1bV+MCaTICrdmuSBTWlUUiFoDD6PYznLwh8ZNdaheCeZ8ewEXgFQ8V+sDroLaN3Xs3MDTXQEMMoNUXMCZEIpg9Vtp9x2oeQ5lAbtt7bYAAHf5l+gJWw3sUfs0/Glw5fpdjT8Uggrr+RMZezGrltJEF293rvTIjWOEB3z5OHyHwQkvdrPDFcTqsLfh+8Hr8g+mf+7zVPEC8nEbqpdl3GPv3A7AwpFp7MA==",
    "key-type": "rsa-encrypted"
  }
]
*/

// TODO(joe): refactor: consider using this "official" struct instead: https://godoc.org/google.golang.org/api/compute/v1#CustomerEncryptionKey

type CSEKBundle []CSEKKey

type CSEKKey struct {
	URI     string `json:"uri" yaml:"uri"`
	Key     string `json:"key" yaml:"key"`
	KeyType string `json:"key-type" yaml:"key-type"`
}

// CreateCSEK generates a CSEKBundle for the resource specified by 'uri'.
// Only key-type 'raw' is currently supported
func CreateCSEK(uri string) (CSEKBundle, error) {
	var bundle CSEKBundle

	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return bundle, err
	}

	bundle = append(bundle, CSEKKey{
		URI:     uri,
		Key:     base64.StdEncoding.EncodeToString(key),
		KeyType: "raw",
	})
	return bundle, nil
}

func (c CSEKBundle) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func (c CSEKBundle) MarshalIndent() ([]byte, error) {
	return json.MarshalIndent(c, "", " ")
}
