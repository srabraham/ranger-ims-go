package conf

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
)

var Cfg *IMSConfig

// DefaultIMS is the base configuration used for the IMS server.
// It gets overridden by values in conf/imsd.toml, then the result
// of that gets overridden by environment variables.
func DefaultIMS() *IMSConfig {
	return &IMSConfig{
		Core: ConfigCore{
			Host:          "localhost",
			Port:          80,
			JWTSecret:     rand.Text(),
			Deployment:    "dev",
			LogLevel:      "INFO",
			Directory:     DirectoryTypeClubhouseDB,
			TokenLifetime: 12 * 60 * 60, // seconds
		},
		Store: Store{
			MySQL: StoreMySQL{
				HostName: "localhost",
				HostPort: 3306,
				Database: "ims",
			},
		},
		Directory: Directory{
			TestUsers: nil,
			ClubhouseDB: ClubhouseDB{
				Hostname: "localhost",
				HostPort: 3306,
				Database: "rangers",
			},
		},
	}
}

func (c *IMSConfig) String() string {
	if c == nil {
		return "nil"
	}
	marshalled, err := json.MarshalIndent(*c, "", "  ")
	if err != nil {
		return "failed to marshal IMSConfig"
	}
	return string(marshalled)
}

type IMSConfig struct {
	Core ConfigCore
	// TODO: finish attachments feature
	//AttachmentsStore struct {
	//	S3 struct {
	//		S3AccessKeyId     string
	//		S3SecretAccessKey string
	//		S3DefaultRegion   string
	//		S3Bucket          string
	//	}
	//}
	Store     Store
	Directory Directory
}

type DirectoryType string

const (
	DirectoryTypeClubhouseDB DirectoryType = "ClubhouseDB"
	DirectoryTypeTestUsers   DirectoryType = "TestUsers"
)

func (d DirectoryType) Validate() error {
	switch d {
	case DirectoryTypeClubhouseDB:
		return nil
	case DirectoryTypeTestUsers:
		return nil
	default:
		return fmt.Errorf("unknown directory type %v", d)
	}
}

type ConfigCore struct {
	ServerRoot      string
	TokenLifetime   int64
	Dataroot        string
	Admins          []string
	DataStore       string
	Directory       DirectoryType
	ConfigRoot      string
	CachedResources string
	Host            string
	Port            int32
	MasterKey       string
	// JWTSecret won't get marshalled as part of String() due to the json "-" tag.
	JWTSecret        string `json:"-"`
	AttachmentsStore string
	Deployment       string

	// LogLevel should be one of DEBUG, INFO, WARN, or ERROR
	LogLevel string
}

type Store struct {
	MySQL StoreMySQL
}

type StoreMySQL struct {
	HostName string
	HostPort int32
	Database string
	Username string
	// Password won't get marshalled as part of String() due to the json "-" tag.
	Password string `json:"-"`
}

type TestUser struct {
	Handle      string
	Email       string
	Status      string
	DirectoryID int64
	Password    string
	Onsite      bool
	Positions   []string
	Teams       []string
}

type Directory struct {
	TestUsers   []TestUser
	ClubhouseDB ClubhouseDB
}

type ClubhouseDB struct {
	Hostname string
	HostPort int32
	Database string
	Username string
	// Password won't get marshalled as part of String() due to the json "-" tag.
	Password string `json:"-"`
}
